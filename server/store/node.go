package store

import (
	"errors"
	"time"
)

var (
	err_AlreadyCommited = errors.New("the iteration is already committed")
	err_OldVersion      = errors.New("the version you are trying to commit is old")
)

// compact order of the fields will lead to smaller size
type Node struct {
	Key              string `json:"key"`
	Value            string `json:"value"`
	Version          int8   `json:"version"`
	PreviousVersions []Node
	Deleted          bool  `json:"deleted"`
	Timestamp        int64 `json:"timestamp"`
	Commit           int8  `json:"commit"`
}

func NewNode(key string, value string) *Node {
	t := time.Now().Unix()
	return &Node{key, value, 0, nil, false, t, int8(WAITING)}
}

func (n *Node) CommitNode() {
	// get current information and put it to prev version
	n.Commit = COMMITTED
}
func (n *Node) Abort() {
	n.Commit = ABORTED
}

func (n *Node) Update(value string) Node {
	// get current information and put it to prev version
	n.PreviousVersions = append(n.PreviousVersions, Node{n.Key, n.Value, n.Version, nil, n.Deleted, n.Timestamp, n.Commit})
	n.Value = value
	n.Version += 1
	return *n
}
func (n *Node) Delete() bool {
	// get current information and put it to prev version
	// n.PreviousVersions = append(n.PreviousVersions, Node{n.Key, n.Value, n.Version, nil, n.Deleted, n.Timestamp})
	n.Deleted = true

	return true
}

func (n *Node) persistanceReady() Node {

	return_node := Node{n.Key, n.Value, n.Version, nil, n.Deleted, n.Timestamp, n.Commit}
	return return_node
}
