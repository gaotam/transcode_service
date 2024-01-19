package utils

import "os"

func CreateFolder(path string) (err error) {
	return os.MkdirAll(path, os.ModePerm)
}
