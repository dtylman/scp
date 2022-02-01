package main

import (
	"os"
	"path/filepath"
	"regexp"
)

//FilesWalker searches pattern matching files by regex
type FilesWalker struct {
	//Matches the matches found by walker
	Matches    []string
	root       string
	expression *regexp.Regexp
}

//NewFilesWalker creates a new files walker with the provided reg-ex for matching
func NewFilesWalker(root string, expr string) (*FilesWalker, error) {
	fw := &FilesWalker{
		root: root,
	}
	var err error
	fw.expression, err = regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return fw, nil
}

//Walk scans the filesystem
func (fw *FilesWalker) Walk() error {
	fw.Matches = make([]string, 0)
	return filepath.WalkDir(fw.root, fw.walkDirFunc)
}

func (fw *FilesWalker) walkDirFunc(path string, d os.DirEntry, err error) error {
	if d == nil {
		return nil
	}
	if d.IsDir() {
		return nil
	}
	if fw.expression.MatchString(path) {
		fw.Matches = append(fw.Matches, path)
	}
	return nil
}
