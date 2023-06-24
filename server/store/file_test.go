package store

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestFileService(t *testing.T) {

}

func TestWriteFileWithDirectories(t *testing.T) {

	fs := NewFileService()

	fps := "../../data/test/testing_Data.txt"
	text := "This is it"
	rootDir, _ := filepath.Abs(filepath.Dir(fps))
	got := fs.WriteFileWithDirectories(rootDir, []byte(text))
	var want error
	if got != want {
		t.Errorf("Wanted %v but got %v", want, got)
	}
	// check if the content is correct in the file.
	file, err := os.Open(rootDir)
	fmt.Println(err)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var l string
	for scanner.Scan() {
		l = scanner.Text()
	}
	if l != text {
		t.Errorf("Wanted text %v, but got %v", text, l)
	}
	os.Remove(rootDir)
}

func TestReadLatestFromFile(t *testing.T) {

	t.Run("Reading Last Entry in the file", func(t *testing.T) {
		fs := NewFileService()
		fps := "../../data/test/testing_Data.txt"
		txt := "This is the test"
		rootDir, _ := filepath.Abs(filepath.Dir(fps))
		for i := 0; i < 10; i++ {

			fs.WriteFileWithDirectories(rootDir, []byte(txt+fmt.Sprint(i)))
		}

		want := txt + "9"
		got, err := fs.ReadLatestFromFile(rootDir)
		if err != nil {
			t.Errorf("Did not expect an error but got %v", err)
		} else {
			if want != got {
				t.Errorf("Expected %v, but got %v", want, got)
			}
		}
		os.Remove(rootDir)
	})
	t.Run("Check for a file with no data", func(t *testing.T) {
		fs := NewFileService()
		fps := "../../data/test/testing_Data1.txt"
		rootDir, _ := filepath.Abs(filepath.Dir(fps))

		_, got := fs.ReadLatestFromFile(rootDir)
		// check if the type is correct
		if _, ok := got.(*os.PathError); !ok {
			t.Errorf("expected error of type os.PathError but got %v", got)
		}

	})
	// how to test for scanner.Err() ??
}

func TestReadLatestFromFileInBytes(t *testing.T) {

	t.Run("Reading Last Entry in the file - output in bytes", func(t *testing.T) {
		fs := NewFileService()
		fps := "../../data/test/testing_Data.txt"
		txt := "This is the test"
		rootDir, _ := filepath.Abs(filepath.Dir(fps))
		for i := 0; i < 10; i++ {

			fs.WriteFileWithDirectories(rootDir, []byte(txt+fmt.Sprint(i)))
		}

		want := []byte(txt + "9")
		got, err := fs.ReadLatestFromFileInBytes(rootDir)
		if err != nil {
			t.Errorf("Did not expect an error but got %v", err)
		} else {
			if string(want) != string(got) {
				t.Errorf("Expected %v, but got %v", string(want), string(got))
			}
		}
		os.Remove(rootDir)
	})
	t.Run("Check for a file with no data", func(t *testing.T) {
		fs := NewFileService()
		fps := "../../data/test/testing_Data1.txt"
		rootDir, _ := filepath.Abs(filepath.Dir(fps))

		_, got := fs.ReadLatestFromFileInBytes(rootDir)
		// check if the type is correct
		if _, ok := got.(*os.PathError); !ok {
			t.Errorf("expected error of type os.PathError but got %v", got)
		}

	})
	// how to test for scanner.Err() ??
}

func TestGetAllFilesInDirectory(t *testing.T) {
	// Test Data
	a := "../../data/mkdirTesting"
	b := "/test"
	c := "/testing_Data.txt"
	txt := "Random Data"
	num_of_files := 20
	fsMap := map[string]bool{}
	fs := NewFileService()
	for i := 0; i < num_of_files; i++ {
		f1 := a + "/" + fmt.Sprint(i) + b + c
		f2 := a + b + "/" + fmt.Sprint(i) + c
		fs.WriteFileWithDirectories(f1, []byte(txt))
		fs.WriteFileWithDirectories(f2, []byte(txt))
		fsMap[f1] = true
		fsMap[f2] = true

	}
	var string_arr []string
	ws := sync.WaitGroup{}
	ws.Add(1)
	go func() {
		defer ws.Done()
		fs.GetAllFilesInDirectory("../../data", &string_arr)
	}()
	ws.Wait()
	if len(string_arr) != 2*num_of_files {
		t.Errorf("Expected %v files but got %v", num_of_files, len(string_arr))
	}
	for i := range string_arr {
		if !fsMap[string_arr[i]] {
			t.Errorf("Expected %v but not found", string_arr[i])
		}
	}

}

