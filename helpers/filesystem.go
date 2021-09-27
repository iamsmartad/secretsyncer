package helpers

import (
	"os"
	"strings"
)

// OSReadDir ...
func OSReadDir(root string) ([]string, error) {
	var files []string
	f, err := os.Open(root)
	if err != nil {
		return files, err
	}
	fileInfo, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return files, err
	}
	for _, file := range fileInfo {
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}
		if file.IsDir() {
			continue
		}
		files = append(files, file.Name())
	}
	return files, nil
}
