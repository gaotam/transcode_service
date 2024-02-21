package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func CreateFolder(path string) (err error) {
	return os.MkdirAll(path, os.ModePerm)
}

func GetFileName(filePath string) string {
	return strings.Split(filepath.Base(filePath), ".")[0]
}
