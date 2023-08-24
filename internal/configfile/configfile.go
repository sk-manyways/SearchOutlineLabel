package configfile

import (
	"github.com/sk-manyways/SearchOutlineLabel/internal/fileutil"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
)

var fileContent = make(map[string]*[]string)

const sectionExclExtensions = "[excl-extensions]"

const sectionExclDirectories = "[excl-directories]"

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

func getLinesInSection(section string, configFileContent []string) []string {
	inSection := false
	result := make([]string, 0)
	for _, line := range configFileContent {
		trimmedLine := strings.TrimSpace(line)

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
	return line == sectionExclExtensions || line == sectionExclDirectories
}

func initFileContent(fullPath string) {
	lines := fileutil.GetLinesFromFile(fullPath, 0, math.MaxInt32)
	fileContent[fullPath] = &lines
}
