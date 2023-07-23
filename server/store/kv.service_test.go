package store

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestKVInit(t *testing.T) {

	// to test kv init
	// we probably should dump some data first
	// so lets persist some data first
	// Now I will save a few keys
	// and I should be able to get data of these keys
	ps_prefix, wal_prefix, dir := PrepKV()
	x := NewKVService(ps_prefix, wal_prefix, dir, 1, 10)
	x.Init()
	got := x.btree.getKeysFromPrefix("key")
	want := 10
	if len(got) != want {
		t.Errorf("Wanted the length to be %v, but got %v", want, len(got))
	}
	// os.RemoveAll(ps_prefix)
	// os.RemoveAll(dir)
}

func TestAdd(t *testing.T) {
	ps_prefix, wal_prefix, dir := PrepKV()
	x := NewKVService(ps_prefix, wal_prefix, dir, 1, 10)
	x.Init()
	k, v := "newKey", "newValue"
	txnID, err := x.Add(k, v)
	if err != nil {
		t.Errorf("Expected nil but got %v", err)
	}
	time.Sleep(2 * time.Second)
	// confirm the transactionID

	file := x.wal.getCurrentFileName()
	data, err := x.ps.fs.ReadLatestFromFile(file)
	if err != nil {
		t.Fatal(err)
	}
	wal_entry, err := deserialize([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	want := wal_entry.TxnID
	got := txnID
	if want != got {
		t.Errorf("want %v , but got %v", want, got)
	}
	os.RemoveAll(dir)
	os.RemoveAll(ps_prefix)
}

func TestUpdate(t *testing.T) {
	ps_prefix, wal_prefix, dir := PrepKV()
	x := NewKVService(ps_prefix, wal_prefix, dir, 1, 10)
	x.Init()
	k, v := "newKey", "newValue"
	txnID, err := x.Update(k, v)
	if err != nil {
		t.Errorf("Expected nil but got %v", err)
	}
	time.Sleep(2 * time.Second)
	// confirm the transactionID

	file := x.wal.getCurrentFileName()
	data, err := x.ps.fs.ReadLatestFromFile(file)
	if err != nil {
		t.Fatal(err)
	}
	wal_entry, err := deserialize([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	want := wal_entry.TxnID
	got := txnID
	if want != got {
		t.Errorf("want %v , but got %v", want, got)
	}
	os.RemoveAll(dir)
	os.RemoveAll(ps_prefix)
}

func TestDelete(t *testing.T) {
	ps_prefix, wal_prefix, dir := PrepKV()
	x := NewKVService(ps_prefix, wal_prefix, dir, 1, 10)
	x.Init()
	k, v := "newKey", "newValue"
	x.Add(k, v)
	txnID, err := x.Delete(k)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Errorf("Expected nil but got %v", err)
	}
	time.Sleep(2 * time.Second)
	// confirm the transactionID

	file := x.wal.getCurrentFileName()
	data, err := x.ps.fs.ReadLatestFromFile(file)
	if err != nil {
		t.Fatal(err)
	}
	wal_entry, err := deserialize([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	want := wal_entry.TxnID
	got := txnID
	if want != got {
		t.Errorf("want %v , but got %v", want, got)
	}

	// Now check if key is returned or not

	_, g := x.hm.Get(k)
	w := err_NodeNotFound
	if w != g {
		t.Errorf("Expected %v but got %v", w, g)
	}

	os.RemoveAll(dir)
	os.RemoveAll(ps_prefix)
}

func PrepKV() (string, string, string) {
	ps_prefix := "../../data/kvservice/persist/"
	os.RemoveAll(ps_prefix)
	ps := NewPersistance(ps_prefix)

	for i := 0; i < 10; i++ {
		node := NewNode(fmt.Sprintf("key%v", i), "Some Value")
		ps.Add(*node)
	}
	wal_prefix := "node"
	dir := "../../data/kvservice/wal/"
	return ps_prefix, wal_prefix, dir
}
