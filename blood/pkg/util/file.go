package util

import (
	"errors"
	"os"
	"path"
	"strings"
)

func CreateDirNotExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		e := os.MkdirAll(dir, os.ModePerm)
		if e != nil {
			return errors.New("Error creating directory: " + e.Error())
		}
	}
	return nil
}
func CreateFileNotExists(filePath string, content ...[]byte) error {
	dir := path.Dir(filePath)
	err := CreateDirNotExists(dir)
	if err != nil {
		return err
	}
	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		f, e := os.Create(filePath)
		if e != nil {
			return e
		}
		if len(content) > 0 {
			for _, c := range content {
				_, err = f.Write(c)
				if err != nil {
					return err
				}
			}
		}
		defer f.Close()
	}
	return nil
}

func GetFileDir(filePath string) string {
	if strings.HasSuffix(filePath, "/") {
		filePath, _ = strings.CutSuffix(filePath, "/")
	}
	return filePath[strings.LastIndex(filePath, "/")+1:]
}

func GetFilesInDir(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var fileList []string
	for _, file := range files {
		if !file.IsDir() {
			fileList = append(fileList, path.Join(dir, file.Name()))
		}
	}
	return fileList, nil
}
