package trie

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/sk-manyways/SearchOutlineLabel/internal/fullfileinfo"
	"github.com/sk-manyways/SearchOutlineLabel/internal/logging"
	"os"
	"strings"
)

type Trie struct {
	root          *TrieNode
	minWordLength int32
}

type TrieNode struct {
	children      []*TrieNode
	terminalNodes []*TerminalNode
	lineNumbers   map[int32]struct{}
}

type TerminalNode struct {
	fullfileinfo.Full
	LineNumber int32
}

func NewTrie(minWordLength int32) *Trie {
	return &Trie{
		// this first, simple version, will just work with the 26 letters of the alphabet + 10 numbers
		root:          newTrieNode(),
		minWordLength: minWordLength,
	}
}

func newTrieNode() *TrieNode {
	return &TrieNode{
		// this first, simple version, will just work with the 26 letters of the alphabet + 10 numbers
		children:    make([]*TrieNode, 37),
		lineNumbers: make(map[int32]struct{}),
	}
}

func (trie Trie) Search(searchTerm string, matchWord bool) ([]*TerminalNode, error) {
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
		if matchWord {
			result = atNode.terminalNodes
		} else {
			result = findAllChildNodeChildren(*atNode)
		}
	}

	return result, err
}

func findAllChildNodeChildren(node TrieNode) []*TerminalNode {
	var result = make([]*TerminalNode, 0)

	result = append(result, node.terminalNodes...)

	for i := 0; i < len(node.children); i++ {
		child := node.children[i]
		if child != nil {
			result = append(result, findAllChildNodeChildren(*child)...)
		}
	}

	return result
}

func determineIdx(c byte) *uint8 {
	var idx *uint8
	if c >= 'a' && c <= 'z' {
		tmp := c - 'a'
		idx = &tmp
	} else if c >= '0' && c <= '9' {
		tmp := c - '0' + 25
		idx = &tmp
	} else if c == '_' {
		tmp := uint8(35 + 1)
		idx = &tmp
	}
	return idx
}

func (trie Trie) addLine(line string, file fullfileinfo.Full, lineNumber int32, consolidateOnLineNumber bool) {
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
			if mayUseWord(trie, *atNode, wordLength, lineNumber, consolidateOnLineNumber) {
				atNode.lineNumbers[lineNumber] = struct{}{}
				atNode.terminalNodes = append(atNode.terminalNodes, &TerminalNode{
					file,
					lineNumber,
				})
			}
			atNode = trie.root
			wordLength = 0
		}
	}

	if mayUseWord(trie, *atNode, wordLength, lineNumber, consolidateOnLineNumber) {
		atNode.terminalNodes = append(atNode.terminalNodes, &TerminalNode{
			file,
			lineNumber,
		})
	}
}

func mayUseWord(trie Trie, node TrieNode, wordLength int32, lineNumber int32, consolidateOnLineNumber bool) bool {
	if wordLength >= trie.minWordLength {
		if _, exists := node.lineNumbers[lineNumber]; !(consolidateOnLineNumber && exists) {
			return true
		}
	}
	return false
}

func (trie Trie) Add(fileInput fullfileinfo.Full) {
	file, err := os.Open(fileInput.FullPath())
	if err != nil {
		logging.Fatal(err.Error())
	}
	defer file.Close()

	scanner := createScanner(file)

	lineNumber := int32(0)
	for scanner.Scan() {
		lineNumber += 1
		line := scanner.Text()
		trie.addLine(line, fileInput, lineNumber, true)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(fmt.Sprintf("Error scanning fullfileinfo %v, error: %v", fileInput.FullPath(), err.Error()))
	}
}

func createScanner(file *os.File) *bufio.Scanner {
	scanner := bufio.NewScanner(file)

	// Double the default buffer size
	const maxCapacity = 2048 * 1024 // 2048KB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)
	return scanner
}
