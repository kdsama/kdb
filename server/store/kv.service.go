package store

import "errors"

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
	hm    HashMap
	btree BTree
	wal   WAL
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

func SetRecord() error {
	// this is for syncing up of data
	// No need transactions are to be generated for this case
	// we will be receiving transaction ID , and the node itself with appropriate information
	return nil
}

func SetRecords() error {
	// This is to set multiple records .
	return nil
}

func GetLastTransaction() error {

	return nil
}

func Load() error {

	return nil
}
