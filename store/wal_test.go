package store

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

var wal_prefix = "node"
var wal_test_directory = "../../data/testWAL/"

func testwalcleanup() {

	os.RemoveAll(wal_test_directory)

}

// DeprecatedTest. Now we are not saving uncommitted transaction
// func TestAddEntry(t *testing.T) {

// 	t.Run("Check the value in buffer array", func(t *testing.T) {

// 		w := NewWAL(wal_prefix, wal_test_directory, fileService{}, 5000, lg)
// 		w.flushall()
// 		key := "Key"
// 		value := "Value"
// 		node := NewNode(key, value)
// 		w.addEntry(*node, "ADD")

// 		wanEntry, err := deserialize(wal_buffer[:])
// 		if err != nil {
// 			t.Error("Did not expect an error here but got ", err)
// 		}
// 		if wanEntry.Node.Key != key {
// 			t.Errorf("Expected key %v but got %v", key, wanEntry.Node.Key)
// 		}
// 		if wanEntry.Node.Value != value {
// 			t.Errorf("Expected key %v but got %v", value, wanEntry.Node.Value)
// 		}
// 		// cleanup
// 		testwalcleanup()
// 	})
// 	t.Run("Check size of buffer after timeout and not insertions", func(t *testing.T) {

// 		w := NewWAL(wal_prefix, wal_test_directory, fileService{}, 1, lg)

// 		key := "Key"
// 		value := "Value"
// 		node := NewNode(key, value)
// 		w.addEntry(*node, "ADD")
// 		time.Sleep(2 * time.Second)
// 		want := 0
// 		got := w.checkSize()
// 		if want != got {
// 			t.Errorf("wanted %v but got %v", want, got)
// 		}
// 		testwalcleanup()
// 	})
// 	// the test below help us realized that we need to make equal sized files for partitioning
// 	// what we need to do is
// 	// keep on appending to an existing wal file until the limit is reached
// 	// once the limit is reached, create a new wal file
// 	// the size of buffer may or maynot be equal to the wal file
// 	// for our case as we want to continuously write to the wal file, we will keep them different

// }
// func BenchmarkAddEntry(b *testing.B) {
// 	// os.RemoveAll(wal_test_directory)
// 	b.Run("Test the same for multiple file creations . The file names should atomically increase", func(b *testing.B) {

// 		fs := fileService{}
// 		w := NewWAL(wal_prefix, wal_test_directory, fs, 1, lg)
// 		key := "Key"
// 		value := "{\"id\":1,\"n\":\"John Doe\",\"a\":30,\"e\":\"johndoejohndoejohndoejohndoejohndoejohndoejohndoe1@example.com\"}"
// 		for i := 0; i < 100000; i++ {
// 			node := NewNode(key, fmt.Sprint(i)+value)
// 			w.addEntry(*node, "ADD")

// 		}

// 		time.Sleep(10 * time.Second)
// 		ws := sync.WaitGroup{}
// 		files := []string{}

// 		ws.Add(1)
// 		go func() {
// 			defer ws.Done()
// 			fs.GetAllFilesInDirectory(wal_test_directory, &files)
// 		}()
// 		ws.Wait()
// 		testwalcleanup()
// 	})

// }

