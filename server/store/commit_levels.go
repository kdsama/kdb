package store

type commitLevel int8

// these are the wal Commit levels
const (
	Waiting commitLevel = iota
	Committed
	Aborted
)
