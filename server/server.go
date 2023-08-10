package server

import (
	"fmt"

	"github.com/kdsama/kdb/config"
	"github.com/kdsama/kdb/consensus"
	"github.com/kdsama/kdb/logger"
	"github.com/kdsama/kdb/store"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ps_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"reqtype"},
	)
)

// all of this should be done in the consensus folder for now
// no time to create another layer
// we will call these functions using gRPC
// that means The task of renaming files have started as well.
// maybe I would still need to create another layer
// it should not be inside the consensus but it should call the consensus
// So a Receiver package that communicates with the Server
// How the restructuring of the the server folder wil go ?
// so there is a sender (leader) and there is receivers
// each receiver receives stuff . I have named it Server
// a better name would be a connection or a serverNodes
// so selfNode will send data to other nodes
// so we can rename ourself as selfNodes
// and others as Nodes ?
// selfNode is a leader or not ? can name it like that ?
// Still need a better name for it
// for the receiver file, it has rpc receiver functions
// What should be its name ?
// so it can be considered as functions for rpcHandlers ?

type Server struct {
	kv      *store.KVService
	cs      *consensus.ConsensusService
	logger  *logger.Logger
	userMap *map[string]string
}

func New(kv *store.KVService, cs *consensus.ConsensusService, logger *logger.Logger) *Server {

	return &Server{
		kv:      kv,
		cs:      cs,
		logger:  logger,
		userMap: &map[string]string{},
	}
}

func (s *Server) Add(key, value string) error {
	requestsTotal.WithLabelValues("Set Key").Inc()

	entry, err := s.kv.Add(key, value)
	if err != nil {
		s.logger.Errorf("%v", err)
		return err
	}

	// when we get the entry we should send the entry to the consensus service
	// should we implement spinning locks for this ?
	// what will be the criteria for that ?
	dat, err := s.kv.SerializeRecord(&entry)
	if err != nil {
		s.logger.Errorf("%v", err)
		return err
	}
	err = s.cs.SendTransaction(dat, entry.TxnID)
	if err != nil {
		s.logger.Fatalf("%v", err)
		return err
	}
	err = s.kv.SetRecord(&dat)
	if err != nil {
		s.logger.Fatalf("%v", err)
	}
	return s.cs.SendTransactionConfirmation(dat, entry.TxnID, config.Commit)

}

func (s *Server) Get(key string) (string, error) {
	requestsTotal.WithLabelValues("Get Key").Inc()
	// 	Now that we are at the server
	// we dont need to do the connection work
	// We just need to return what kv revturns here
	// we just need to modify the data a littlebit
	// the entry point to server will still be consensus
	// but the consensus package wont have access to the kv service , it has to go through server
	// but if server is entry point to consensus
	// how is server doing the manipulation ?
	// so its the handler thats present in the consensus that we are interested in
	// hence I strongly believe that the rpc handler should move to the server  folder

	node, err := s.kv.GetNode(key)
	if err != nil {
		return "", err
	}
	return node.Value, nil

}

func (s *Server) AcknowledgeRecord(data *[]byte) error {
	return s.kv.AcknowledgeRecord(data)
}

func (s *Server) SetRecord(data *[]byte) error {
	return s.kv.SetRecord(data)
}
func (s *Server) Broadcast(addr, leader string) error {
	fmt.Println("CS is ", s.cs)
	return s.cs.Broadcast(addr, leader)
}

func init() {
	prometheus.MustRegister(requestsTotal)
}
