package fileutil

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func GetLinesFromFile(fullPath string, lineNoStart int32, lineNoEnd int32) []string {
	file, err := os.Open(fullPath)
	if err != nil {
		log.Fatal(err.Error())
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
