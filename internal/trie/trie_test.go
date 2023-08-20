package trie

import (
	"github.com/sk-manyways/SearchOutlineLabel/internal/fullfileinfo"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewTrie(t *testing.T) {
	trie := NewTrie(5)
	assert.Equal(t, int32(5), trie.minWordLength)
	assert.NotEqual(t, nil, trie.root)
}

func TestTrie_addLineBasic1(t *testing.T) {
	trie := NewTrie(5)
	trie.addLine("hello There", createDummyFileInfo(), 20, false)
	trie.addLine("there are apples", createDummyFileInfo(), 21, false)

	// search for "hello"
	result, _ := trie.Search("hello")
	assert.Equal(t, 1, len(result))
	result1 := *result[0]
	assert.Equal(t, "/a/file.out", result1.FullPath())
	assert.Equal(t, int32(20), result1.LineNumber)

	// search for "there"
	result, _ = trie.Search("there")
	assert.Equal(t, 2, len(result))
	result1 = *result[0]
	assert.Equal(t, "/a/file.out", result1.FullPath())
	assert.Equal(t, int32(20), result1.LineNumber)
	result2 := *result[1]
	assert.Equal(t, "/a/file.out", result2.FullPath())
	assert.Equal(t, int32(21), result2.LineNumber)

	// search for "there1"
	result, _ = trie.Search("there1")
	assert.Equal(t, 0, len(result))

	// search for "are"
	result, _ = trie.Search("are")
	assert.Equal(t, 0, len(result)) // not found because of min word length is 5
}

func TestTrie_addLineUnderscore1(t *testing.T) {
	trie := NewTrie(5)
	trie.addLine("hello There a_request, 123", createDummyFileInfo(), 20, false)
	trie.addLine("there are apples", createDummyFileInfo(), 21, false)

	// search for "a_request"
	result, _ := trie.Search("a_request")
	assert.Equal(t, 1, len(result))
	result1 := *result[0]
	assert.Equal(t, "/a/file.out", result1.FullPath())
	assert.Equal(t, int32(20), result1.LineNumber)
}

func TestTrie_addLineNumbers1(t *testing.T) {
	trie := NewTrie(2)
	trie.addLine("hello There a_request, 123", createDummyFileInfo(), 20, false)
	trie.addLine("there are apples 12, z1", createDummyFileInfo(), 21, false)

	// search for "123"
	result, _ := trie.Search("123")
	assert.Equal(t, 1, len(result))
	result1 := *result[0]
	assert.Equal(t, "/a/file.out", result1.FullPath())
	assert.Equal(t, int32(20), result1.LineNumber)

	// search for "z1"
	result, _ = trie.Search("z1")
	assert.Equal(t, 1, len(result))
	result1 = *result[0]
	assert.Equal(t, "/a/file.out", result1.FullPath())
	assert.Equal(t, int32(21), result1.LineNumber)
}

func TestTrie_Add(t *testing.T) {
	trie := NewTrie(1)
	fileInfo, _ := os.Stat("./testdata/test_trie_add_1.txt")
	trie.Add(fullfileinfo.NewFull(fileInfo, "./testdata/test_trie_add_1.txt"))

	// search for "main"
	result, _ := trie.Search("main") // should only find two results, because we are consolidating on lineNumber, meaning line 3 should only appear once in the search results
	assert.Equal(t, 2, len(result))
	result1 := *result[0]
	assert.Equal(t, "./testdata/test_trie_add_1.txt", result1.FullPath())
	assert.Equal(t, int32(2), result1.LineNumber)
	result2 := *result[1]
	assert.Equal(t, "./testdata/test_trie_add_1.txt", result2.FullPath())
	assert.Equal(t, int32(3), result2.LineNumber)

	// search for "0"
	result, _ = trie.Search("0")
	assert.Equal(t, 1, len(result))
	result1 = *result[0]
	assert.Equal(t, "./testdata/test_trie_add_1.txt", result1.FullPath())
	assert.Equal(t, int32(4), result1.LineNumber)

	// search for "random"
	result, _ = trie.Search("random")
	assert.Equal(t, 1, len(result))
	result1 = *result[0]
	assert.Equal(t, "./testdata/test_trie_add_1.txt", result1.FullPath())
	assert.Equal(t, int32(1), result1.LineNumber)

	// search for "random"
	result, _ = trie.Search("line")
	assert.Equal(t, 1, len(result))
	result1 = *result[0]
	assert.Equal(t, "./testdata/test_trie_add_1.txt", result1.FullPath())
	assert.Equal(t, int32(6), result1.LineNumber)
}

func createDummyFileInfo() fullfileinfo.Full {
	return fullfileinfo.NewFull(nil, "/a/file.out")
}
