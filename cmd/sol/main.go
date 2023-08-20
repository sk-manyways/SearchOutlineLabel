package main

import (
	"bufio"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/sk-manyways/SearchOutlineLabel/internal/fullfileinfo"
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
   return noPrefixArg, before, after
*/
func parseArgs(args []string) (string, int32, int32) {
	var noPrefixArg *string
	before := int32(0)
	after := int32(0)

	for idx, arg := range args {
		args[idx] = strings.TrimSpace(arg)
	}

	skip := false
	for idx, arg := range args {
		if skip {
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
				skip = true
			} else if arg[1:] == "B" {
				beforeCandidate, err := strconv.Atoi(args[idx+1])
				if err != nil {
					logging.Fatal(fmt.Sprintf("Invalid argument to B %s", args[idx+1]))
				}
				before = int32(beforeCandidate)
				skip = true
			} else {
				logging.Fatal(fmt.Sprintf("Unexpected arg %s", arg))
			}
		} else {
			duplicateArg := arg
			noPrefixArg = &duplicateArg
		}
	}

	if noPrefixArg == nil {
		logging.Fatal("Expected the noPrefixArg as input")
	}

	return *noPrefixArg, before, after
}

func printHelp() {
	fmt.Println("sol pathToScan\n" +
		"During execution: [-B int] [-A int] search\n" +
		"-B: print num lines of leading context before matching lines. \n" +
		"-A: print num lines of trailing context after matching lines.\n" +
		"\n" +
		"Note flags can be placed anywhere, e.g. this is valid: [-B int] search [-A int]")
}

func min(a int32, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func splitIncludingTerm(s string, term string) []string {
	if len(s) < len(term) {
		return make([]string, 0)
	}

	var result []string

	splitOnTerm := strings.Split(s, term)

	if s[0:len(term)] == term {
		result = append(result, term)
	}

	for idx, split := range splitOnTerm {
		if idx%2 == 1 {
			result = append(result, term)
		}
		result = append(result, split)
	}

	if len(s) > len(term) && s[len(s)-len(term):] == term {
		result = append(result, term)
	}

	return result
}

func main() {
	if len(os.Args) < 2 {
		logging.Fatal("Expected at least one argument - the path to scan")
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

	filesToScan := fullfileinfo.FindFilesRecursive(pathToScan, ignoreFileExtensions, ignoreDirectories, ignoreDirectoryWithPrefix)

	minWordLength := int32(4)
	limitLineLength := int32(120)

	newTrie := trie.NewTrie(minWordLength)
	for _, file := range filesToScan {
		newTrie.Add(file)
	}
	fmt.Printf("Found # files: %v\n", len(filesToScan))

	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4"))

	for true {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Search: ")
		userInput, _ := reader.ReadString('\n')
		split := strings.Split(userInput, " ")
		toSearchFor, linesBefore, linesAfter := parseArgs(split)
		toSearchFor = strings.ToLower(toSearchFor)
		searchResult, err := newTrie.Search(toSearchFor)
		if err != nil {
			fmt.Println("Error: " + err.Error())
		} else {
			for _, sr := range searchResult {
				fmt.Printf("Line: %v, Path: %v\n", sr.LineNumber, sr.FullPath())
				if linesBefore != 0 || linesAfter != 0 {
					lines := getLinesFromFile(sr.FullPath(), sr.LineNumber-linesBefore, sr.LineNumber+linesAfter+1)
					for _, line := range lines {
						lineLengthToShow := min(limitLineLength, int32(len(line)))
						lineLengthAdditional := ""
						if lineLengthToShow < int32(len(line)) {
							lineLengthAdditional = "..."
						}
						cappedLine := line[0:lineLengthToShow]

						lineSplitOnSearchTerm := splitIncludingTerm(strings.ToLower(cappedLine), toSearchFor)
						idxAt := 0
						for _, linePart := range lineSplitOnSearchTerm {
							lineInCorrectCase := cappedLine[idxAt : idxAt+len(linePart)]
							idxAt += len(linePart)
							if linePart == toSearchFor {
								fmt.Print(style.Render(lineInCorrectCase))
							} else {
								fmt.Print(lineInCorrectCase)
							}
						}

						fmt.Println(lineLengthAdditional)
					}
					fmt.Println()
				}
			}
		}
	}
}
