package store

import (
	"strings"
	"sync"
	"time"

	"github.com/google/btree"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var btreeLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "ps_btree_requests",
	Help:    "btree inserts :: btree layer",
	Buckets: []float64{0.0, 20.0, 40.0, 60.0, 80.0, 100.0, 160.0, 180.0, 200.0, 400.0, 800.0, 1600.0},
}, []string{"reqtype"})

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
	logger *zap.SugaredLogger
	mut    *sync.RWMutex
}

func newBTree(degree int, lg *zap.SugaredLogger) *GoogleBTree {
	t := btree.New(degree)
	return &GoogleBTree{
		tree:   t,
		logger: lg,
		mut:    &sync.RWMutex{},
	}
}

func (bt *GoogleBTree) addKeyString(key string) bool {
	t := time.Now()
	// concurrent mutations is not safe
	bt.mut.Lock()
	bt.tree.ReplaceOrInsert(Key(key))
	bt.mut.Unlock()
	btreeLatency.WithLabelValues("put (ms)").Observe(float64(time.Since(t).Milliseconds()))
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

func init() {
	prometheus.MustRegister(btreeLatency)
}
