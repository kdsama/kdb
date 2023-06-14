package store

import "time"

type Node struct {
	key           string
	value         string
	version       int
	prev_versions []Node
	deleted       bool
	timestamp     int64
}

func NewNode(key string, value string) *Node {
	t := time.Now().Unix()
	return &Node{key, value, 0, nil, false, t}
}

func (n *Node) Update(value string) Node {
	// get current information and put it to prev version
	n.prev_versions = append(n.prev_versions, Node{n.key, n.value, n.version, nil, n.deleted, n.timestamp})
	n.value = value
	n.version += 1
	return *n
}
func (n *Node) Delete() bool {
	// get current information and put it to prev version
	// n.prev_versions = append(n.prev_versions, Node{n.key, n.value, n.version, nil, n.deleted, n.timestamp})
	n.deleted = true

	return true
}
