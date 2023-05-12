package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func naiveSearch(text, pattern string) bool {
	for i := 0; i <= len(text)-len(pattern); i++ {
		if text[i:i+len(pattern)] == pattern {
			return true
		}
	}
	return false
}

func kmpSearch(text, pattern string) bool {
	lps := make([]int, len(pattern))
	length := 0
	i := 1

	for i < len(pattern) {
		if pattern[i] == pattern[length] {
			length++
			lps[i] = length
			i++
		} else {
			if length != 0 {
				length = lps[length-1]
			} else {
				lps[i] = 0
				i++
			}
		}
	}

	i = 0
	j := 0
	for i < len(text) {
		if pattern[j] == text[i] {
			i++
			j++
		}

		if j == len(pattern) {
			return true
		} else if i < len(text) && pattern[j] != text[i] {
			if j != 0 {
				j = lps[j-1]
			} else {
				i++
			}
		}
	}

	return false
}

func isFileType(filePath, fileType, algorithm string) (bool, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	text := string(data)
	switch algorithm {
	case "--naive":
		return naiveSearch(text, fileType), nil
	case "--KMP":
		return kmpSearch(text, fileType), nil
	default:
		return false, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

func main() {
	algorithm := os.Args[1]
	filePath := os.Args[2]
	fileType := os.Args[3]
	outputIfMatch := os.Args[4]

	start := time.Now()
	match, err := isFileType(filePath, fileType, algorithm)
	if err != nil {
		log.Fatal(err)
	}
	elapsed := time.Since(start)

	if !match {
		fmt.Println("Unknown file type")
	} else {
		fmt.Println(outputIfMatch)
	}
	fmt.Printf("It took %f seconds\n", elapsed.Seconds())
}
