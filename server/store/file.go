package store

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	wg   = sync.WaitGroup{}
	lock = sync.Mutex{}
)

func WriteFileWithDirectories(fp string, data []byte) error {
	dir := filepath.Dir(fp)

	// Create directories recursively if they don't exist
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	// Write file using ioutil.WriteFile
	err = ioutil.WriteFile(fp, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func ReadLatestFromFile(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lastEntry string

	for scanner.Scan() {
		lastEntry = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return lastEntry, nil
}

func TruncateFile(filepath string, linesToRemove int) error {
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if linesToRemove > len(lines) {
		linesToRemove = len(lines)
	}

	truncatedLines := lines[linesToRemove:]
	truncatedContent := []byte("")
	for _, line := range truncatedLines {
		truncatedContent = append(truncatedContent, []byte(line+"\n")...)
	}

	if err := file.Truncate(0); err != nil {
		return err
	}

	if _, err := file.WriteAt(truncatedContent, 0); err != nil {
		return err
	}

	return nil
}

func GetLastLineDataOfFile(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lastLineData []byte

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lastLineData = scanner.Bytes()
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lastLineData, nil
}

func GetAllFilesInDirectory(root string, string_arr []string) {

	files, err := ioutil.ReadDir(root)
	if err != nil {
		log.Fatal(1)
	}
	for _, file := range files {
		if file.IsDir() {
			wg.Add(1)
			go GetAllFilesInDirectory(filepath.Join(root, file.Name()), string_arr)
		} else {
			lock.Lock()
			string_arr = append(string_arr, filepath.Join(root, file.Name()))
			lock.Unlock()
		}
	}
	wg.Done()
}
