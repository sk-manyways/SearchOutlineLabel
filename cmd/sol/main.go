package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/sk-manyways/SearchOutlineLabel/internal/configfile"
	"github.com/sk-manyways/SearchOutlineLabel/internal/fileutil"
	"github.com/sk-manyways/SearchOutlineLabel/internal/fullfileinfo"
	"github.com/sk-manyways/SearchOutlineLabel/internal/trie"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

/*
   return noPrefixArg, before, after
*/
func parseExecutionArgs(args []string) (*string, int32, int32, error) {
	var parseArgsErr error
	var noPrefixArg *string
	before := int32(-1)
	after := int32(-1)

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
				if len(args) <= idx+1 {
					parseArgsErr = errors.New(fmt.Sprintf("missing argument for A"))
					break
				} else {
					afterCandidate, err := strconv.Atoi(args[idx+1])
					if err != nil {
						parseArgsErr = errors.New(fmt.Sprintf("invalid argument to A %s", args[idx+1]))
					}
					after = int32(afterCandidate)
					skip = true
				}
			} else if arg[1:] == "B" {
				if len(args) <= idx+1 {
					parseArgsErr = errors.New(fmt.Sprintf("missing argument for B"))
					break
				} else {
					beforeCandidate, err := strconv.Atoi(args[idx+1])
					if err != nil {
						parseArgsErr = errors.New(fmt.Sprintf("invalid argument to B %s", args[idx+1]))
					}
					before = int32(beforeCandidate)
					skip = true
				}
			} else {
				parseArgsErr = errors.New(fmt.Sprintf("unexpected arg %s", arg))
			}
		} else {
			duplicateArg := arg
			noPrefixArg = &duplicateArg
		}
	}

	if noPrefixArg == nil {
		parseArgsErr = errors.New("expected a search term as input")
	}

	if before == -1 && after != -1 {
		before = 0
	} else if after == -1 && before != -1 {
		after = 0
	}
	return noPrefixArg, before, after, parseArgsErr
}

/*
   return pathToScan, additionalFileExtensionsToIgnore
*/
func parseArgs(args []string) (*string, []string, error) {
	var parseArgsErr error
	var pathToScan *string
	additionalFileExtensionsToIgnore := make([]string, 0)

	for idx, arg := range args {
		args[idx] = strings.TrimSpace(arg)
	}

	skip := 0
	for idx, arg := range args {
		if skip > 0 {
			skip--
			continue
		}
		if arg == "--help" {
			printHelp()
			os.Exit(0)
		} else if arg[0:1] == "-" {
			if arg[1:] == "EE" {
				if len(args) <= idx+1 {
					parseArgsErr = errors.New(fmt.Sprintf("missing argument for EE"))
					break
				} else {
					for k := idx + 1; k < len(args); k++ {
						extensionsCandidate := args[k]
						if extensionsCandidate[0:1] == "-" {
							break
						}
						skip++
						additionalFileExtensionsToIgnore = append(additionalFileExtensionsToIgnore, "."+args[k])
					}
				}
			} else {
				parseArgsErr = errors.New(fmt.Sprintf("unexpected arg %s", arg))
			}
		} else {
			duplicateArg := arg
			pathToScan = &duplicateArg
		}
	}

	if pathToScan == nil {
		parseArgsErr = errors.New("expected a pathToScan as input")
	}

	return pathToScan, additionalFileExtensionsToIgnore, parseArgsErr
}

