package store

import (
	"errors"
	"sync"
)

var (
	err_NodeNotFound      = errors.New("node not found")
	err_NoNodeFound       = errors.New("none of the nodes are present")
	err_SomeNodesNotFound = errors.New("some nodes not found")
)

var (
	COMMITTED = int8(1)
	WAITING   = int8(0)
	ABORTED   = int8(2)
)

type keyValue map[string]*Node

type HashMap struct {
	kv  keyValue
	mut sync.RWMutex
}

func NewHashMap() *HashMap {
	kv := map[string]*Node{}
	mut := sync.RWMutex{}
	return &HashMap{kv, mut}
}
func (hm *HashMap) Add(key string, value string) *Node {
	node := NewNode(key, value)

	hm.mut.Lock()
	hm.kv[key] = node
	hm.mut.Unlock()
	return node
}

func (hm *HashMap) AddNode(node *Node) error {
	n := *node

	hm.mut.Lock()
	hm.kv[n.Key] = node
	hm.mut.Unlock()
	return nil
}

func (hm *HashMap) Get(key string) (Node, error) {
	hm.mut.RLock()
	n, ok := hm.kv[key]
	hm.mut.RUnlock()
	if !ok || n.Deleted {
		return Node{}, err_NodeNotFound
	}
	return *n, nil
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
		return []Node{}, missing_keys, err_NoNodeFound
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
	if _, ok := hm.kv[key]; !ok {
		panic(err_NodeNotFound)
	}
	n := (*hm.kv[key])
	if n.Version > int8(version) {
		return err_OldVersion
	}

	if n.Commit != COMMITTED {
		return err_AlreadyCommited
	}
	n.CommitNode()
	return nil
}
