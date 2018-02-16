package installer

import (
	"io"
	"os"
)

func Copy(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}

func Exists(path string) bool {
	file, err := os.Open(path)
	defer file.Close()

	if os.IsNotExist(err) {
		return false
	}

	return true
}
