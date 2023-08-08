package store

import (
	"strings"

	"github.com/google/btree"
	"github.com/kdsama/kdb/logger"
)

type Key string

func (kv Key) Less(than btree.Item) bool {
	return kv < than.(Key)
}

type BTree interface {
	addKeyString(key string) bool
	getKeysFromPrefix(prefix string) []string
}

type GoogleBTree struct {
	tree   *btree.BTree
	logger *logger.Logger
}

func newBTree(degree int, lg *logger.Logger) *GoogleBTree {
	t := btree.New(degree)
	return &GoogleBTree{t, lg}
}

func (bt *GoogleBTree) addKeyString(key string) bool {
	// bt.logger.WARNf(key)
	bt.tree.ReplaceOrInsert(Key(key))
	return true
}

func (bt *GoogleBTree) getKeysFromPrefix(prefix string) []string {
	matching_keys := []string{}
	bt.tree.AscendGreaterOrEqual(Key(prefix), func(item btree.Item) bool {
		if strings.HasPrefix(string(item.(Key)), prefix) {
			matching_keys = append(matching_keys, string(item.(Key)))
			return true
		}
		return false
	})
	return matching_keys
}

func (bt *GoogleBTree) getKeyString(key string) (string, bool) {
	val := bt.tree.Get(Key(key))
	if val != nil {
		return string(val.(Key)), true
	}
	return "", false
}
