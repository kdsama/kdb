package store

import (
	"encoding/json"
)

type WalEntry struct {
	Node      *Node  `json:"node"`
	Operation string `json:"operation"`
	TxnID     string `json:"txnID"`
}

func NewWalEntry(node *Node, operation string, txnID string) *WalEntry {

	return &WalEntry{node, operation, txnID}
}

// serializing operations.
func (we *WalEntry) serialize() ([]byte, error) {

	b := Node{we.Node.Key, we.Node.Value, we.Node.Version, nil, we.Node.Deleted, we.Node.Timestamp, we.Node.Commit}
	a := WalEntry{&b, we.Operation, we.TxnID}
	arr, err := json.Marshal(a)
	if err != nil {
		return []byte{}, err
	}
	return arr, nil
}

// static method for WalEntry
// dont need to initiate the object to have access to it. But still placed in the same struct file
func deserialize(data []byte) (*WalEntry, error) {
	var obj WalEntry
	if err := json.Unmarshal(data, &obj); err != nil {
		return &WalEntry{}, err

	}
	return &obj, nil
}
