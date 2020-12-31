package cli

var binaryExtensions = []string{
	"", ".o", ".out", ".bin",
}

func isExtBinary(ext string) bool {
	for _, e := range binaryExtensions {
		if ext == e {
			return true
		}
	}
	return false
}
