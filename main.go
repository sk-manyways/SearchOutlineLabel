package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

type FileInfoFull struct {
	fs.FileInfo
	FullPath string
}

func mayUseFile(file fs.FileInfo, ignoreFileExtensions map[string]struct{}) bool {
	extension := filepath.Ext(file.Name())

	if _, exists := ignoreFileExtensions[extension]; exists {
		return false
	}

	return true
}

func findFilesToScan(pathToScan string, ignoreFileExtensions map[string]struct{}, ignoreDirectories map[string]struct{}) []FileInfoFull {
	files, err := ioutil.ReadDir(pathToScan)

	if err != nil {
		fatal(err.Error())
	}

	var nextToScan []string
	var result []FileInfoFull

	for _, file := range files {
		if file.IsDir() {
			if _, exists := ignoreDirectories[file.Name()]; !exists {
				nextToScan = append(nextToScan, filepath.Join(pathToScan, file.Name()))
			}
		} else {
			if mayUseFile(file, ignoreFileExtensions) {
				abs, err := filepath.Abs(filepath.Join(pathToScan, file.Name()))
				if err != nil {
					fatal(err.Error())
				}

				result = append(result, FileInfoFull{
					FileInfo: file,
					FullPath: abs,
				})
			}
		}
	}

	for _, nextDir := range nextToScan {
		result = append(result, findFilesToScan(nextDir, ignoreFileExtensions, ignoreDirectories)...)
	}

	return result
}

func main() {
	if len(os.Args) < 2 {
		fatal("Expected at least one argument - the path to scan")
	}

	pathToScan := os.Args[1]

	var ignoreFileExtensions = make(map[string]struct{})
	ignoreFileExtensions[".class"] = struct{}{}
	ignoreFileExtensions[".jar"] = struct{}{}
	ignoreFileExtensions[".exe"] = struct{}{}
	ignoreFileExtensions[".jpg"] = struct{}{}
	ignoreFileExtensions[".jpeg"] = struct{}{}
	ignoreFileExtensions[".png"] = struct{}{}
	ignoreFileExtensions[".zip"] = struct{}{}
	ignoreFileExtensions[".7z"] = struct{}{}
	ignoreFileExtensions[".kotlin_module"] = struct{}{}
	ignoreFileExtensions[".iml"] = struct{}{}

	var ignoreDirectories = make(map[string]struct{})

	ignoreDirectories[".git"] = struct{}{}
	ignoreDirectories[".idea"] = struct{}{}

	filesToScan := findFilesToScan(pathToScan, ignoreFileExtensions, ignoreDirectories)

	for _, file := range filesToScan {
		fmt.Println(file.FullPath)
	}
}

func fatal(message string) {
	fmt.Println(message)
	os.Exit(1)
}
