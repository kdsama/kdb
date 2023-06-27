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

func (we *WalEntry) serialize() ([]byte, error) {

	b := Node{we.Node.Key, we.Node.Value, we.Node.Version, nil, we.Node.Deleted, we.Node.Timestamp}
	a := WalEntry{&b, we.Operation, we.TxnID}
	arr, err := json.Marshal(a)
	if err != nil {
		return []byte{}, err
	}
	return arr, nil
}

// func (we *WalEntry) deserialize(data []byte) (*Node, error) {

// }
