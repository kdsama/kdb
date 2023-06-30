package store

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

var wal_prefix = "pref"
var wal_test_directory = "../../data/testWAL/"

func testwalcleanup(w *WAL) {
	w.flushall()
	os.RemoveAll(wal_test_directory)

}
func TestAddEntry(t *testing.T) {

	t.Run("Check the value in buffer array", func(t *testing.T) {

		w := NewWAL(wal_prefix, wal_test_directory, fileService{}, 5000)

		key := "Key"
		value := "Value"
		node := NewNode(key, value)
		w.addEntry(*node, "ADD")

		wanEntry, err := deserialize(wal_buffer[:])
		if err != nil {
			t.Error("Did not expect an error here but got ", err)
		}
		if wanEntry.Node.Key != key {
			t.Errorf("Expected key %v but got %v", key, wanEntry.Node.Key)
		}
		if wanEntry.Node.Value != value {
			t.Errorf("Expected key %v but got %v", value, wanEntry.Node.Value)
		}
		// cleanup
		testwalcleanup(w)
	})
	t.Run("Check size of buffer after timeout and not insertions", func(t *testing.T) {

		w := NewWAL(wal_prefix, wal_test_directory, fileService{}, 1)

		key := "Key"
		value := "Value"
		node := NewNode(key, value)
		w.addEntry(*node, "ADD")
		time.Sleep(2 * time.Second)
		want := 0
		got := w.checkSize()
		if want != got {
			t.Errorf("wanted %v but got %v", want, got)
		}
		testwalcleanup(w)
	})
	// the test below help us realized that we need to make equal sized files for partitioning
	// what we need to do is
	// keep on appending to an existing wal file until the limit is reached
	// once the limit is reached, create a new wal file
	// the size of buffer may or maynot be equal to the wal file
	// for our case as we want to continuously write to the wal file, we will keep them different

}
func BenchmarkAddEntry(b *testing.B) {
	// os.RemoveAll(wal_test_directory)
	b.Run("Test the same for multiple file creations . The file names should atomically increase", func(b *testing.B) {

		fs := fileService{}
		w := NewWAL(wal_prefix, wal_test_directory, fs, 1)
		key := "Key"
		value := "{\"id\":1,\"n\":\"John Doe\",\"a\":30,\"e\":\"johndoejohndoejohndoejohndoejohndoejohndoejohndoe1@example.com\"}"
		for i := 0; i < 100000; i++ {
			node := NewNode(key, fmt.Sprint(i)+value)
			w.addEntry(*node, "ADD")

		}

		time.Sleep(10 * time.Second)
		ws := sync.WaitGroup{}
		files := []string{}

		ws.Add(1)
		go func() {
			defer ws.Done()
			fs.GetAllFilesInDirectory(wal_test_directory, &files)
		}()
		ws.Wait()

	})

}
