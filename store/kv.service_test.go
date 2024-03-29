package store

import (
	"fmt"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
)

var (
	kv_wal_prefix = "node"
	kv_dir        = "../../data/kvservice/wal/"
	ps_prefix     = "../../data/kvservice/persist/"
)

// next big thing that I need to do is , save the last transactionID which was persisted in the storage layer
// and create a service that will pick the ones that are after that
// how can we pick those ?
// Its important to iterate the fields after that
// thats because We might have multiple adds or saves in the wal entry
// we would also have to skip the acknowledgements and other stuff
// Should we also have a persistance buffer ?
// wal buffer doesn't make sense as of now.
// Or does it ?
var (
	rg, _ = zap.NewDevelopment()
	lg    = rg.Sugar()
)

func TestKVInit(t *testing.T) {

	// to test kv init
	// we probably should dump some data first
	// so lets persist some data first
	// Now I will save a few keys
	// and I should be able to get data of these keys
	PrepKV()
	x := NewKVService(ps_prefix, kv_wal_prefix, kv_dir, 1, 10, lg)
	x.Init()
	got := x.btree.getKeysFromPrefix("key")
	fmt.Println(got)
	want := 10
	if len(got) != want {
		t.Errorf("Wanted the length to be %v, but got %v", want, len(got))
	}
	// os.RemoveAll(ps_prefix)
	// os.RemoveAll(kv_dir)
}

func TestGetNode(t *testing.T) {
	// the test has made me realize that the setting part, in the hashmap is inconsistent.
	// saving key9 at key8 is just plain wrong
	PrepKV()
	x := NewKVService(ps_prefix, kv_wal_prefix, kv_dir, 1, 10, lg)
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
	PrepKV()
	x := NewKVService(ps_prefix, kv_wal_prefix, kv_dir, 1, 10, lg)
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
	PrepKV()
	x := NewKVService(ps_prefix, kv_wal_prefix, kv_dir, 1, 10, lg)
	x.Init()
	k, v := "newKey", "newValue"
	walEntry, err := x.Add(k, v)
	if err != nil {
		t.Errorf("Expected nil but got %v", err)
	}
	we, err := walEntry.serialize()
	if err != nil {
		t.Fatal("WTFAS", err)
	}
	x.SetRecord(&we)
	time.Sleep(3 * time.Second)
	// confirm the transactionID

	file := x.wal.getCurrentFileName()
	data, err := x.ps.fs.ReadLatestFromFile(file)
	if err != nil {
		t.Fatal("WTF", err)
	}
	wal_entry, err := deserialize([]byte(data))
	if err != nil {
		t.Fatal("WTFAS", err)
	}
	want := wal_entry.TxnID
	got := walEntry.TxnID
	if want != got {
		t.Errorf("want %v , but got %v", want, got)
	}
	os.RemoveAll(kv_dir)
	os.RemoveAll(ps_prefix)
}

// func TestUpdate(t *testing.T) {
// 	PrepKV()
// 	x := NewKVService(ps_prefix, kv_wal_prefix, kv_dir, 1, 10, lg)
// 	x.Init()
// 	k, v := "newKey", "newValue"
// 	txnID, err := x.Update(k, v)
// 	if err != nil {
// 		t.Errorf("Expected nil but got %v", err)
// 	}
// 	time.Sleep(2 * time.Second)
// 	// confirm the transactionID

// 	file := x.wal.getCurrentFileName()
// 	data, err := x.ps.fs.ReadLatestFromFile(file)
// 	if err != nil {
// 		t.Fatal("ASASSASA", err)
// 	}
// 	wal_entry, err := deserialize([]byte(data))
// 	if err != nil {
// 		t.Fatal("ASdasqwe12312312", err)
// 	}
// 	want := wal_entry.TxnID
// 	got := txnID
// 	if want != got.TxnID {
// 		t.Errorf("want %v , but got %v", want, got)
// 	}
// 	os.RemoveAll(kv_dir)
// 	os.RemoveAll(ps_prefix)
// }

