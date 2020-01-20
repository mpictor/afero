// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/afero"
)

func main() {
	i := flag.String("i", "", "input protobuf, pretty-print json")
	o := flag.String("o", "", "output sample protobuf")
	s := flag.Bool("s", true, "use same name for all sample actions (-o)")
	d := flag.String("d", "", "do ops described in protobuf")
	flag.Parse()
	if len(*i) > 0 {
		buf := pb2json(*i)
		fmt.Printf("%s\n", buf)
		return
	}
	if len(*o) > 0 {
		writepb(*o, *s)
		return
	}
	if len(*d) > 0 {
		readpb(*d)
		return
	}
	flag.Usage()
	os.Exit(1)
}

func pb2json(i string) []byte {
	data, err := ioutil.ReadFile(i)
	if err != nil {
		panic(err)
	}
	pb := &afero.DataSet{}
	proto.Unmarshal(data, pb)
	buf, err := json.MarshalIndent(pb, "", "  ")
	if err != nil {
		panic(err)
	}
	return buf
}

func writepb(o string, s bool) {
	pb := &afero.DataSet{}
	pb.Actions = []*afero.Action{
		&afero.Action{
			FsInterface: &afero.Action_MsgCreate{
				MsgCreate: &afero.MsgCreate{
					Name: name("MsgCreateName", s),
				},
			},
		},
		&afero.Action{
			FsInterface: &afero.Action_MsgMkdir{
				MsgMkdir: &afero.MsgMkdir{
					Name: name("MsgMkdirName", s),
					Perm: &afero.FileMode{Mode: 0644},
				},
			},
		},
		&afero.Action{
			FsInterface: &afero.Action_MsgMkdirAll{
				MsgMkdirAll: &afero.MsgMkdirAll{
					Path: name("MsgMkdirAllPath", s),
					Perm: &afero.FileMode{Mode: 0644},
				},
			},
		},
		&afero.Action{
			FsInterface: &afero.Action_MsgOpen{
				MsgOpen: &afero.MsgOpen{
					Name: name("MsgOpenName", s),
				},
			},
		},
		&afero.Action{
			FsInterface: &afero.Action_MsgOpenFile{
				MsgOpenFile: &afero.MsgOpenFile{
					Name: name("MsgOpenFileName", s),
					Flag: 12345,
					Perm: &afero.FileMode{Mode: 0644},
				},
			},
		},
		&afero.Action{
			FsInterface: &afero.Action_MsgRemove{
				MsgRemove: &afero.MsgRemove{
					Name: name("MsgRemoveName", s),
				},
			},
		},
		&afero.Action{
			FsInterface: &afero.Action_MsgRemoveAll{
				MsgRemoveAll: &afero.MsgRemoveAll{
					Path: name("MsgRemoveAllPath", s),
				},
			},
		},
		&afero.Action{
			FsInterface: &afero.Action_MsgRename{
				MsgRename: &afero.MsgRename{
					Oldname: name("MsgRenameOldname", s),
					Newname: name("MsgRenameNewname", s),
				},
			},
		},
		&afero.Action{
			FsInterface: &afero.Action_MsgStat{
				MsgStat: &afero.MsgStat{
					Name: name("MsgStatName", s),
				},
			},
		},
		&afero.Action{
			FsInterface: &afero.Action_MsgName{
				MsgName: &afero.MsgName{},
			},
		},
		&afero.Action{
			FsInterface: &afero.Action_MsgChmod{
				MsgChmod: &afero.MsgChmod{
					Name: name("MsgChmodName", s),
					Mode: &afero.FileMode{Mode: 0644},
				},
			},
		},
		&afero.Action{
			FsInterface: &afero.Action_MsgChtimes{
				MsgChtimes: &afero.MsgChtimes{
					Name:  name("MsgChtimesName", s),
					Atime: &afero.Time{TS: 1234},
					Mtime: &afero.Time{TS: 4567},
				},
			},
		},
		&afero.Action{
			FsInterface: &afero.Action_MsgLstat{
				MsgLstat: &afero.MsgLstat{
					Name: name("MsgLstatName", s),
				},
			},
		},
		&afero.Action{
			FsInterface: &afero.Action_MsgSymlink{
				MsgSymlink: &afero.MsgSymlink{
					Oldname: name("MsgSymlinkOldname", s),
					Newname: name("MsgSymlinkNewname", s),
				},
			},
		},
		&afero.Action{
			FsInterface: &afero.Action_MsgReadlink{
				MsgReadlink: &afero.MsgReadlink{
					Name: name("MsgReadlinkName", s),
				},
			},
		},
		&afero.Action{
			FsInterface: &afero.Action_MsgEvalSymlinks{
				MsgEvalSymlinks: &afero.MsgEvalSymlinks{
					Path: name("MsgEvalSymlinksPath", s),
				},
			},
		},
	}
	buf, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(o, buf, 0644)
	if err != nil {
		panic(err)
	}
}

func name(n string, s bool) string {
	if s {
		return "filename"
	}
	return n
}

func readpb(r string) {
	data, err := ioutil.ReadFile(r)
	if err != nil {
		panic(err)
	}
	defer func() {
		rec := recover()
		if rec != nil {
			fmt.Printf("recover %s\n", rec)
			fmt.Println(string(pb2json(r)))
			panic(rec)
		}
	}()
	rc := afero.FuzzLinkWrapMem(data)
	fmt.Printf("rc=%d\n", rc)
}

/*
grep message afero/afero_fuzz.proto |cut -d ' ' -f2|grep Msg|\
sed 's/\(Msg\)\(.*\)/a=afero.Action_\1\2{\1\2:&afero.\1\2{Name:"\1\2Name"}}\npb.Actions=append(pb.Actions,&afero.Action{FsInterface:a})/'

grep message afero/afero_fuzz.proto |cut -d ' ' -f2|grep Msg|\
sed 's/\(Msg\)\(.*\)/pb.Actions=append(pb.Actions,\&afero.Action{\nFsInterface:\&afero.Action_\1\2{\n\1\2:\&afero.\1\2{Name:"\1\2Name"},\n},\n})/'

*/
