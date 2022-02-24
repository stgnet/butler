package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

var extPriority = map[string]int{
	".yaml": 1,
	".json": 2,
	".html": 3,
	".php":  4,
}

func extSort(a string, b string) bool {
	aPri, aExists := extPriority[filepath.Ext(a)]
	if !aExists {
		aPri = 99999
	}

	bPri, bExists := extPriority[filepath.Ext(b)]
	if !bExists {
		bPri = 99999
	}

	return aPri > bPri
}

func matches(dir string, target string) []string {
	files, rdErr := ioutil.ReadDir(dir)
	if rdErr != nil {
		panic(fmt.Errorf("Unable to read dir '%s': %v", dir, rdErr))
	}
	matches := []string{}

	for _, file := range files {
		fmt.Println(file.Name(), file.IsDir())
	}
	sort.Slice(matches, func(i, j int) bool { return extSort(matches[i], matches[2]) })
	return matches
}

func locate(w io.Writer, path string) {
	fmt.Printf("Locate: '%s'\n", path)
	info, pathErr := os.Stat(path)
	if pathErr == nil {
		// path exists, but is it a file?
		if !info.IsDir() {
			// no need to locate it, just render it
			renderFile(w, path)
		}
		// path is a directory, we need an index
		locate(w, filepath.Join(path, "index"))
		return
	}

	dir, file := filepath.Split(path)
	ext := filepath.Ext(file)
	justfile := file[0 : len(file)-len(ext)]
	fmt.Printf("dir '%s' file '%s' just '%s' ext '%s'\n", dir, file, justfile, ext)

	if dir == "" {
		dir = "."
	}

	// need to scan the directory and determine what files might match
}
