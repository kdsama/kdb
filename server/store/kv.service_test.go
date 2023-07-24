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

func TestGetNode(t *testing.T) {
	// the test has made me realize that the setting part, in the hashmap is inconsistent.
	// saving key9 at key8 is just plain wrong
	ps_prefix, wal_prefix, dir := PrepKV()
	x := NewKVService(ps_prefix, wal_prefix, dir, 1, 10)
	x.Init()
	want := "key1"
	node, err := x.GetNode(want)
	if err != nil {
		t.Error("Expected no error but got ", err)
	}
	got := node.Key
	if want != got {
		t.Errorf("Wanted %v, but got %v", want, got)
	}

	t.Run("Error scenario for GetNode", func(t *testing.T) {
		_, got := x.GetNode("AbsentKey")
		want := err_NodeNotFound
		if got != want {
			t.Errorf("want %v but got %v", want, got)
		}
	})

}

func TestGetManyNodes(t *testing.T) {
	// the test has made me realize that the setting part, in the hashmap is inconsistent.
	// saving key9 at key8 is just plain wrong
	ps_prefix, wal_prefix, dir := PrepKV()
	x := NewKVService(ps_prefix, wal_prefix, dir, 1, 10)
	x.Init()
	key := "key"
	nodes, err := x.GetManyNodes(key)
	if err != nil {
		t.Error("Expected no error but got ", err)
	}
	got := len(nodes)
	want := 10
	if got != want {
		t.Errorf("want %v but got %v", want, got)
	}

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

func TestGetLastTransaction(t *testing.T) {
	ps_prefix, wal_prefix, dir := PrepKV()
	x := NewKVService(ps_prefix, wal_prefix, dir, 1, 10)
	x.Init()
	k, v := "anotherkey", "anotherValue"
	want, err := x.Add(k, v)
	if err != nil {
		t.Error("Didnot expect error but got ", err)

	}
	got, err := x.GetLastTransaction()
	if err != nil {
		t.Error("Didnot expect error but got ", err)

	}
	time.Sleep(2 * time.Second)
	if want != got {
		t.Errorf("want %v , but got %v", want, got)
	}
	os.RemoveAll(dir)
	os.RemoveAll(ps_prefix)
	t.Run("Add thousands of transactions so we have multiple files, reinitiate the kv service and get the latest transaction", func(t *testing.T) {
		ps_prefix, wal_prefix, dir := PrepKV()
		x := NewKVService(ps_prefix, wal_prefix, dir, 1, 10)
		x.Init()
		k, v := "anotherkey", "anotherValue"
		var want string
		for i := 0; i < 1000000; i++ {
			want, err = x.Add(k+fmt.Sprintf("%v", i), v+fmt.Sprintf("%v", i))
			if err != nil {
				t.Error("Didnot expect error but got ", err)

			}

		}
		time.Sleep(4 * time.Second)

		got, err := x.GetLastTransaction()
		if err != nil {
			t.Error("Didnot expect error but got ", err)

		}

		if want != got {
			t.Errorf("want %v , but got %v", want, got)
		}
	})
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

func TestSetRecord(t *testing.T) {
	ps_prefix, wal_prefix, dir := PrepKV()
	x := NewKVService(ps_prefix, wal_prefix, dir, 1, 10)
	x.Init()
	n := NewNode("someKey", "SomeValue")
	txnID := "SomeID100000"
	entry := NewWalEntry(n, ADD, txnID)
	en, err := entry.serialize()
	if err != nil {
		t.Error("Expected no error but got", err)
	}
	err = x.SetRecord(&en)
	if err != nil {
		t.Error("Expected no error but got", err)
	}
	time.Sleep(2 * time.Second)
	// compare transactionRecords
	got, err := x.GetLastTransaction()
	if err != nil {
		t.Error("Expected no error but got", err)
	}
	want := txnID
	if want != got {
		t.Errorf("want %v , but got %v", want, got)
	}
}

func PrepKV() (string, string, string) {
	ps_prefix := "../../data/kvservice/persist/"
	os.RemoveAll(ps_prefix)
	ps := NewPersistance(ps_prefix)

	for i := 0; i < 10; i++ {
		node := NewNode(fmt.Sprintf("key%v", i), fmt.Sprintf("Some Value:%v", i))
		ps.Add(*node)
	}
	wal_prefix := "node"
	dir := "../../data/kvservice/wal/"
	return ps_prefix, wal_prefix, dir
}
