package main

import (
	"io/fs"
	"path/filepath"

	"github.com/goyek/goyek/v2"
)

func find(a *goyek.A, ext string) []string {
	a.Helper()
	var files []string
	err := filepath.WalkDir(dirRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(d.Name()) == ext {
			files = append(files, filepath.ToSlash(path))
		}
		return nil
	})
	if err != nil {
		a.Fatal(err)
	}
	return files
}
