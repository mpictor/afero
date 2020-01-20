package afero

import (
	"testing"
)

//func (w *SymlinkingWrapper) translate(name string) (string, error)
func TestSWTranslate(t *testing.T) {
	m := NewMemMapFs()
	w := &SymlinkingWrapper{Underlying: m}
	err := w.Mkdir("/a", 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = w.MkdirAll("/c/d", 0755)
	if err != nil {
		t.Fatal(err)
	}
	success, err := w.SymlinkIfPossible("/c/d", "/a/b")
	if success != true || err != nil {
		t.Fatal("Out of Cheese Error. Redo From Start.", success, err)
	}
	success, err = w.SymlinkIfPossible("../f", "/c/d/e")
	if success != true || err != nil {
		t.Fatal("Out of Cheese Error. Redo From Tart.", success, err)
	}
	for _, td := range []struct {
		name, in, want string
	}{
		{
			name: "complete",
			in:   "/a/b",
			want: "/c/d",
		},
		{
			name: "last",
			in:   "/a/b/c",
			want: "/c/d/c",
		},
		{
			name: "identity",
			in:   "/c/d",
			want: "/c/d",
		},
		{
			name: "relative",
			in:   "/c/d/e",
			want: "/c/f",
		},
	} {
		t.Run(td.name, func(t *testing.T) {
			got := w.translate(td.in)
			if got != td.want {
				t.Errorf("\nwant: %s\n got: %s", td.want, got)
			}
		})
	}
}

//func (w *SymlinkingWrapper) SymlinkIfPossible(oldname, newname string) (bool, error)
func TestSWSymlinkIfPossible(t *testing.T) {
	m := NewMemMapFs()
	w := &SymlinkingWrapper{Underlying: m}
	success, err := w.SymlinkIfPossible("/c/d", "/a/b")
	if success != true {
		t.Fatal("returned false")
	}
	if err == nil {
		t.Error("should not succeed - /a does not exist so is not dir")
	}
	//create a file. error?
	f, err := w.Create("/g")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	success, err = w.SymlinkIfPossible("/g/h", "/e/f")
	if success != true {
		t.Fatal("returned false")
	}
	if err == nil {
		t.Error("should not succeed - /g is not dir")
	}
	err = w.Mkdir("/i", 0644)
	if err != nil {
		t.Error(err)
	}
	success, err = w.SymlinkIfPossible("/k/l", "/i/j")
	if success != true || err != nil {
		t.Error("should succeed - /i is  dir", success, err)
	}
}
