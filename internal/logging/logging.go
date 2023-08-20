package logging

import (
	"fmt"
	"os"
)

func Fatal(message string) {
	fmt.Println(message)
	os.Exit(1)
}
