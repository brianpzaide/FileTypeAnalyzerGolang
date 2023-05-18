package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type PatternPriority struct {
	Priority              int
	Pattern, DocumentType string
}

type ByPriority []PatternPriority

func (p ByPriority) Len() int {
	return len(p)
}

func (p ByPriority) Less(i, j int) bool {
	return -p[i].Priority < -p[j].Priority
}

func (p ByPriority) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p ByPriority) Sort() {
	sort.Sort(p)
}
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
func genPatternsDB(filePath string) ByPriority {
	patternStrings := make([]string, 0)
	patternPriority := make([]PatternPriority, 0)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		patternStrings = append(patternStrings, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	for _, pattern := range patternStrings {
		patternPieces := strings.Split(pattern, ";")
		i, e := strconv.Atoi(patternPieces[0])
		if e != nil {
			log.Fatal(err)
		}
		patternPriority = append(patternPriority,
			PatternPriority{Priority: i,
				Pattern:      patternPieces[1][1:(len(patternPieces[1]) - 1)],
				DocumentType: patternPieces[2][1:(len(patternPieces[2]) - 1)]})
	}

	patterns := ByPriority(patternPriority)
	patterns.Sort()
	return patterns

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
	patternsDBfile := os.Args[2]

	patterns := genPatternsDB(patternsDBfile)

	filePaths := getAllFileNames(folderName)
	// this introduces some async behaviour
	msgCh := make(chan string, len(filePaths))
	var wgWorkers sync.WaitGroup
	wgWorkers.Add(len(filePaths))

	for _, fileName := range filePaths {
		filePath := fileName
		go func() {
			defer wgWorkers.Done()
			// going through patterns in order from the highest priority to the least priority
			for _, pattern := range patterns {
				match, err := isFileType(filePath, pattern.Pattern, "--KMP")
				if err != nil {
					msgCh <- fmt.Sprintf("%s: %s", "Error", err.Error())
					return
				}

				if match {
					msgCh <- fmt.Sprintf("%s: %s", filepath.Base(filePath), pattern.DocumentType)
					return
				}
			}
			msgCh <- fmt.Sprintf("%s: %s", filepath.Base(filePath), "Unknown file type")
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
