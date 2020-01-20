// +build fuzz

package afero

// to generate (only needed for fuzzing):
// go generate -tags fuzz .

//go:generate protoc -I. --go_out=. afero_fuzz.proto
//go:generate sed -i  "1i// +build fuzz\\n\\n// Generated with protoc + sed - see link_wrap_fuzz.go\\n" afero_fuzz.pb.go

// To fuzz: run the following commands, replacing FuzzFn with the name of the
// fuzz function you wish to use.
//
// go-fuzz-build -tags fuzz -func FuzzFn github.com/spf13/afero
// go-fuzz

import (
	"os"
	"time"

	"github.com/golang/protobuf/proto"
)

type Sym3Fs interface {
	Fs
	Lstater
	Symlinker
	Readlinker
}

// return values:
// -1 avoid
//  0 meh
//  1 interesting
func fuzz(buf []byte, fs Sym3Fs) int {
	pb := &DataSet{}
	err := proto.Unmarshal(buf, pb)
	if err != nil {
		return -1
	}
	for _, act := range pb.Actions {
		var s bool //used with *IfPossible funcs
		iface := act.GetFsInterface()
		switch x := iface.(type) {
		case *Action_MsgCreate:
			if x == nil || x.MsgCreate == nil {
				return 0
			}
			_, err = fs.Create(x.MsgCreate.Name)
		case *Action_MsgMkdir:
			if x == nil || x.MsgMkdir == nil || x.MsgMkdir.Perm == nil {
				return 0
			}
			err = fs.Mkdir(x.MsgMkdir.Name, os.FileMode(x.MsgMkdir.Perm.Mode))
		case *Action_MsgMkdirAll:
			if x == nil || x.MsgMkdirAll == nil || x.MsgMkdirAll.Perm == nil {
				return 0
			}
			err = fs.MkdirAll(x.MsgMkdirAll.Path, os.FileMode(x.MsgMkdirAll.Perm.Mode))
		case *Action_MsgOpen:
			if x == nil || x.MsgOpen == nil {
				return 0
			}
			_, err = fs.Open(x.MsgOpen.Name)
		case *Action_MsgOpenFile:
			if x == nil || x.MsgOpenFile == nil || x.MsgOpenFile.Perm == nil {
				return 0
			}
			_, err = fs.OpenFile(x.MsgOpenFile.Name, int(x.MsgOpenFile.Flag), os.FileMode(x.MsgOpenFile.Perm.Mode))
		case *Action_MsgRemove:
			if x == nil || x.MsgRemove == nil {
				return 0
			}
			err = fs.Remove(x.MsgRemove.Name)
		case *Action_MsgRemoveAll:
			if x == nil || x.MsgRemoveAll == nil {
				return 0
			}
			err = fs.RemoveAll(x.MsgRemoveAll.Path)
		case *Action_MsgRename:
			if x == nil || x.MsgRename == nil {
				return 0
			}
			err = fs.Rename(x.MsgRename.Oldname, x.MsgRename.Newname)
		case *Action_MsgStat:
			if x == nil || x.MsgStat == nil {
				return 0
			}
			_, err = fs.Stat(x.MsgStat.Name)
		case *Action_MsgName:
			if x == nil || x.MsgName == nil {
				return 0
			}
			fs.Name()
		case *Action_MsgChmod:
			if x == nil || x.MsgChmod == nil || x.MsgChmod.Mode == nil {
				return 0
			}
			err = fs.Chmod(x.MsgChmod.Name, os.FileMode(x.MsgChmod.Mode.Mode))
		case *Action_MsgChtimes:
			if x == nil || x.MsgChtimes == nil {
				return 0
			}
			err = fs.Chtimes(x.MsgChtimes.Name, t(x.MsgChtimes.Atime), t(x.MsgChtimes.Mtime))
		case *Action_MsgLstat:
			if x == nil || x.MsgLstat == nil {
				return 0
			}
			_, s, err = fs.LstatIfPossible(x.MsgLstat.Name)
			if !s {
				return 1
			}
		case *Action_MsgSymlink:
			if x == nil || x.MsgSymlink == nil {
				return 0
			}
			s, err = fs.SymlinkIfPossible(x.MsgSymlink.Oldname, x.MsgSymlink.Newname)
			if !s {
				return 1
			}
		case *Action_MsgReadlink:
			if x == nil || x.MsgReadlink == nil {
				return 0
			}
			var str string
			str, s, err = fs.ReadlinkIfPossible(x.MsgReadlink.Name)
			if str == "" {
				//not sure if possible
				return 1
			}
			if !s {
				return 1
			}
		case *Action_MsgEvalSymlinks:
			if x == nil || x.MsgEvalSymlinks == nil {
				return 0
			}
			_, s, err = EvalSymlinks(fs, x.MsgEvalSymlinks.Path)
			if !s {
				return 1
			}
		default:
			return 0
		}
		if err != nil {
			return 0
		}
	}
	return 1
}

func t(tm *Time) time.Time {
	if tm == nil {
		return time.Time{}
	}
	return time.Unix(0, tm.TS)
}
