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

const (
	a uint64 = 53
	m uint64 = 1000000009
)

type PatternPriority struct {
	Priority              int
	Pattern, DocumentType string
}

// ByPriority implements the sort interface, sorting is based on priority
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

func rabinKarpSearch(text, pattern string) bool {
	textRunes := []byte(text)
	patternRunes := []byte(pattern)

	textLength := len(text)
	patternLength := len(pattern)

	if patternLength > textLength {
		return false
	}
	if patternLength == textLength {
		return text == pattern
	}

	var patternHash uint64 = 0
	var currSubstrHash uint64 = 0
	var pow uint64 = 1

	for i := 0; i < len(patternRunes); i++ {
		patternHash += uint64(patternRunes[i]) * pow
		patternHash %= m

		currSubstrHash += uint64(textRunes[textLength-patternLength+i]) * pow
		currSubstrHash %= m

		if i != patternLength-1 {
			pow = pow * a % m
		}
	}

	for i := textLength; i >= patternLength; i-- {
		if patternHash == currSubstrHash {
			patternIsFound := true

			for j := 0; j < patternLength; j++ {
				if textRunes[i-patternLength+j] != patternRunes[j] {
					patternIsFound = false
					break
				}
			}
			if patternIsFound {
				return true
			}
		}
		if i > patternLength {
			currSubstrHash = (currSubstrHash - uint64(textRunes[i-1])*pow%m + m) * a % m
			currSubstrHash = (currSubstrHash + uint64(textRunes[i-patternLength-1])) % m
		}
	}
	return false
}

// reads the patterns file, generates a sorted list(in decreasing order of priority) of patterns
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
	// sorting the patterns in decreasing order of priority
	patterns.Sort()
	//for _, p := range patterns {
	//	fmt.Printf("%d : %s : %s\n", p.Priority, p.Pattern, p.DocumentType)
	//}
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
	case "--RB":
		return rabinKarpSearch(text, fileType), nil
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
	var wgWorkers, wgChannelCloser sync.WaitGroup
	wgWorkers.Add(len(filePaths))
	// main goroutine is the only one that throws exception,
	//I wanted all the other goroutines to finish before any exception happens
	wgChannelCloser.Add(1)

	// starting a new goroutine taht
	go func() {
		wgWorkers.Wait()
		close(msgCh)
		wgChannelCloser.Done()
	}()

	// start := time.Now()
	for _, fileName := range filePaths {
		filePath := fileName
		go func() {
			defer wgWorkers.Done()
			// going through patterns in order from highest priority to the least priority
			for _, pattern := range patterns {
				match, err := isFileType(filePath, pattern.Pattern, "--RB")
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

	wgChannelCloser.Wait()
	for msg := range msgCh {
		if strings.HasPrefix(msg, "Error:") {
			log.Fatal(msg[6:])
		} else {
			fmt.Println(msg)
		}
	}
	// elapsed := time.Since(start)
	// fmt.Printf("It took %f seconds\n", elapsed.Seconds())
}
