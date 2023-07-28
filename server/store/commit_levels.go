package store

type commitLevel int8

// these are the wal Commit levels
const (
	WAITING commitLevel = iota
	COMMITTED
	ABORTED
)
