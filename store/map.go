package store

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	err_NodeNotFound      = errors.New("node not found")
	err_NoneNodeFound     = errors.New("none of the nodes are present")
	err_SomeNodesNotFound = errors.New("some nodes not found")
	err_Upserted          = errors.New("the node was already present in hashmap, so the information was added instead of being updated ")
)

var (
	requestsTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ps_map_size",
		Help: "Shares the size of inmemory map in kb",
	})
)

type keyValue map[string]*Node

type HashMap struct {
	kv     keyValue
	mut    *sync.RWMutex
	logger *zap.SugaredLogger
}

func NewHashMap(lg *zap.SugaredLogger) *HashMap {
	kv := map[string]*Node{}
	mut := sync.RWMutex{}

	h := &HashMap{
		kv:     kv,
		mut:    &mut,
		logger: lg}
	go h.mapSize()
	return h
}
func (hm *HashMap) Add(key string, value string) *Node {
	node := NewNode(key, value)

	hm.mut.Lock()
	hm.kv[key] = node
	hm.mut.Unlock()
	return node
}

func (hm *HashMap) AddNode(node *Node) (*Node, error) {

	hm.mut.Lock()
	_, ok := hm.kv[node.Key]
	hm.kv[node.Key] = node
	hm.kv[node.Key].CommitNode()
	hm.mut.Unlock()
	if ok {
		return node, err_Upserted
	}
	return node, nil
}

func (hm *HashMap) Get(key string) (*Node, error) {
	hm.mut.RLock()
	n, ok := hm.kv[key]
	hm.mut.RUnlock()
	if !ok || n.Deleted {
		return &Node{}, err_NodeNotFound
	}
	return n, nil
}

func (hm *HashMap) GetSeveral(keys []string) ([]Node, []string, error) {

	hm.mut.RLock()
	missing_keys := []string{}
	node_list := []Node{}
	for i := range keys {
		n, ok := hm.kv[keys[i]]
		if !ok {
			missing_keys = append(missing_keys, keys[i])
		} else {
			node_list = append(node_list, *n)
		}
	}
	hm.mut.RUnlock()
	if len(missing_keys) == len(keys) {
		return []Node{}, missing_keys, err_NoneNodeFound
	} else if len(missing_keys) == 0 {
		return node_list, missing_keys, nil
	}
	return node_list, missing_keys, err_SomeNodesNotFound

}

func (hm *HashMap) Update(key string, value string) (bool, error) {

	hm.mut.RLock()
	n, ok := hm.kv[key]
	hm.mut.RUnlock()
	if !ok {
		atomic.AddUint32(&n.Version, uint32(1))
		return false, err_NodeNotFound
	}
	_ = n.Update(value)

	return true, nil
}

func (hm *HashMap) Delete(key string) (*Node, error) {

	hm.mut.RLock()
	n, ok := hm.kv[key]
	hm.mut.RUnlock()
	if !ok {
		return nil, err_NodeNotFound
	}
	_ = n.Delete()

	return n, nil
}

func (hm *HashMap) Commit(key string, version int) error {
	hm.kv[key].Version++
	fmt.Println("Committed value : ", hm.kv[key])

	return nil

}

func (hm *HashMap) mapSize() {
	for {
		// Simulate temperature reading
		// temperature := generateRandomTemperature()
		// Update the Gauge value with the current temperature
		requestsTotal.Set(float64(unsafe.Sizeof(hm.kv)))

		time.Sleep(5 * time.Second)
	}
}

func init() {
	prometheus.MustRegister(requestsTotal)
}
