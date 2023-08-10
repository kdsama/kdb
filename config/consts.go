package config

type RecordState int32

const (
	Acknowledge RecordState = iota
	Commit
	Abort
)
