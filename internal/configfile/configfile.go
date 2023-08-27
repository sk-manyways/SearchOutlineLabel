package configfile

import (
	"fmt"
	"github.com/sk-manyways/SearchOutlineLabel/internal/fileutil"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var fileContent = make(map[string]*[]string)

const sectionExclExtensions = "[excl-extensions]"
const sectionExclDirectories = "[excl-directories]"
const sectionMinWordLength = "[min-word-length]"
const sectionLimitLineLength = "[limit-line-length]"
const sectionMatchBackgroundColour = "[match-background-clr]"
const sectionMatchForegroundColour = "[match-foreground-clr]"
const sectionFileMatchBackgroundColour = "[file-match-background-clr]"
const sectionFileMatchForegroundColour = "[file-match-foreground-clr]"

var allSections = map[string]struct{}{
	sectionExclExtensions:        {},
	sectionExclDirectories:       {},
	sectionMinWordLength:         {},
	sectionLimitLineLength:       {},
	sectionMatchBackgroundColour: {},
	sectionMatchForegroundColour: {},
}

func GetExcludedExtensions(fullPath string) []string {
	if _, exists := fileContent[fullPath]; !exists {
		initFileContent(fullPath)
	}

	return getLinesInSection(sectionExclExtensions, *fileContent[fullPath])
}

func GetExcludedDirectories(fullPath string) []string {
	if _, exists := fileContent[fullPath]; !exists {
		initFileContent(fullPath)
	}

	return getLinesInSection(sectionExclDirectories, *fileContent[fullPath])
}

func GetMinWordLength(fullPath string) int32 {
	if _, exists := fileContent[fullPath]; !exists {
		initFileContent(fullPath)
	}

	return getSingleIntInSection(sectionMinWordLength, fullPath)
}

func GetLimitLineLength(fullPath string) int32 {
	if _, exists := fileContent[fullPath]; !exists {
		initFileContent(fullPath)
	}

	return getSingleIntInSection(sectionLimitLineLength, fullPath)
}

func GetMatchBackgroundColour(fullPath string) string {
	if _, exists := fileContent[fullPath]; !exists {
		initFileContent(fullPath)
	}

	return getSingleStringInSection(sectionMatchBackgroundColour, fullPath)
}

func GetMatchForegroundColour(fullPath string) string {
	if _, exists := fileContent[fullPath]; !exists {
		initFileContent(fullPath)
	}

	return getSingleStringInSection(sectionMatchForegroundColour, fullPath)
}

func GetFileMatchBackgroundColour(fullPath string) string {
	if _, exists := fileContent[fullPath]; !exists {
		initFileContent(fullPath)
	}

	return getSingleStringInSection(sectionFileMatchBackgroundColour, fullPath)
}

func GetFileMatchForegroundColour(fullPath string) string {
	if _, exists := fileContent[fullPath]; !exists {
		initFileContent(fullPath)
	}

	return getSingleStringInSection(sectionFileMatchForegroundColour, fullPath)
}

func getSingleIntInSection(section string, fullPath string) int32 {
	result, err := strconv.Atoi(getLinesInSection(section, *fileContent[fullPath])[0])
	if err != nil {
		log.Fatal(fmt.Sprintf("Could not read config for %s, message: %s", section, err.Error()))
	}
	return int32(result)
}

func getSingleStringInSection(section string, fullPath string) string {
	return getLinesInSection(section, *fileContent[fullPath])[0]
}

func getLinesInSection(section string, configFileContent []string) []string {
	inSection := false
	result := make([]string, 0)
	for _, line := range configFileContent {
		trimmedLine := trimLine(line)

		if trimmedLine == "" {
			continue
		}

		if trimmedLine == section {
			inSection = true
			continue
		}

		if inSection && sectionEnded(trimmedLine) {
			break
		}

		if inSection {
			result = append(result, trimmedLine)
		}
	}

	return result
}

func CreateDefaultConfig(directoryPath string) string {
	data := []byte(`[excl-extensions]
class
jar
exe
jpg
jpeg
png
zip
7z
kotlin_module
iml
gif
svg
ico
ttf
mp3
wav
pdf
mp4
mpeg
bin
dll
o

[excl-directories]
.git
.idea
node_modules
target
__pycache__
venv
lib
lib64
parts
sdist
dist

[min-word-length] # words shorter than this are not searchable; lower number = higher RAM usage
4

[limit-line-length] # output will be limited to 120 characters per search result
120

[match-background-clr]
7D56F4

[match-foreground-clr]
FAFAFA

[file-match-background-clr]
f47d56

[file-match-foreground-clr]
FAFAFA
`)
	if err := os.MkdirAll(directoryPath, 0755); err != nil {
		log.Fatal(err)
	}

	dest := filepath.Join(directoryPath, ".solconfig")
	_, err := os.Stat(dest)
	if os.IsNotExist(err) {
		err := ioutil.WriteFile(dest, data, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	return dest
}

func sectionEnded(line string) bool {
	lineTrimmed := trimLine(line)
	_, exists := allSections[lineTrimmed]
	return exists
}

func trimLine(line string) string {
	hashIdx := strings.Index(line, "#")
	var lineWithoutComment string
	if hashIdx == -1 {
		lineWithoutComment = line
	} else {
		lineWithoutComment = line[0:hashIdx]
	}
	return strings.TrimSpace(lineWithoutComment)
}

func initFileContent(fullPath string) {
	lines := fileutil.GetLinesFromFile(fullPath, 0, math.MaxInt32)
	fileContent[fullPath] = &lines
}
