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
	Version          uint32 `json:"version"`
	PreviousVersions []Node `json:"-"`
	Deleted          bool   `json:"deleted"`
	Timestamp        int64  `json:"timestamp"`
	CommitTimestamp  int64  `json:"commitTimestamp"`
}

func NewNode(key string, value string) *Node {
	t := time.Now().UnixMilli()
	return &Node{Key: key,
		Value:            value,
		Version:          0,
		PreviousVersions: nil,
		Deleted:          false,
		Timestamp:        t,
		// Commit:           commitLevel(Waiting)}
	}

}
func (n *Node) Update(value string) Node {
	// get current information and put it to prev version
	n.PreviousVersions = append(n.PreviousVersions, Node{n.Key, n.Value, n.Version, nil, n.Deleted, n.Timestamp, 0})
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

	return_node := Node{n.Key, n.Value, n.Version, nil, n.Deleted, n.Timestamp, 0}
	return return_node
}
