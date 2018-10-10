package util

//DuplicateBytes creates a new slice the same size as src, and copies src into it before returning it
func DuplicateBytes(src []byte) []byte {
	dst := make([]byte, len(src), len(src))
	copy(dst, src)
	return dst
}