func printHelp() {
	fmt.Println("sol pathToScan [-EE space delimited list] \n" +
		"-EE: excluded extensions, files with these extensions will not be searched; for example -EE exe sql\n" +
		"\n" +
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

	for idx, split := range splitOnTerm {
		if idx > 0 {
			result = append(result, term)
		}
		result = append(result, split)
	}

	return result
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Expected at least one argument - the path to scan")
	}
	pathToScan, additionalFileExtensionsToIgnore, err := parseArgs(os.Args[1:])

	if err != nil {
		log.Fatal(err.Error())
	}

	homeDir := getHomeDir()
	solDirPath := filepath.Join(homeDir, ".sol")
	solDirConfigPath := configfile.CreateDefaultConfig(solDirPath)

	var ignoreFileExtensions = make(map[string]struct{})
	baseFileExtensionsToIgnore := configfile.GetExcludedExtensions(solDirConfigPath)
	for _, ext := range baseFileExtensionsToIgnore {
		ext = "." + ext
		ignoreFileExtensions[ext] = struct{}{}
	}
	for _, ext := range additionalFileExtensionsToIgnore {
		ignoreFileExtensions[ext] = struct{}{}
	}

	var ignoreDirectories = make(map[string]struct{})
	baseDirectoriesToIgnore := configfile.GetExcludedDirectories(solDirConfigPath)
	for _, dir := range baseDirectoriesToIgnore {
		ignoreDirectories[dir] = struct{}{}
	}

	var ignoreDirectoryWithPrefix = make(map[string]struct{})
	ignoreDirectoryWithPrefix["."] = struct{}{}

	filesToScan := fullfileinfo.FindFilesRecursive(*pathToScan, ignoreFileExtensions, ignoreDirectories, ignoreDirectoryWithPrefix)

	minWordLength := configfile.GetMinWordLength(solDirConfigPath)
	limitLineLength := configfile.GetLimitLineLength(solDirConfigPath)

	matchForegroundColour := configfile.GetMatchForegroundColour(solDirConfigPath)
	matchBackgroundColour := configfile.GetMatchBackgroundColour(solDirConfigPath)

	fileMatchForegroundColour := configfile.GetFileMatchForegroundColour(solDirConfigPath)
	fileMatchBackgroundColour := configfile.GetFileMatchBackgroundColour(solDirConfigPath)

	newTrie := trie.NewTrie(minWordLength)
	for _, file := range filesToScan {
		newTrie.Add(file)
	}
	fmt.Printf("Found # files: %v\n", len(filesToScan))

	var styleMatch = lipgloss.NewStyle().
		Bold(true).
		Foreground("#" + lipgloss.Color(matchForegroundColour)).
		Background("#" + lipgloss.Color(matchBackgroundColour))

	var styleFileMatch = lipgloss.NewStyle().
		Bold(true).
		Foreground("#" + lipgloss.Color(fileMatchForegroundColour)).
		Background("#" + lipgloss.Color(fileMatchBackgroundColour)).
		Bold(true)

	for true {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Search: ")
		userInput, _ := reader.ReadString('\n')
		split := strings.Split(userInput, " ")
		toSearchForPtr, linesBefore, linesAfter, err := parseExecutionArgs(split)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		toSearchFor := *toSearchForPtr
		toSearchFor = strings.ToLower(toSearchFor)
		matchWord := true
		if toSearchFor[len(toSearchFor)-1] == '*' {
			toSearchFor = toSearchFor[0 : len(toSearchFor)-1]
			matchWord = false
		}
		searchResult, err := newTrie.Search(toSearchFor, matchWord)
		searchResultMap := createMapFromFileToTerminalNodes(searchResult)
		alreadyProcessed := make(map[string]struct{})
		if err != nil {
			fmt.Println("Error: " + err.Error())
		} else {
			for key, searchResultForFileList := range searchResultMap {
				if _, exists := alreadyProcessed[key]; exists {
					continue
				}

				if linesBefore != -1 || linesAfter != -1 {
					for _, sr := range searchResultForFileList {
						fmt.Print(styleFileMatch.Render(fmt.Sprintf("%v#%v", sr.FullPath(), sr.LineNumber)))
						fmt.Println()
						printBeforeAndAfterMatchLines(linesBefore, linesAfter, sr, limitLineLength, toSearchFor, styleMatch)
					}
				} else {
					lineNumbers := make([]int32, 0)
					for _, searchResultForFile := range searchResultForFileList {
						lineNumbers = append(lineNumbers, searchResultForFile.LineNumber)
					}
					fmt.Print(styleFileMatch.Render(fmt.Sprintf("%v", key)))
					fmt.Printf(" #%v", formatLineNumbersToString(lineNumbers))
					fmt.Println()
				}
			}
		}
	}
}

func formatLineNumbersToString(toConvert []int32) string {
	result := ""
	sort.Slice(toConvert, func(i, j int) bool { return toConvert[i] < toConvert[j] })
	for idx, num := range toConvert {
		converted := strconv.Itoa(int(num))
		if idx == 0 {
			result += converted
		} else {
			result += ", " + converted
		}
	}
	return result
}

func printBeforeAndAfterMatchLines(
	linesBefore int32,
	linesAfter int32,
	sr trie.TerminalNode,
	limitLineLength int32,
	toSearchFor string,
	styleMatch lipgloss.Style,
) {
	lines := fileutil.GetLinesFromFile(sr.FullPath(), sr.LineNumber-linesBefore, sr.LineNumber+linesAfter+1)
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
			// this maxIdx is done, because for cyrillic, characters are lost during toLower
			maxIdx := min(int32(len(cappedLine)), int32(idxAt+len(linePart)))
			lineInCorrectCase := cappedLine[idxAt:maxIdx]
			idxAt += len(linePart)
			if linePart == toSearchFor {
				fmt.Print(styleMatch.Render(lineInCorrectCase))
			} else {
				fmt.Print(lineInCorrectCase)
			}
		}

		fmt.Println(lineLengthAdditional)
	}
	fmt.Println()
}

func createMapFromFileToTerminalNodes(nodes []*trie.TerminalNode) map[string][]trie.TerminalNode {
	result := make(map[string][]trie.TerminalNode)
	if nodes == nil {
		return result
	}

	for _, node := range nodes {
		existing := result[node.FullPath()]
		if existing == nil {
			existing = make([]trie.TerminalNode, 0)
		}
		existing = append(existing, *node)
		result[node.FullPath()] = existing
	}

	return result
}

func getHomeDir() string {
	var homeDir string

	if runtime.GOOS == "windows" {
		homeDir = os.Getenv("USERPROFILE")
		if homeDir == "" {
			homeDir = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		}
	} else {
		homeDir = os.Getenv("HOME")
	}

	if homeDir == "" {
		log.Fatal("The home directory is not set.")
	}

	return homeDir
}
