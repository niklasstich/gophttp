package common

import (
	"io/fs"
	"os"
	"path/filepath"
)

func ListFilesRecursive(directory string) ([]string, error) {
	return listEntriesRecursive(directory, true)
}

func ListDirsRecursive(directory string) ([]string, error) {
	return listEntriesRecursive(directory, false)
}

func listEntriesRecursive(directory string, listFiles bool) ([]string, error) {
	var entries []string
	err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if listFiles != d.IsDir() && (listFiles || d.IsDir()) {
			//make path relative to directory
			rel, err := filepath.Rel(directory, path)
			if err != nil {
				return err
			}
			entries = append(entries, rel)
		}
		return nil
	})
	return entries, err
}

func FilesInDirectory(directory string) ([]string, error) {
	return entriesInDirectory(directory, true)
}

func DirsInDirectory(directory string) ([]string, error) {
	return entriesInDirectory(directory, false)
}

func entriesInDirectory(directory string, listFiles bool) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var entriesFiltered []string
	for _, entry := range entries {
		if listFiles != entry.IsDir() && (listFiles || entry.IsDir()) {
			entriesFiltered = append(entriesFiltered, entry.Name())
		}
	}

	return entriesFiltered, nil
}
