package store

import (
	"bytes"
	"encoding/gob"
	"sync"
)

type WAL struct {
	mut sync.Mutex
}

const (
	COMPACTIONSIZE = 5
	PREFIX         = "/data"
)

// we need to encode the data in a way that we can save it
// one way to do it is to use json marshaller . But there will be too many steps in this case.
// I need to find a serialisation technique here .
// I also need to do it in a way that I am not reputting the same data of prev_versions again . remove prev_versions and then serialise the data

func (w *WAL) Add(node Node) error {

	buffer := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buffer)
	// need to make sure that the node is persistanceready
	node = node.persistanceReady()
	enc.Encode(node)
	return w.Save(node.key, buffer)
}

// There is no need for update . Update means we are saving a new byte array to the file
func (w *WAL) Save(key string, buffer *bytes.Buffer) error {
	w.mut.Lock()
	err := WriteFileWithDirectories(key, buffer.Bytes())
	w.mut.Unlock()
	return err
}

func (w *WAL) Get(key string) (Node, error) {
	dir := PREFIX + key

	var n Node
	node_in_bytes, err := GetLastLineDataOfFile(dir)
	if err != nil {

		return Node{}, err

	}
	buffer := bytes.NewBuffer(node_in_bytes)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(&n)
	if err != nil {
		return Node{}, err
	}

	return n, nil

}

func (w *WAL) GetALL() ([]Node, error) {

}
