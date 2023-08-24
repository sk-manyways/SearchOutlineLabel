package configfile

import (
	"github.com/sk-manyways/SearchOutlineLabel/internal/fileutil"
	"math"
)

var fileContent *[]string = nil

const sectionExclExtensions = "excl-extensions"

const sectionExclDirectories = "excl-directories"

func GetExcludedExtensions(fullPath string) []string {
	if fileContent == nil {
		initFileContent(fullPath)
	}

	return getLinesInSection(sectionExclExtensions, *fileContent)
}

func GetExcludedDirectories(fullPath string) []string {
	if fileContent == nil {
		initFileContent(fullPath)
	}

	return getLinesInSection(sectionExclDirectories, *fileContent)
}

func getLinesInSection(section string, configFileContent []string) []string {
	inSection := false
	result := make([]string, 0)
	for _, line := range configFileContent {

		if line == section {
			inSection = true
			continue
		}

		if sectionEnded(line) {
			break
		}

		if inSection {
			result = append(result, line)
		}
	}

	return result
}

func sectionEnded(line string) bool {
	return line == sectionExclExtensions || line == sectionExclDirectories
}

func initFileContent(fullPath string) {
	lines := fileutil.GetLinesFromFile(fullPath, 0, math.MaxInt32)
	fileContent = &lines
}
