package store

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"sync"
)

type Persistance struct {
	fs     *fileService
	prefix string
	mut    sync.Mutex
	wg     sync.WaitGroup
	nodes  []Node
}

var (
	err_NoDataFound = errors.New("No data present, make sure the file-directory is correct")
)

const (
	COMPACTIONSIZE = 5

	NUM_THREADS = 2
)

func NewPersistance(prefix string) *Persistance {
	fs := NewFileService()

	return &Persistance{fs, prefix, sync.Mutex{}, sync.WaitGroup{}, []Node{}}
}

// we need to encode the data in a way that we can save it
// one way to do it is to use json marshaller . But there will be too many steps in this case.
// I need to find a serialisation technique here .
// I also need to do it in a way that I am not reputting the same data of prev_versions again . remove prev_versions and then serialise the data

func (p *Persistance) Add(node Node) error {

	buffer := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(buffer)
	// need to make sure that the node is persistanceready
	node = node.persistanceReady()
	enc.Encode(node)
	return p.Save(node.key, buffer)
}

// There is no need for update . Update means we are saving a new byte array to the file
func (p *Persistance) Save(key string, buffer *bytes.Buffer) error {
	p.mut.Lock()
	err := p.fs.WriteFileWithDirectories(p.prefix+key, buffer.Bytes())
	p.mut.Unlock()
	return err
}

func (p *Persistance) GetNodeFromKey(key string) (Node, error) {
	dir := p.prefix + key
	return p.GetNode(dir)

}

func (p *Persistance) GetNode(dir string) (Node, error) {
	var n Node
	node_in_bytes, err := p.fs.ReadLatestFromFileInBytes(p.prefix + dir)
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

func (p *Persistance) GetNodesInParallel(buffered_channel chan string) {
	for file_dir := range buffered_channel {
		n, err := p.GetNode(file_dir)
		if err != nil {
			log.Println(err)
		} else {
			p.mut.Lock()
			p.nodes = append(p.nodes, n)
			p.mut.Unlock()
		}
	}
}

func (p *Persistance) GetALLNodes() {

	node_filedirs := []string{}
	p.fs.GetAllFilesInDirectory(p.prefix, &node_filedirs)
	if len(node_filedirs) == 0 {
		log.Println("No data found")
		// panic here ?
	}
	// now below different files are going to be opened
	// should I do everything concurrently here ? or sequentially
	// this is basically something thats going to run at the start .
	// now is this where we can use channels ? Ummm lets think
	// we can use unbuffered channels
	// but we will use buffered channels here .
	// Keep the buffer size 100 . Overflow means it will have to wait until a previous request is done . I am fine with that

	buffered_channel := make(chan string, 100)

	for i := 0; i < NUM_THREADS; i++ {
		p.GetNodesInParallel(buffered_channel)
	}
	p.wg.Add(NUM_THREADS)
	for i := range node_filedirs {
		buffered_channel <- node_filedirs[i]
	}
	close(buffered_channel)
	p.wg.Wait()

}
