package configfile

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetExcludedDirectories(t *testing.T) {
	excludedDirs := GetExcludedDirectories("testdata/.solconfig-1")

	expected := [...]string{
		".git",
		".idea",
		"node_modules",
		"target",
		"__pycache__",
		"venv",
		"lib",
		"lib64",
		"parts",
		"sdist",
		"dist",
	}

	for i, _ := range excludedDirs {
		assert.Equal(t, expected[i], excludedDirs[i])
	}
}

func TestGetExcludedExtensions(t *testing.T) {
	excludedDirs := GetExcludedExtensions("testdata/.solconfig-1")

	expected := [...]string{
		"class",
		"jar",
		"exe",
		"jpg",
		"jpeg",
		"png",
		"zip",
		"7z",
		"kotlin_module",
		"iml",
		"gif",
		"svg",
		"ico",
		"ttf",
		"mp3",
		"wav",
		"pdf",
		"mp4",
		"mpeg",
		"bin",
		"dll",
	}

	for i, _ := range excludedDirs {
		assert.Equal(t, expected[i], excludedDirs[i])
	}
}