func TestGetLastTransaction(t *testing.T) {
	PrepKV()
	x := NewKVService(ps_prefix, kv_wal_prefix, kv_dir, 1, 10, lg)
	x.Init()

	k, v := "newKey", "newValue"
	walEntry, err := x.Add(k, v)
	if err != nil {
		t.Errorf("Expected nil but got %v", err)
	}
	we, err := walEntry.serialize()
	if err != nil {
		t.Fatal("WTFAS", err)
	}
	x.SetRecord(&we)
	want := walEntry.TxnID
	got, _ := x.GetLastTransaction()
	time.Sleep(2 * time.Second)

	if want != got {
		t.Errorf("want %v , but got %v", want, got)
	}
	os.RemoveAll(kv_dir)
	os.RemoveAll(ps_prefix)

}

func BenchmarkLastTransactions(t *testing.B) {

	PrepKV()
	x := NewKVService(ps_prefix, kv_wal_prefix, kv_dir, 1, 10, lg)
	x.Init()
	k, v := "anotherkey", "anotherValue"
	var want string
	for j := 0; j < t.N; j++ {

		for i := 0; i < 100000; i++ {
			// k, v := "newKey", "newValue"
			walEntry, err := x.Add(k, v)
			want = walEntry.TxnID
			if err != nil {
				t.Errorf("Expected nil but got %v", err)
			}
			we, err := walEntry.serialize()
			if err != nil {
				t.Fatal("WTFAS", err)
			}
			x.SetRecord(&we)

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

}

// the test cases need to change because of distributed transaction nature now

func TestDelete(t *testing.T) {
	PrepKV()
	x := NewKVService(ps_prefix, kv_wal_prefix, kv_dir, 1, 10, lg)
	x.Init()
	k, v := "newKey", "newValue"
	x.Add(k, v)
	time.Sleep(2 * time.Second)
	want := err_NodeNotFound
	_, got := x.Delete(k)
	if got != want {
		t.Errorf("Wanted %v but got %v", want, got)
	}
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// if err != nil {
	// 	t.Errorf("Expected nil but got %v", err)
	// }
	// time.Sleep(2 * time.Second)
	// // confirm the transactionID

	// file := x.wal.getCurrentFileName()
	// data, err := x.ps.fs.ReadLatestFromFile(file)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// wal_entry, err := deserialize([]byte(data))
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// want := wal_entry.TxnID
	// got := wle.TxnID
	// if want != got {
	// 	t.Errorf("want %v , but got %v", want, got)
	// }

	// // Now check if key is returned or not

	// _, g := x.hm.Get(k)
	// w := err_NodeNotFound
	// if w != g {
	// 	t.Errorf("Expected %v but got %v", w, g)
	// }

	os.RemoveAll(kv_dir)
	os.RemoveAll(ps_prefix)
}

func TestSetRecord(t *testing.T) {
	PrepKV()
	x := NewKVService(ps_prefix, kv_wal_prefix, kv_dir, 1, 10, lg)
	x.Init()
	n := NewNode("someKey", "SomeValue")
	txnID := "node100000"
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

func TestCommit(t *testing.T) {
	PrepKV()

	x := NewKVService(ps_prefix, kv_wal_prefix, kv_dir, 1, 10, lg)
	x.Init()

	n := NewNode("someKey", "SomeValue")
	txnID := "node100000"
	entry := NewWalEntry(n, ADD, txnID)
	got := x.Commit(entry)

	var want error
	if want != got {
		t.Errorf("want %v , but got %v", want, got)
	}

	t.Run("Trying to commit the same version as already saved", func(t *testing.T) {
		want = err_Upserted
		got = x.Commit(entry)
		if want != got {
			t.Errorf("want %v , but got %v", want, got)
		}

	})
	t.Run("Check the data in the file", func(t *testing.T) {
		key := n.Key
		node, err := x.ps.GetNodeFromKey(key)
		if err != nil {
			t.Errorf("Expected nil but got %v", err)
		}
		if n.Key != node.Key && n.Value != node.Value && n.Version != node.Version {
			t.Errorf("Wanted %v but got %v", n, node)
		}
	})

}

func PrepKV() {
	// ps_prefix := "../../data/kvservice/persist/"
	os.RemoveAll(ps_prefix)
	os.RemoveAll(kv_dir)
	ps := NewPersistance(ps_prefix, lg)

	for i := 0; i < 10; i++ {
		node := NewNode(fmt.Sprintf("key%v", i), fmt.Sprintf("Some Value:%v", i))
		ps.Add(*node)
	}
	// kv_wal_prefix := "node"
	// kv_dir := "../../data/kvservice/wal/"
	// return ps_prefix, kv_wal_prefix, kv_dir
}
