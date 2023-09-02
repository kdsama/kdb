package server

import (
	"github.com/kdsama/kdb/config"
	"github.com/kdsama/kdb/consensus"
	"github.com/kdsama/kdb/store"
	"go.uber.org/zap"
)

type Server struct {
	kv     store.KVer
	cs     *consensus.ConsensusService
	logger *zap.SugaredLogger
	TsMap  *map[string]string
}

func New(kv *store.KVService, cs *consensus.ConsensusService, logger *zap.SugaredLogger) *Server {

	return &Server{
		kv:     kv,
		cs:     cs,
		logger: logger,
		TsMap:  &map[string]string{},
	}
}

// serializes the record, adds to wal. And when it gets enough acknowledgements,
// commits the data and send confirmation to all the node
func (s *Server) Add(key, value string) error {
	requestsTotal.WithLabelValues("Set Key").Inc()

	entry, err := s.kv.Add(key, value)
	if err != nil {
		s.logger.Errorf("%v", err)
		return err
	}

	dat, err := s.kv.SerializeRecord(&entry)
	if err != nil {
		s.logger.Errorf("%v", err)
		return err
	}
	err = s.cs.SendTransaction(dat, entry.TxnID)

	if err != nil {
		s.logger.Errorf("%v", err)
		return err
	}
	// Here we have to do something with transactionID
	err = s.kv.SetRecord(&dat)
	if err != nil {
		s.logger.Fatalf("%v", err)
	}
	return s.cs.SendTransactionConfirmation(dat, entry.TxnID, config.Commit)

}

func (s *Server) Get(key string) (string, error) {
	requestsTotal.WithLabelValues("Get Key").Inc()
	node, err := s.kv.GetNode(key)
	if err != nil {
		return "", err
	}
	return node.Value, nil

}

func (s *Server) HeartbeatAck() {
	s.cs.HeartbeatAck()
}

func (s *Server) AcknowledgeRecord(data *[]byte) error {
	return s.kv.AcknowledgeRecord(data)
}

func (s *Server) SetRecord(data *[]byte) error {
	return s.kv.SetRecord(data)
}
func (s *Server) Broadcast(addr []string, leader string) error {

	return s.cs.Broadcast(addr)
}

func (s *Server) Vote(term int, leader string, votes []string) (string, bool) {
	return s.cs.Vote(term, leader, votes)
}

func (s *Server) LeaderInfo() (string, error) {
	return s.cs.LeaderInfo()
}
