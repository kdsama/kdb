package store

import (
	"reflect"
	"testing"
)

func TestWalSerializeAndDeserialize(t *testing.T) {
	t.Run("Testing Serialization and Deserialization of WAL Entry", func(t *testing.T) {
		a := NewNode("test_case", "value")
		w := NewWalEntry(a, "operation", "transactionID")
		data, err := w.serialize()
		if err != nil {
			t.Error("Did not expect an error here but got ", err)
		}
		// deserialize data and compare two struct objects
		new_w, err := deserialize(data)
		if err != nil {
			t.Error("Did not expect an error here but got", err)
		}
		if !reflect.DeepEqual(w, new_w) {
			t.Errorf("Expected the two values to be equal but the are not. \n want %v but got  \n %v", w, new_w)

		}

	})

}
