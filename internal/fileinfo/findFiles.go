package fileinfo

import (
	"github.com/sk-manyways/SearchOutlineLabel/internal/logging"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"
)

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

func mayUseFile(file fs.FileInfo, ignoreFileExtensions map[string]struct{}) bool {
	extension := strings.ToLower(filepath.Ext(file.Name()))

	if _, exists := ignoreFileExtensions[extension]; exists {
		return false
	}

	return true
}

func FindFilesRecursive(pathToScan string,
	ignoreFileExtensions map[string]struct{},
	ignoreDirectories map[string]struct{},
	ignoreDirectoryWithPrefix map[string]struct{}) []Full {
	files, err := ioutil.ReadDir(pathToScan)

	if err != nil {
		logging.Fatal(err.Error())
	}

	var nextToScan []string
	var result []Full

	for _, file := range files {
		if file.IsDir() {
			if mayUseDirectory(file, ignoreDirectories, ignoreDirectoryWithPrefix) {
				nextToScan = append(nextToScan, filepath.Join(pathToScan, file.Name()))
			}
		} else {
			if mayUseFile(file, ignoreFileExtensions) {
				abs, err := filepath.Abs(filepath.Join(pathToScan, file.Name()))
				if err != nil {
					logging.Fatal(err.Error())
				}

				result = append(result, NewFull(file, abs))
			}
		}
	}

	for _, nextDir := range nextToScan {
		result = append(result, FindFilesRecursive(nextDir, ignoreFileExtensions, ignoreDirectories, ignoreDirectoryWithPrefix)...)
	}

	return result
}
