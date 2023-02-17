package busybox

import (
	"io"
	"os"
)

func Read(filePath string) (string, error) {
	fileReader, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fileReader.Close()
	fd, err := io.ReadAll(fileReader)
	return string(fd), nil
}
