package store

import (
	"strings"

	"github.com/google/btree"
)

type Key string

func (kv Key) Less(than btree.Item) bool {
	if kv < than.(Key) {
		return true
	} else if kv < than.(Key) {
		return false
	}
	return false

}

type BTree interface {
	addKeyString(key string) bool
	getKeysFromPrefix(prefix string) []string
}

type GoogleBTree struct {
	tree *btree.BTree
}

func newBTree(degree int) *GoogleBTree {
	t := btree.New(degree)
	return &GoogleBTree{t}
}

func (bt *GoogleBTree) addKeyString(key string) bool {
	bt.tree.ReplaceOrInsert(Key(key))
	return true
}

func (bt *GoogleBTree) getKeysFromPrefix(prefix string) []string {
	matching_keys := []string{}
	bt.tree.AscendGreaterOrEqual(Key(prefix), func(item btree.Item) bool {
		if strings.Contains(string(item.(Key)), prefix) {
			matching_keys = append(matching_keys, string(item.(Key)))
			return true
		}
		return false
	})
	return matching_keys
}
