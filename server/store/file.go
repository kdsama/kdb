package store

import (
	"bufio"
	"fmt"
	"io/fs"
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

type fileService struct {
}

func NewFileService() *fileService {
	return &fileService{}
}

// Write a file. Will create directory  if not present
func (fsv *fileService) WriteFileWithDirectories(fp string, data []byte) error {
	dir := filepath.Dir(fp)

	// Create directories recursively if they don't exist
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	// Write file using ioutil.WriteFile
	file, err := os.OpenFile(fp, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	lock.Lock()
	_, err = fmt.Fprintln(file, string(data))
	lock.Unlock()
	return err

}

// Returns Last line from the file.
func (fsv *fileService) ReadLatestFromFile(filepath string) (string, error) {
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

// secondary function . Dont need to test it now . Can work on it in future iterations
// remove the comment once test cases are completed
func (fsv *fileService) TruncateFile(filepath string, linesToRemove int) error {
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

func (fsv *fileService) ReadLatestFromFileInBytes(filepath string) ([]byte, error) {
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

func (fsv *fileService) GetAllFilesInDirectory(root string, string_arr *[]string) {
	wg := sync.WaitGroup{}
	files, err := ioutil.ReadDir(root)
	if err != nil {
		log.Fatal(1)
	}
	for _, file := range files {

		if file.IsDir() {
			wg.Add(1)

			go func(file fs.FileInfo) {
				defer wg.Done()

				fsv.GetAllFilesInDirectory(filepath.Join(root, file.Name()), string_arr)
			}(file)

		} else {
			lock.Lock()
			(*string_arr) = append((*string_arr), filepath.Join(root, file.Name()))
			lock.Unlock()
		}
	}

	wg.Wait()

}
