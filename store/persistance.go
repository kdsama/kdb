package store

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const (
	COMPACTIONSIZE = 5

	NUM_THREADS = 2
)

var (
	err_NoDataFound = errors.New("No data present, make sure the file-directory is correct")
)
var (
	persistanceRequestsTotal = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ps_file_requests",
		Help:    "file requests :: persistance layer",
		Buckets: []float64{0.0, 20.0, 40.0, 60.0, 80.0, 100.0, 160.0, 180.0, 200.0, 400.0, 800.0, 1600.0},
	}, []string{"reqtype"})
)

type Persistance struct {
	fs     *fileService
	prefix string
	mut    *sync.Mutex
	wg     sync.WaitGroup
	nodes  []Node
	logger *zap.SugaredLogger
}

func NewPersistance(prefix string, lg *zap.SugaredLogger) *Persistance {
	fs := NewFileService()

	return &Persistance{fs: fs,
		prefix: prefix,
		mut:    &sync.Mutex{},
		wg:     sync.WaitGroup{},
		nodes:  []Node{},
		logger: lg}
}

// we need to encode the data in a way that we can save it
// one way to do it is to use json marshaller . But there will be too many steps in this case.
// I need to find a serialisation technique here .
// I also need to do it in a way that I am not reputting the same data of prev_versions again . remove prev_versions and then serialise the data

func (p *Persistance) Add(node Node) error {

	// need to make sure that the node is persistanceready
	buffer, node := serializeNode(node)

	return p.Save(node.Key, &buffer)
}

func serializeNode(node Node) ([]byte, Node) {
	bytes, err := json.Marshal(node)
	if err != nil {
		log.Fatal("perpepreppreprp", err)
	}
	return bytes, node
}

// There is no need for update . Update means we are saving a new byte array to the file
func (p *Persistance) Save(key string, buffer *[]byte) error {
	t := time.Now()
	p.mut.Lock()

	err := p.fs.WriteFileLnWithDirectories(p.prefix+key, *buffer)
	p.mut.Unlock()

	persistanceRequestsTotal.WithLabelValues("Save File (ms)").Observe(float64(time.Since(t).Milliseconds()))
	return err
}

func (p *Persistance) GetNodeFromKey(key string) (Node, error) {
	dir := p.prefix + key
	return p.GetNodeFromAbsolutePath(dir)

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
	t := time.Now()
	for file_dir := range buffered_channel {
		n, err := p.GetNodeFromAbsolutePath(file_dir)
		if err != nil {
			// p.logger.Fatalf(err.Error())
			p.logger.Infof("For now we will say there is no error but terhe is %s", err.Error())
		} else {
			p.mut.Lock()
			p.nodes = append(p.nodes, n)
			p.mut.Unlock()
		}
	}

	p.wg.Done()
	persistanceRequestsTotal.WithLabelValues("Get Many").Observe(float64(time.Since(t).Milliseconds()))
}
func (p *Persistance) GetNodeFromAbsolutePath(dir string) (Node, error) {

	t := time.Now()
	var n Node
	node_in_bytes, err := p.fs.ReadLatestFromFileInBytes(dir)
	if err != nil {

		return Node{}, err

	}
	if err := json.Unmarshal(node_in_bytes, &n); err != nil {
		return Node{}, err
	}
	persistanceRequestsTotal.WithLabelValues("Get File (ms)").Observe(float64(time.Since(t).Milliseconds()))
	return n, nil
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
		go p.GetNodesInParallel(buffered_channel)
	}
	p.wg.Add(NUM_THREADS)
	for i := range node_filedirs {
		buffered_channel <- node_filedirs[i]
	}
	close(buffered_channel)
	p.wg.Wait()

}

func init() {
	prometheus.MustRegister(persistanceRequestsTotal)
}