// GetLatestFile doesnot require  benchmarking
func TestGetLatestFile(t *testing.T) {

	t.Run("Testing with No Sub-Directories", func(t *testing.T) {
		// Test Data
		a := "../../data/mkdirTesting"

		c := "/testing_Data.txt"
		txt := "Random Data"
		num_of_files := 20

		fs := NewFileService()
		for i := 0; i < num_of_files; i++ {

			f2 := a + c + fmt.Sprint(i)
			// note : we tried here to compare time of modification, but it almost the same.
			// for all the files. Although We dont expect the same.
			// each wal file wont be created at the same time .
			// so we will put a sleep timer between writing files to directories

			time.Sleep(10 * time.Millisecond)
			fs.WriteFileWithDirectories(f2, []byte(txt))
		}
		want := a + c + fmt.Sprint(19)

		got, err := fs.GetLatestFile(a)
		if err != nil {
			t.Error("Expected no error but got one ::::", err)
		}
		if want != got {
			t.Errorf("Wanted %v but got %v", want, got)
		}
		os.RemoveAll("../../data/mkdirTesting")
	})

	t.Run("Testing with Sub-Directories", func(t *testing.T) {
		// Test Data
		a := "../../data/mkdirTesting"
		b := "/testing"
		c := "/testing_Data.txt"
		txt := "Random Data"
		num_of_files := 20

		fs := NewFileService()
		for i := 0; i < num_of_files; i++ {

			f2 := a + b + fmt.Sprint(i) + c
			// note : we tried here to compare time of modification, but it almost the same.
			// for all the files. Although We dont expect the same.
			// each wal file wont be created at the same time .
			// so we will put a sleep timer between writing files to directories

			time.Sleep(10 * time.Millisecond)
			fs.WriteFileWithDirectories(f2, []byte(txt))
		}
		want := a + b + fmt.Sprint(19) + c

		got, err := fs.GetLatestFile(a)
		if err != nil {
			t.Error("Expected no error but got one ::::", err)
		}
		if want != got {
			t.Errorf("Wanted %v but got %v", want, got)
		}
	})
	os.RemoveAll("../../data/mkdirTesting")
}

// This should have benchmarking
func TestGetFileSize(t *testing.T) {

	t.Run("Testing FileSize for File that is not present", func(t *testing.T) {

		a := "../../data/mkdirTesting"
		b := "/testing"
		c := "/testing_Data.txt"
		fs := NewFileService()
		num_of_lines := 10
		content := "This is the content . 5 times this content's bytes length should be the response"
		for i := 0; i < num_of_lines; i++ {
			fs.WriteFileWithDirectories(a+b+c, []byte(content))
		}
		// It will also include linebreaks hence , we are adding num_of_lines as well
		want := int64(len([]byte(content))*6 + num_of_lines)
		got, err := fs.GetFileSize(a + b + c)

		if err != nil {
			t.Errorf("Did not expect an error here but got ")
		}
		if want != got {
			t.Errorf("Wanted %v but got %v", want, got)
		}
		os.RemoveAll(a)
	})

	t.Run("Testing FileSize for File is present", func(t *testing.T) {

		filepath := "../data/mkdirTesting/NotPresent"
		fs := NewFileService()
		_, err := fs.GetFileSize(filepath)
		t.Log(err)
		if err == nil {
			t.Error("Expected error but got nil ", err)
		}

	})

}

func BenchmarkGetFileSize(b *testing.B) {

	// we need to recreate the scenario where multiple Gets are happening for a single file
	// In Integration test we will benchmark the scenario for read + write in WAL file .
	// So multiple files will be gettnig the size of wal and then writing to the wal.
	// we also know that it wont' be like the situation where we need concurrent access to
	// read file size for different files as one file will be active at a time.
	b.Run("Multi Access to the same file", func(b *testing.B) {
		a := "../../data/mkdirTesting"

		c := "/testing_Data.txt"
		fs := NewFileService()
		num_of_lines := 10
		content := "This is the content . 5 times this content's bytes length should be the response"
		for i := 0; i < num_of_lines; i++ {
			fs.WriteFileWithDirectories(a+c, []byte(content))
		}

		want := int64(len([]byte(content))*num_of_lines + num_of_lines)
		for i := 0; i < b.N; i++ {
			got, err := fs.GetFileSize(a + c)
			if err != nil {
				b.Errorf("Did not expect an error here but got ")
			}
			if want != got {
				b.Errorf("Wanted %v but got %v", want, got)
			}
		}
		os.RemoveAll(a)

	})

}
