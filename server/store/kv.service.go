package store

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
//

type KVI interface {
	add(key string, value string, delete bool) error
	load() error
	persist()
}
