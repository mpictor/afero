// +build fuzz

package afero

// for fuzzing instructions, see afero_fuzz.go; you will need the name of one of
// the below funcs.

func FuzzLinkWrapNul(buf []byte) int {
	fs := &SymlinkingWrapper{}
	return fuzz(buf, fs)
}

func FuzzLinkWrapMem(buf []byte) int {
	fs := &SymlinkingWrapper{
		Underlying: NewMemMapFs(),
	}
	return fuzz(buf, fs)
}
