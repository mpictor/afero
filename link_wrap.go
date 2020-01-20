package afero

import (
	"fmt"
	"os"
	fp "path/filepath"
	"strings"
	"sync"
	"time"
)

var ErrUnderlyingNil = fmt.Errorf("SymlinkingWrapper.Underlying is nil")

//SymlinkingWrapper wraps another fs and provides in-memory symlinks
type SymlinkingWrapper struct {
	Underlying Fs
	mu         sync.RWMutex
	symlinks   map[string]*syminfo
}

//must impl interface Fs
var _ Fs = (*SymlinkingWrapper)(nil)

//funcs for Fs interface
func (w *SymlinkingWrapper) Create(name string) (File, error) {
	if w.Underlying != nil {
		return w.Underlying.Create(w.translate(name))
	}
	return nil, ErrUnderlyingNil
}
func (w *SymlinkingWrapper) Mkdir(name string, perm os.FileMode) error {
	if w.Underlying != nil {
		return w.Underlying.Mkdir(w.translate(name), perm)
	}
	return ErrUnderlyingNil
}
func (w *SymlinkingWrapper) MkdirAll(path string, perm os.FileMode) error {
	if w.Underlying != nil {
		return w.Underlying.MkdirAll(w.translate(path), perm)
	}
	return ErrUnderlyingNil
}
func (w *SymlinkingWrapper) Open(name string) (File, error) {
	if w.Underlying != nil {
		return w.Underlying.Open(w.translate(name))
	}
	return nil, ErrUnderlyingNil
}
func (w *SymlinkingWrapper) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	if w.Underlying != nil {
		return w.Underlying.OpenFile(w.translate(name), flag, perm)
	}
	return nil, ErrUnderlyingNil
}
func (w *SymlinkingWrapper) Remove(name string) error {
	tr := w.translate(name)
	if w.isLink(tr) {
		w.rmLink(tr)
		return nil
	}
	if w.Underlying != nil {
		return w.Underlying.Remove(tr)
	}
	return ErrUnderlyingNil
}

//if path is a symlink, remove the symlink AND target
func (w *SymlinkingWrapper) RemoveAll(path string) error {
	tr := w.translate(path)
	if w.isLink(tr) {
		w.rmLink(tr)
	}
	if w.Underlying != nil {
		return w.Underlying.RemoveAll(tr)
	}
	return ErrUnderlyingNil
}
func (w *SymlinkingWrapper) Rename(oldname, newname string) error {
	oldname = w.translate(oldname)
	newname = w.translate(newname)
	if w.Underlying != nil {
		return w.Underlying.Rename(oldname, newname)
	}
	return ErrUnderlyingNil
}
func (w *SymlinkingWrapper) Stat(name string) (os.FileInfo, error) {
	tr := w.translate(name)
	if w.isLink(tr) {
		fi := w.symlinks[tr]
		return fi, nil
	}
	if w.Underlying != nil {
		return w.Underlying.Stat(tr)
	}
	return nil, ErrUnderlyingNil
}
func (w *SymlinkingWrapper) Name() string {
	if w.Underlying != nil {
		return "memlink+" + w.Underlying.Name()
	}
	return ""
}
func (w *SymlinkingWrapper) Chmod(name string, mode os.FileMode) error {
	if w.Underlying != nil {
		return w.Underlying.Chmod(w.translate(name), mode)
	}
	return ErrUnderlyingNil
}
func (w *SymlinkingWrapper) Chtimes(name string, atime time.Time, mtime time.Time) error {
	if w.Underlying != nil {
		return w.Underlying.Chtimes(w.translate(name), atime, mtime)
	}
	return ErrUnderlyingNil
}