func TestSetCounterFromFileName(t *testing.T) {

	t.Run("Checking for counter, when loaded for the first time", func(t *testing.T) {
		w := NewWAL(wal_prefix, wal_test_directory, fileService{}, 1, lg)

		key := "Key"
		value := "Value"
		node := NewNode(key, value)
		we, _ := w.addEntry(*node, "ADD")
		r, _ := we.serialize()
		w.AddWALEntry(&r)
		time.Sleep(2 * time.Second)

		want := int64(0)
		got := w.file_counter
		if want != got {
			t.Errorf("Wanted %v, but got %v", want, got)
		}
		testwalcleanup()
	})

	t.Run("Checking for counter, when one  file exist", func(t *testing.T) {

		// this function made me realize that the logic of new wal file creation is not strong at all
		fs := fileService{}
		testwalcleanup()
		w := NewWAL(wal_prefix, wal_test_directory, fs, 1, lg)

		key := "Key"
		value := "{\"id\":1,\"n\":\"John Doe\",\"a\":30,\"e\":\"johndoejohndoejohndoejohndoejohndoejohndoejohndoe1@example.com\"}"
		for i := 0; i < MAX_FILE_SIZE/1000; i++ {
			node := NewNode(key, fmt.Sprint(i)+value)
			we, _ := w.addEntry(*node, "ADD")
			r, _ := we.serialize()
			w.AddWALEntry(&r)
			w.addEntry(*node, "ADD")

		}

		time.Sleep(5 * time.Second)
		ws := sync.WaitGroup{}
		files := []string{}

		ws.Add(1)
		go func() {
			defer ws.Done()
			fs.GetAllFilesInDirectory(wal_test_directory, &files)
		}()
		ws.Wait()
		want := 1
		got := len(files)
		if want != got {
			t.Errorf("Expected %v but got %v", got, want)
		}
		testwalcleanup()
	})

	// the test is running fine, just that the default time given to this test case is making it time out
	t.Run("Checking for counter, when 10  files exist", func(t *testing.T) {

		// this function made me realize that the logic of new wal file creation is not strong at all
		fs := fileService{}
		testwalcleanup()
		w := NewWAL(wal_prefix, wal_test_directory, fs, 1, lg)

		key := "Key"
		value := "{\"id\":1,\"n\":\"John Doe\",\"a\":30,\"e\":\"johndoejohndoejohndoejohndoejohndoejohndoejohndoe1@example.com\"}"

		for i := 0; i < 1200000; i++ {
			node := NewNode(key, fmt.Sprint(i)+value)
			we, _ := w.addEntry(*node, "ADD")
			r, _ := we.serialize()
			w.AddWALEntry(&r)

		}

		time.Sleep(5 * time.Second)
		files := []string{}

		fs.GetAllFilesInDirectory(wal_test_directory, &files)

		want := 3
		got := len(files)
		if want != got {
			t.Errorf("Expected %v but got %v", got, want)
		}
		testwalcleanup()
	})
}

func TestSetLatestCounter(t *testing.T) {
	os.RemoveAll(wal_test_directory)
	t.Run("On initial setup , counter should return as zero", func(t *testing.T) {
		fs := fileService{}
		w := NewWAL(wal_prefix, wal_test_directory, fs, 1, lg)

		// check counter in 1 second.
		want := 0
		got := w.counter
		if int64(want) != got {
			t.Errorf("wanted %v but got %v", want, got)
		}
		testwalcleanup()
	})

	// checkagain Test case ....
	t.Run("Counter on top of several entries single file", func(t *testing.T) {

		fs := fileService{}
		w := NewWAL(wal_prefix, wal_test_directory, fs, 1, lg)

		// check counter in 1 second.
		want := 100000
		key := "/Something"
		value := "cff "
		for i := 0; i < 100000; i++ {
			node := NewNode(key, fmt.Sprint(i)+value)
			we, _ := w.addEntry(*node, "ADD")
			r, _ := we.serialize()
			w.AddWALEntry(&r)
		}
		time.Sleep(1 * time.Second)
		got := w.counter
		if int64(want) != got {
			t.Errorf("wanted %v but got %v", want, got)
		}
		testwalcleanup()
	})
	t.Run("Counter on top of several entries, several files ", func(t *testing.T) {
		// return
		fs := fileService{}
		w := NewWAL(wal_prefix, wal_test_directory, fs, 1, lg)

		// check counter in 1 second.
		want := 1200000
		key := "/Something"
		value := "bff"
		for i := 0; i < 1200000; i++ {
			node := NewNode(key, fmt.Sprint(i)+value)
			we, _ := w.addEntry(*node, "ADD")
			r, _ := we.serialize()
			w.AddWALEntry(&r)
		}
		time.Sleep(2 * time.Second)
		got := w.counter
		if int64(want) != got {
			t.Errorf("wanted %v but got %v", want, got)
		}
		testwalcleanup()
	})

}
