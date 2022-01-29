package infras

import (
	"crypto/sha256"
	"fmt"
	"github.com/mholt/archiver"
	"io"
	"os"
	"path/filepath"
)

func CheckSum(target string) (string, error) {
	file, err := os.Open(target)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func Decompress(target, dir string) error {
	if len(dir) == 0 {
		dir = filepath.Dir(target)
	}
	return archiver.Unarchive(target, dir)
}
