package trie

import (
	"errors"
	"fmt"
	"github.com/sk-manyways/SearchOutlineLabel/internal/fileinfo"
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
}

type TerminalNode struct {
	fileinfo.Full
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

func (trie Trie) AddLine(line string, file fileinfo.Full, lineNumber int32) {
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

func (trie Trie) Add(fileInput fileinfo.Full) {
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
		trie.AddLine(line, fileInput, lineNumber)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(fmt.Sprintf("Error scanning fileinfo %v, error: %v", fileInput.FullPath(), err.Error()))
	}
}
