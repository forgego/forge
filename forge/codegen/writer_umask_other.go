//go:build !unix

package generator

func processUmask() int {
	return 0o022
}
