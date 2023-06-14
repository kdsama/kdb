package store

import (
	"errors"
	"sync"
)

var (
	err_NodeNotFound = errors.New("node not found")
)

type keyValue map[string]*Node

type HashMap struct {
	kv  keyValue
	mut sync.RWMutex
}

func (hm *HashMap) Add(key string, value string) bool {
	node := NewNode(key, value)

	hm.mut.Lock()
	hm.kv[key] = node
	hm.mut.Unlock()
	return true
}

func (hm *HashMap) Get(key string) (Node, error) {
	hm.mut.RLock()
	n, ok := hm.kv[key]
	hm.mut.RUnlock()
	if !ok {
		return Node{}, err_NodeNotFound
	}
	return *n, nil
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

func (hm *HashMap) Delete(key string) (bool, error) {

	hm.mut.RLock()
	n, ok := hm.kv[key]
	hm.mut.RUnlock()
	if !ok {
		return false, err_NodeNotFound
	}
	_ = n.Delete()

	return true, nil
}
