package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Trie struct {
	root *TrieNode
}

type TrieNode struct {
	children      []*TrieNode
	terminalNodes []*TerminalNode
}

type TerminalNode struct {
	FileInfoFull
	LineNumber int32
}

func newTrie() *Trie {
	return &Trie{
		// this first, simple version, will just work with the 26 letters of the alphabet + 10 numbers
		root: newTrieNode(),
	}
}

func newTrieNode() *TrieNode {
	return &TrieNode{
		// this first, simple version, will just work with the 26 letters of the alphabet + 10 numbers
		children: make([]*TrieNode, 36),
	}
}

func (trie Trie) search(searchTerm string) ([]*TerminalNode, error) {
	var result []*TerminalNode
	var err error
	var atNode = trie.root
	didComplete := true
	for i := 0; i < len(searchTerm); i++ {
		c := searchTerm[i]
		idx := determineIdx(c)

		if idx == nil {
			err = errors.New(fmt.Sprintf("Invalid character in search query %s", string(c)))
		} else {
			targetChild := atNode.children[*idx]
			if targetChild != nil {
				atNode = targetChild
			} else {
				didComplete = false
			}
		}
	}
	if didComplete {
		result = atNode.terminalNodes
	}

	return result, err
}

func determineIdx(c byte) *uint8 {
	if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '0') {
		var idx uint8
		if c >= 'a' && c <= 'z' {
			idx = c - 'a'
		} else {
			idx = c - '0' + 26
		}
		return &idx
	}
	return nil
}

func (trie Trie) AddLine(line string, file FileInfoFull, lineNumber int32) {
	line = strings.ToLower(line)
	atNode := trie.root
	wordLength := 0
	for i := 0; i < len(line); i++ {
		c := line[i]
		idx := determineIdx(c)
		if idx != nil {
			wordLength++
			targetChild := atNode.children[*idx]
			if targetChild == nil {
				targetChild = newTrieNode()
				atNode.children[*idx] = targetChild
			}
			atNode = targetChild
		} else {
			// create terminal node
			if wordLength > 0 {
				atNode.terminalNodes = append(atNode.terminalNodes, &TerminalNode{
					file,
					lineNumber,
				})
			}
			atNode = trie.root
			wordLength = 0
		}
	}
}

func (trie Trie) Add(fileInput FileInfoFull) {
	file, err := os.Open(fileInput.FullPath())
	if err != nil {
		fatal(err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Double the default buffer size
	const maxCapacity = 2048 * 1024 // 2048KB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	lineNumber := int32(0)
	for scanner.Scan() {
		lineNumber += 1
		line := scanner.Text()
		trie.AddLine(line, fileInput, lineNumber)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(fmt.Sprintf("Error scanning file %v, error: %v", fileInput.FullPath(), err.Error()))
	}
}

type FileInfoFull struct {
	fs.FileInfo
	fullPath string
}

func (f FileInfoFull) FullPath() string {
	return f.fullPath
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
					fullPath: abs,
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
	ignoreFileExtensions[".mp3"] = struct{}{}
	ignoreFileExtensions[".wav"] = struct{}{}
	ignoreFileExtensions[".pdf"] = struct{}{}
	ignoreFileExtensions[".mp4"] = struct{}{}
	ignoreFileExtensions[".mpeg"] = struct{}{}
	ignoreFileExtensions[".bin"] = struct{}{}

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

	trie := newTrie()
	for _, file := range filesToScan {
		trie.Add(file)
	}
	fmt.Printf("Found # files: %v\n", len(filesToScan))

	for true {
		var userInput string
		fmt.Print("Search: ")
		fmt.Scanln(&userInput)
		userInput = strings.ToLower(userInput)
		searchResult, err := trie.search(userInput)
		if err != nil {
			fmt.Println("Error: " + err.Error())
		} else {
			for _, sr := range searchResult {
				fmt.Printf("Line: %v, Path: %v\n", sr.LineNumber, sr.FullPath())
			}
		}
	}
}

func fatal(message string) {
	fmt.Println(message)
	os.Exit(1)
}
