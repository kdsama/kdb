package store

import (
	"os"
	"testing"
	"time"
)

var wal_prefix = "pref"
var wal_test_directory = "../../data/testWAL/"

func testwalcleanup(w *WAL) {
	w.flushall()
	os.RemoveAll(wal_prefix)

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

	// as 5000 seconds , then buffer will not be loaded to the file wal_test_directory yet

}
