package common

import (
	"io/fs"
	"path/filepath"
)

func ListFilesRecursive(directory string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			//make path relative to directory
			rel, err := filepath.Rel(directory, path)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}
		return nil
	})
	return files, err
}
