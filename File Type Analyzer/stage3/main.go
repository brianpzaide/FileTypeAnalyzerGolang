package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func getAllFileNames(folderPath string) []string {
	filesinfo, err := ioutil.ReadDir(folderPath)
	if err != nil {
		log.Fatal(err)
	}

	filePaths := make([]string, 0)

	for _, fileInfo := range filesinfo {
		absPath := filepath.Join(folderPath, fileInfo.Name())
		filePaths = append(filePaths, absPath)
	}
	return filePaths
}

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
	folderName := os.Args[1]
	fileType := os.Args[2]
	outputIfMatch := os.Args[3]

	filePaths := getAllFileNames(folderName)
	// this introduces some async behaviour
	msgCh := make(chan string, len(filePaths))
	var wgWorkers sync.WaitGroup // , wgChannelCloser sync.WaitGroup
	wgWorkers.Add(len(filePaths))

	for _, fileName := range filePaths {
		filePath := fileName
		go func() {
			defer wgWorkers.Done()
			var ft string
			match, err := isFileType(filePath, fileType, "--KMP")
			if err != nil {
				msgCh <- fmt.Sprintf("%s: %s", "Error", err.Error())
				return
			}

			if !match {
				ft = "Unknown file type"
			} else {
				ft = outputIfMatch
			}
			msgCh <- fmt.Sprintf("%s: %s", filepath.Base(filePath), ft)

		}()
	}

	wgWorkers.Wait()
	close(msgCh)
	for msg := range msgCh {
		if strings.HasPrefix(msg, "Error:") {
			log.Fatal(msg[6:])
		} else {
			fmt.Println(msg)
		}
	}
}
