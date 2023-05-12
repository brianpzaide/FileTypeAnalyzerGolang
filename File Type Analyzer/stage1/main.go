/*
[File Type Analyzer - Stage 1/5: Is this a PDF?](https://hyperskill.org/projects/64/stages/343/implement)
-------------------------------------------------------------------------------
[Primitive types](https://hyperskill.org/learn/topic/1807)
[Input/Output](https://hyperskill.org/learn/topic/1506)
[Slices](https://hyperskill.org/learn/topic/1672)
[Control statements](https://hyperskill.org/learn/topic/1728)
[String search](https://hyperskill.org/learn/topic/2063)
[Errors](https://hyperskill.org/learn/topic/1795)
[Command-line arguments and flags](https://hyperskill.org/learn/topic/1948)
[Reading files](https://hyperskill.org/learn/topic/1787)
*/

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func isFileType(filePath, fileType string) (bool, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(data), fileType), nil
}

func main() {
	filePath := os.Args[1]
	fileType := os.Args[2]
	outputIfMatch := os.Args[3]

	match, err := isFileType(filePath, fileType)
	if err != nil {
		log.Fatal(err)
	}

	if !match {
		fmt.Println("Unknown file type")
		return
	}
	fmt.Println(outputIfMatch)
}
