package store

import (
	"errors"
	"fmt"
	"sync"
)

// Now here is the service that will bind everything together
// A request will come to add or delete or update
// we just pretend all of them to be the same , and just call the add function here
// we will also have a loadup function, that will load up data from the data store. This is usually on restart
// we probably would need to write a sync up function as welll. Put it on an interface for now
// sync function I think is not required here

// we need a function which on acknowledgement will push the content to persistance layer.
// and on failure will revert the WAL transaction, or we can just put a cancelled tag on it

// here we have another issue
// so we dont have to push data in the WAL or need to push it in a way that will tell if its unconfirmed or not
// once it is confirmed then we have to push it again

// so this is a kv service
// not the consensus service
// think of it as the one that can do it all . It can persist, write to WAL, write to memory
// we dont have to write the part that the consensus algorithm will initiate itself
// First question is , is it parallel to WAL or above WAL ?
// Technically its above WAL
// we have to add commits on these requests
// so the map should ideally also have a commit key . Which changes on acknowledgement.
// so we will add to the map , and the wal
// we will get transactionID
// we will persist that transactionID in the memory for now
// Ideally we would also like to have the file in which that transaction is saved in uncommitted state
// Ideally we will have multiple states --> waiting, committed and aborted

var (
	ADD    = "ADD"
	UPDATE = "UPDATE"
	DELETE = "DELETE"
)

var (
	err_InvalidAction = errors.New("invalid action")
)

type KVI interface {
	Add(key string, value string) (string, error)
	Update(key string, value string) (string, error)
	Delete(key string) (string, error)
	SetRecord() error
	Load() error
}

type KVService struct {
	hm    *HashMap
	btree BTree
	wal   *WAL
	ps    *Persistance
	mut   sync.Mutex
}

func NewKVService(dataPrefix, walPrefix, directory string, duration int, degree int) *KVService {
	// load all the data
	fs := NewFileService()
	wal := NewWAL(walPrefix, directory, *fs, duration)
	hm := NewHashMap()
	btree := newBTree(degree)
	ps := NewPersistance(dataPrefix)
	return &KVService{hm, btree, wal, ps, sync.Mutex{}}
}
func (kvs *KVService) Init() {
	// loading part
	// load values from persistance layer to btree and hashmap
	// no role of wal here
	// this will set all nodes in the ps object
	kvs.ps.GetALLNodes()
	for _, node := range kvs.ps.nodes {
		kvs.btree.addKeyString(node.Key)
		kvs.hm.AddNode(&node)
	}
	// I dont mind everything to be sequential here as it is just once.
}

// returns transactionID
func (kvs *KVService) Add(key string, value string) (string, error) {

	// write to hashmap
	// write to b-tree
	// write to WAL
	node := kvs.hm.Add(key, value)
	kvs.btree.addKeyString(key)
	return kvs.wal.addEntry(*node, ADD)
}

// updates the key value and returns the transactionID
func (kvs *KVService) Update(key string, value string) (string, error) {

	// write to hashmap
	// write to b-tree
	// write to WAL
	node := kvs.hm.Add(key, value)
	kvs.btree.addKeyString(key)
	return kvs.wal.addEntry(*node, UPDATE)
}

func (kvs *KVService) Delete(key string) (string, error) {
	// dont do anything on the btree part let the key remain for now
	node, err := kvs.hm.Delete(key)
	if err != nil {
		return "", err
	}
	return kvs.wal.addEntry(*node, DELETE)
}

func (kvs *KVService) SetRecord(data *[]byte) error {

	// this is for syncing up of data
	// No newtransactions are to be generated for this case
	// we will be receiving transaction ID , and the node itself with appropriate information
	// so we will have to save a new node in the key value store
	// the node will come with its own timestamp
	// the WAL transaction will also be saved without generating a new transactionID
	// No checks for it for now
	// as we are getting WAL logs we need to serialize it
	walEntry, err := deserialize(*data)
	if err != nil {
		return err
	}
	kvs.btree.addKeyString(walEntry.Node.Key)
	kvs.hm.AddNode(walEntry.Node)
	// better if we send buffer itself here instead of serializing and deserializing again.
	kvs.wal.AddWALEntry(data)
	return nil
}

// how  multiple records will be shared ??

func (kvs *KVService) SetRecords() error {
	// we can chuck this for now
	// the outer layer will just call our function multiple number of times.

	return nil
}

func (kvs *KVService) GetLastTransaction() error {
	// so is this a processed transaction or any transaction ?
	// we need to find the latest one, which is probably in the memory and not in the storage
	// Or it is in the storage
	// so we need to fetch both of them
	// and figure out which one is the latest one
	// And return that
	// We can also save the latest transaction in memory separately
	// so it is always accessible
	// we also would need to set a lock here
	// which means kvs should have a lock of itself
	kvs.mut.Lock()
	latestTransaction, err := deserialize(kvs.wal.latestEntry)
	kvs.mut.Unlock()
	if err != nil {
		return err
	}
	fmt.Println(latestTransaction)
	return nil
}

func (kvs *KVService) Load() error {

	return nil
}
