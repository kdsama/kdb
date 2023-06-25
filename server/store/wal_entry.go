package store

import (
	"encoding/json"
)

type WalEntry struct {
	node      *Node
	operation string
	txnID     string
}

func NewWalEntry(node *Node, operation string, txnID string) *WalEntry {

	return &WalEntry{node, operation, txnID}
}

func (we *WalEntry) serialize() ([]byte, error) {
	type JsonNode struct {
		Key       string `json:"key"`
		Value     string `json:"value"`
		Version   int8   `json:"version"`
		Deleted   bool   `json:"deleted"`
		Timestamp int64  `json:"timestamp"`
	}
	type JsonWAL struct {
		Node      JsonNode `json:"node"`
		Operation string   `json:"operation"`
		TxnID     string   `json:"txnID"`
	}
	b := JsonNode{we.node.key, we.node.value, we.node.version, we.node.deleted, we.node.timestamp}
	a := JsonWAL{b, we.operation, we.txnID}
	arr, err := json.Marshal(a)
	if err != nil {
		return []byte{}, err
	}
	return arr, nil
}
