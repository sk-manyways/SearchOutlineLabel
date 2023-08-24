package configfile

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestGetExcludedDirectories(t *testing.T) {
	excludedDirs := GetExcludedDirectories("testdata/.solconfig-1")

	expected := getExpectedDirs()

	assert.Equal(t, len(expected), len(excludedDirs))

	for i, _ := range excludedDirs {
		assert.Equal(t, expected[i], excludedDirs[i])
	}
}

func getExpectedDirs() [11]string {
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
	return expected
}

func TestGetExcludedExtensions(t *testing.T) {
	excludedExtensions := GetExcludedExtensions("testdata/.solconfig-1")

	expected := getExpectedExtensions()

	assert.Equal(t, len(expected), len(excludedExtensions))

	for i, _ := range excludedExtensions {
		assert.Equal(t, expected[i], excludedExtensions[i])
	}
}

func getExpectedExtensions() [21]string {
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
	return expected
}

func TestCreateDefaultConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tmpDir")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tempDir)

	finalPath := filepath.Join(tempDir, ".sol")
	CreateDefaultConfig(finalPath)

	excludedExtensions := GetExcludedExtensions(filepath.Join(finalPath, ".solconfig"))

	expected := getExpectedExtensions()

	for i, _ := range excludedExtensions {
		assert.Equal(t, expected[i], excludedExtensions[i])
	}

	excludedDirs := GetExcludedDirectories(filepath.Join(finalPath, ".solconfig"))

	expected2 := getExpectedDirs()

	for i, _ := range excludedDirs {
		assert.Equal(t, expected2[i], excludedDirs[i])
	}
}
