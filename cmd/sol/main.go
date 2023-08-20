package main

import (
	"bufio"
	"fmt"
	"github.com/sk-manyways/SearchOutlineLabel/internal/fileinfo"
	"github.com/sk-manyways/SearchOutlineLabel/internal/logging"
	"github.com/sk-manyways/SearchOutlineLabel/internal/trie"
	"os"
	"strconv"
	"strings"
)

func getLinesFromFile(fullPath string, lineNoStart int32, lineNoEnd int32) []string {
	file, err := os.Open(fullPath)
	if err != nil {
		logging.Fatal(err.Error())
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
					logging.Fatal(fmt.Sprintf("Invalid argument to A %s", args[idx+1]))
				}
				after = int32(afterCandidate)
				displayContext = true
				skip = true
			} else if arg[1:] == "B" {
				beforeCandidate, err := strconv.Atoi(args[idx+1])
				if err != nil {
					logging.Fatal(fmt.Sprintf("Invalid argument to B %s", args[idx+1]))
				}
				before = int32(beforeCandidate)
				displayContext = true
				skip = true
			} else if arg[1:] == "D" {
				displayContext = true
			} else {
				logging.Fatal(fmt.Sprintf("Unexpected arg %s", arg))
			}
		} else {
			duplicateArg := arg
			pathToScan = &duplicateArg
		}
	}

	if pathToScan == nil {
		logging.Fatal("Expected the pathToScan as input")
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
		logging.Fatal("Expected at least one argument - the path to scan")
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

	filesToScan := fileinfo.FindFilesRecursive(pathToScan, ignoreFileExtensions, ignoreDirectories, ignoreDirectoryWithPrefix)

	minWordLength := int32(3)
	newTrie := trie.NewTrie(minWordLength)
	for _, file := range filesToScan {
		newTrie.Add(file)
	}
	fmt.Printf("Found # files: %v\n", len(filesToScan))

	for true {
		var userInput string
		fmt.Print("Search: ")
		fmt.Scanln(&userInput)
		userInput = strings.ToLower(userInput)
		searchResult, err := newTrie.Search(userInput)
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
