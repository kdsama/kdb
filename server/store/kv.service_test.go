package store

import (
	"fmt"
	"os"
	"testing"
)

func TestKVInit(t *testing.T) {

	// to test kv init
	// we probably should dump some data first
	// so lets persist some data first
	ps_prefix := "../../data/kvservice/persist/"
	os.RemoveAll(ps_prefix)
	ps := NewPersistance(ps_prefix)

	// Now I will save a few keys
	// and I should be able to get data of these keys
	for i := 0; i < 10; i++ {
		node := NewNode(fmt.Sprintf("key%v", i), "Some Value")
		ps.Add(*node)
	}
	wal_prefix := "node"
	dir := "../../data/kvservice/wal/"
	x := NewKVService(ps_prefix, wal_prefix, dir, 1, 10)
	x.Init()
	got := x.btree.getKeysFromPrefix("key")
	want := 10
	if len(got) != want {
		t.Errorf("Wanted the length to be %v, but got %v", want, len(got))
	}
	os.RemoveAll(ps_prefix)
	os.RemoveAll(dir)
}
