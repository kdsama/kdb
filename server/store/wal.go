package store

import (
	"fmt"
	"io/fs"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kdsama/kdb/logger"
)

/* Key: The key represents the unique identifier or name associated with the data being stored. It is used to retrieve or reference the data in the database.

Value: The value corresponds to the actual data that is being stored or persisted in the database.
It can be any binary or structured data that the application or user wants to store.

Metadata: Metadata provides additional information about the entry, such as timestamps,
version numbers, or flags. This information can be used for various purposes,
including data consistency checks, concurrency control, or tracking modifications.

Operation Type: The operation type indicates the nature of the operation being performed on the entry.
Common operation types include "create" (inserting a new key-value pair),
"update" (modifying an existing key-value pair), or "delete" (removing a key-value pair).

Transoperation ID or Sequence Number: To ensure atomicity and consistency,
WAL entries often include a transoperation ID or a sequence number.
This identifier helps in tracking the order of operations
and ensuring that changes are applied in the correct sequence during recovery or replication scenarios.

Additional Metadata or Flags: Depending on the specific database system and its requirements,
additional metadata or flags may be included in the WAL entry.
These can provide information for handling durability, replication, or other specific functionalities provided by the database.
*/

// If I am writing my own WAL Implementation , this is what will be required atleast
// lampost timestamp , I can write about it in the documentation as well - Use atomic operation for increment of this
// key
// value
// operation
// metadata - What is the information that we should put in metadata.
// we can focus on this metadata when we start with the consensus algorithm
// we will have a separate struct for WAL Log ?

type WAL struct {
	prefix       string
	directory    string
	counter      int64
	file_counter int64
	lock         *sync.Mutex
	fs           fileService
	ticker       time.Ticker
	latestEntry  []byte
	logger       *logger.Logger
}

var (
	wal_buffer = []byte{}
)

const (
	MAX_BUFFER_SIZE = 30000
	MAX_FILE_SIZE   = 100000000
)

var (
	file_counter = 0
)

func NewWAL(prefix, directory string, fs fileService, duration int, lg *logger.Logger) *WAL {

	t := time.Duration(duration) * time.Second
	ticker := *time.NewTicker(t)
	wal := WAL{prefix, directory, 0, 0, &sync.Mutex{}, fs, ticker, []byte{}, lg}

	wal.setLatestFileCounter()
	wal.setLatestCounter()

	go wal.Schedule()
	return &wal
}

// make it in a way that it makes sure file is present. If not present yet, it probably means
// the file counter function returned 0 .
// This means this should also return zero
func (w *WAL) setLatestCounter() {
	// w.fs.
	data, err := w.fs.ReadLatestFromFile(w.getCurrentFileName())

	if err != nil && err == fs.ErrNotExist {
		w.counter = 0
		return

	}
	if len(data) == 0 {
		w.counter = 0
		return
	}

	entry, err := deserialize([]byte(data))
	if err != nil {
		log.Print(err)
	}
	w.counter = w.GetCounterFromTransactionID(entry.TxnID)

}

func (w *WAL) setLatestFileCounter() {
	filename, err := w.fs.GetLatestFile(w.directory)
	if err != nil && err == err_NoFilesInDirectory {
		w.file_counter = 0
		return
	}
	if err != nil {
		w.logger.Errorf(err.Error())
	}
	counter := w.GetCounterFromFileName(filename)
	w.file_counter = counter
}

// atomic incrementing the counter
func (w *WAL) IncrementCounter() int64 {
	atomic.AddInt64(&w.counter, 1)
	return w.counter
}

// atomic incrementing the counter
func (w *WAL) IncrementFileCounter() int64 {
	atomic.AddInt64(&w.file_counter, 1)
	fmt.Println("New file counter ", w.file_counter)
	return w.file_counter
}

func (w *WAL) addEntry(node Node, operation string) (WalEntry, error) {
	newCounter := w.IncrementCounter()
	txnID := w.prefix + fmt.Sprint(newCounter)
	walEntry := NewWalEntry(&node, operation, txnID)
	toAppendData, err := walEntry.serialize()
	toAppendData = append(toAppendData, byte('\n'))
	if err != nil {
		return WalEntry{}, err
	}
	w.lock.Lock()
	w.latestEntry = toAppendData
	wal_buffer = append(wal_buffer, toAppendData...)
	w.lock.Unlock()
	if len(wal_buffer) > MAX_BUFFER_SIZE {
		go w.BufferUpdate()
	}
	return *walEntry, nil
}

func (w *WAL) AddWALEntry(wal *[]byte) {

	arr := append(*wal, byte('\n'))

	w.lock.Lock()

	// we probably can read the last one first
	// check its transactionID
	// if it is bigger than current one
	// panic
	// but for now leaving it as is
	w.latestEntry = arr
	wal_buffer = append(wal_buffer, arr...)
	w.lock.Unlock()
}
func (w *WAL) Schedule() bool {

	for {
		select {
		case <-w.ticker.C:
			{
				w.BufferUpdate()
			}
		}
	}

}
func (w *WAL) BufferUpdate() {
	w.lock.Lock()
	len_buffer := len(wal_buffer)

	// increment counter when the latest file size has exceeded the size we expect the file to be
	size, err := w.fs.GetFileSize(w.getCurrentFileName())
	if err != nil && w.file_counter != 0 {
		// what about the case when there are no files , aka the first time the application is run
		w.logger.Fatalf("Didnot expect an error here %v, closing the application", err)
	}
	// basically the scenario popped up where
	// when we went to the check for max file size , Nothing in previous file was pushed. So everything got pushed into second file.
	// There is no issue with the incrementer but the policy of bufferUpdate

	if int64(len_buffer)+size > int64(MAX_BUFFER_SIZE) {

		w.IncrementFileCounter()
	}

	w.fs.WriteFileWithDirectories(w.getCurrentFileName(), wal_buffer)
	w.flushall()
	w.lock.Unlock()

}

func (w *WAL) flushall() {
	// w.lock.Lock()
	wal_buffer = []byte{}
	// w.lock.Unlock()
}

func (w *WAL) checkSize() int {
	return len(wal_buffer)
}

func (w *WAL) getCurrentFileName() string {
	return w.directory + w.prefix + "-" + fmt.Sprint(w.file_counter) + ".wal"
}

func (w *WAL) GetCounterFromFileName(filename string) int64 {
	if filename == "" {
		return int64(0)
	}
	arr := strings.Split(filename, "-")
	arr1 := strings.Split(arr[1], ".")
	counter, err := strconv.Atoi(arr1[0])
	if err != nil {
		log.Print(err)
		return 0

	}
	return int64(counter)

}

func (w *WAL) GetCounterFromTransactionID(txn string) int64 {
	arr := strings.Split(txn, w.prefix)
	counter, err := strconv.Atoi(arr[1])
	if err != nil {
		log.Print(err)
		return 0
	}
	return int64(counter)
}
