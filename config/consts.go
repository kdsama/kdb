package config

type RecordState int32

const (
	Acknowledge RecordState = iota
	Commit
	Abort
)

const (
	DataPrefix        = "./data/kvservice/persist/"
	WalPrefix         = "node"
	Directory         = "./data/kvservice/wal/"
	WalBufferInterval = 1
	BtreeDegree       = 1000
)
