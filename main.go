package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type FileInfoFull struct {
	fs.FileInfo
	FullPath string
}

func mayUseFile(file fs.FileInfo, ignoreFileExtensions map[string]struct{}) bool {
	extension := strings.ToLower(filepath.Ext(file.Name()))

	if _, exists := ignoreFileExtensions[extension]; exists {
		return false
	}

	return true
}

func mayUseDirectory(file fs.FileInfo, ignoreDirectories map[string]struct{}, ignoreDirectoryWithPrefix map[string]struct{}) bool {
	fileName := strings.ToLower(file.Name())
	firstChar := string(fileName[0])
	if _, exists := ignoreDirectoryWithPrefix[firstChar]; exists {
		return false
	}

	if _, exists := ignoreDirectories[fileName]; exists {
		return false
	}

	return true
}

func findFilesToScan(pathToScan string,
	ignoreFileExtensions map[string]struct{},
	ignoreDirectories map[string]struct{},
	ignoreDirectoryWithPrefix map[string]struct{}) []FileInfoFull {
	files, err := ioutil.ReadDir(pathToScan)

	if err != nil {
		fatal(err.Error())
	}

	var nextToScan []string
	var result []FileInfoFull

	for _, file := range files {
		if file.IsDir() {
			if mayUseDirectory(file, ignoreDirectories, ignoreDirectoryWithPrefix) {
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
		result = append(result, findFilesToScan(nextDir, ignoreFileExtensions, ignoreDirectories, ignoreDirectoryWithPrefix)...)
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
	ignoreFileExtensions[".gif"] = struct{}{}
	ignoreFileExtensions[".svg"] = struct{}{}
	ignoreFileExtensions[".ico"] = struct{}{}
	ignoreFileExtensions[".ttf"] = struct{}{}

	var ignoreDirectories = make(map[string]struct{})
	ignoreDirectories[".git"] = struct{}{}
	ignoreDirectories[".idea"] = struct{}{}
	ignoreDirectories["node_modules"] = struct{}{}
	ignoreDirectories["target"] = struct{}{}
	ignoreDirectories["__pycache__"] = struct{}{}
	ignoreDirectories["venv"] = struct{}{}

	var ignoreDirectoryWithPrefix = make(map[string]struct{})
	ignoreDirectoryWithPrefix["."] = struct{}{}

	filesToScan := findFilesToScan(pathToScan, ignoreFileExtensions, ignoreDirectories, ignoreDirectoryWithPrefix)

	for _, file := range filesToScan {
		fmt.Println(file.FullPath)
	}
}

func fatal(message string) {
	fmt.Println(message)
	os.Exit(1)
}
