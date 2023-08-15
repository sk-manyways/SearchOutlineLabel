package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Trie struct {
	root          *TrieNode
	minWordLength int32
}

type TrieNode struct {
	children      []*TrieNode
	terminalNodes []*TerminalNode
}

type TerminalNode struct {
	FileInfoFull
	LineNumber int32
}

func newTrie(minWordLength int32) *Trie {
	return &Trie{
		// this first, simple version, will just work with the 26 letters of the alphabet + 10 numbers
		root:          newTrieNode(),
		minWordLength: minWordLength,
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
	wordLength := int32(0)
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
			// create a terminal node
			if wordLength >= trie.minWordLength {
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

	scanner := createScanner(file)

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

func getLinesFromFile(fullPath string, lineNoStart int32, lineNoEnd int32) []string {
	file, err := os.Open(fullPath)
	if err != nil {
		fatal(err.Error())
	}
	defer file.Close()

	scanner := createScanner(file)

	var result []string

	lineNumber := int32(0)
	for scanner.Scan() {
		lineNumber += 1
		if lineNumber >= lineNoStart && lineNumber < lineNoEnd {
			line := scanner.Text()
			result = append(result, line)
		} else if lineNumber >= lineNoEnd {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(fmt.Sprintf("Error on line number scanning for file %v, error: %v", fullPath, err.Error()))
	}

	return result
}

func createScanner(file *os.File) *bufio.Scanner {
	scanner := bufio.NewScanner(file)

	// Double the default buffer size
	const maxCapacity = 2048 * 1024 // 2048KB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	return scanner
}

/*
   return pathToScan, before, after, displayContext
*/
func parseArgs(args []string) (string, int32, int32, bool) {
	if len(args) < 2 {
		printHelp()
		os.Exit(1)
	}

	var pathToScan *string
	before := int32(0)
	after := int32(0)
	displayContext := false

	skip := false
	for idx, arg := range args {
		if idx == 0 {
			continue
		} else if skip {
			skip = false
			continue
		}
		if arg == "--help" {
			printHelp()
			os.Exit(0)
		} else if arg[0:1] == "-" {
			if arg[1:] == "A" {
				afterCandidate, err := strconv.Atoi(args[idx+1])
				if err != nil {
					fatal(fmt.Sprintf("Invalid argument to A %s", args[idx+1]))
				}
				after = int32(afterCandidate)
				displayContext = true
				skip = true
			} else if arg[1:] == "B" {
				beforeCandidate, err := strconv.Atoi(args[idx+1])
				if err != nil {
					fatal(fmt.Sprintf("Invalid argument to B %s", args[idx+1]))
				}
				before = int32(beforeCandidate)
				displayContext = true
				skip = true
			} else if arg[1:] == "D" {
				displayContext = true
			} else {
				fatal(fmt.Sprintf("Unexpected arg %s", arg))
			}
		} else {
			duplicateArg := arg
			pathToScan = &duplicateArg
		}
	}

	if pathToScan == nil {
		fatal("Expected the pathToScan as input")
	}

	return *pathToScan, before, after, displayContext
}

func printHelp() {
	fmt.Println("sol pathToScan [-B int] [-A int] [-D]\n" +
		"-B: print num lines of leading context before matching lines. \n" +
		"-A: print num lines of trailing context after matching lines. \n" +
		"-D: display the context lines, will be true if -B or -A is used")
}

func main() {
	if len(os.Args) < 2 {
		fatal("Expected at least one argument - the path to scan")
	}

	pathToScan, linesBefore, linesAfter, displayContext := parseArgs(os.Args)

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
	ignoreFileExtensions[".dll"] = struct{}{}

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

	minWordLength := int32(3)
	trie := newTrie(minWordLength)
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
				if displayContext {
					lines := getLinesFromFile(sr.FullPath(), sr.LineNumber-linesBefore, sr.LineNumber+linesAfter+1)
					for _, line := range lines {
						fmt.Println(line)
					}
					fmt.Println()
				}
			}
		}
	}
}

func fatal(message string) {
	fmt.Println(message)
	os.Exit(1)
}