//translate symlink to real path
//can only be called when w.mu is locked or rlocked
func (w *SymlinkingWrapper) translate(name string) string {
	name = strings.TrimRight(name, "/")
	if w.symlinks == nil || len(w.symlinks) == 0 {
		//no symlinks exist
		return name
	}
	translated, ok := w.symlinks[name]
	if ok {
		//happy path - this exact item exists as a symlink
		tgt := translated.target
		if strings.Contains(tgt, "..") {
			abstgt, err := resolveRelative(fp.Dir(name), tgt)
			if err == nil {
				tgt = abstgt
			}
		}
		return tgt
	}
	//look up each component and translate
	//split on slashes, start on left
	orig := name
	remain := name
	var j int
	for {
		i := strings.Index(remain, "/")
		if i == -1 {
			break
		}
		if i >= len(remain) {
			break
		}
		remain = remain[i+1:]
		j += i
		if j > 0 {
			//ignore leading slash
			segment := name[:j]
			translated, ok := w.symlinks[segment]
			if ok {
				//update loop vars and keep going
				//FIXME target can be relative - need to resolve that
				name = fp.Join(translated.target, name[j:])
				remain = name
				j = 0
			}
		}
		j += 1
	}
	if strings.Contains(name, "..") {
		absname, err := resolveRelative(fp.Base(orig), name)
		if err == nil {
			name = absname
		}
	}
	return name
}

//returns true if given name is a link
//can only be called when w.mu is locked or rlocked
func (w *SymlinkingWrapper) isLink(name string) bool {
	if w.symlinks == nil || len(w.symlinks) == 0 {
		//no symlinks exist
		return false
	}
	_, ok := w.symlinks[name]
	return ok
}

//delete given link
//can only be called when w.mu is locked or rlocked
func (w *SymlinkingWrapper) rmLink(name string) {
	delete(w.symlinks, name)
}

//Lstater
func (w *SymlinkingWrapper) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if w.symlinks != nil {
		//resolve any symlinks in name
		name := w.translate(name)
		if si, ok := w.symlinks[name]; ok {
			return si, true, nil
		}
	}
	return nil, true, ErrFileNotFound
}

//Symlinker
func (w *SymlinkingWrapper) SymlinkIfPossible(oldname, newname string) (bool, error) {
	//need Underlying later on
	if w.Underlying == nil {
		return true, ErrUnderlyingNil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.symlinks == nil {
		w.symlinks = make(map[string]*syminfo)
	}
	fmt.Println(oldname, newname)
	newname = w.translate(newname)
	fmt.Println(newname)
	//make sure newname's dir exists and is a dir; if not, creating a real symlink would fail.
	fi, err := w.Underlying.Stat(fp.Dir(newname))
	if err != nil {
		return true, err
	}
	if !fi.IsDir() {
		return true, &os.LinkError{"symlink", oldname, newname, os.ErrNotExist}
	}
	//the dir exists. does the actual link?
	_, exists := w.symlinks[newname]
	if exists {
		return true, ErrFileExists
	}
	//...and what about a regular file by that name?
	_, err = w.Underlying.Stat(newname)
	if err == nil {
		return true, ErrFileExists
	}
	w.symlinks[newname] = NewSymInfo(oldname, fp.Base(newname))
	return true, nil
}

//Readlinker
func (w *SymlinkingWrapper) ReadlinkIfPossible(name string) (string, bool, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if w.symlinks != nil {
		//resolve any symlinks in name
		name = w.translate(name)
		_, ok := w.symlinks[name]
		if ok {
			return name, true, nil
		}
	}
	return "", true, ErrFileNotFound
}

//FileInfo for symlinks
type syminfo struct {
	name   string
	target string //as with real symlinks, this can be relative
	mtime  time.Time
}

var _ os.FileInfo = (*syminfo)(nil)

func (*syminfo) Mode() os.FileMode    { return os.ModeSymlink | os.ModePerm }
func (*syminfo) IsDir() bool          { return false }
func (s *syminfo) Size() int64        { return int64(len(s.target)) }
func (s *syminfo) ModTime() time.Time { return s.mtime }
func (s *syminfo) Name() string       { return s.name }
func (*syminfo) Sys() interface{}     { return nil }
func NewSymInfo(target, name string) *syminfo {
	return &syminfo{
		mtime:  time.Now(),
		name:   name,
		target: target,
	}
}
