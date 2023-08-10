package server

import (
	"fmt"

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
	userMap map[string]string
}

func New(kv *store.KVService, cs *consensus.ConsensusService, logger *logger.Logger) *Server {

	return &Server{
		kv:      kv,
		cs:      cs,
		logger:  logger,
		userMap: map[string]string{},
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
	return s.cs.SendTransactionConfirmation(dat, entry.TxnID, consensus.Commit)

}

func (s *Server) Get(user string, key string) (string, error) {
	requestsTotal.WithLabelValues("Get Key").Inc()
	// here what we can do is
	// set a Server connection to a particular Server for the current user for reading purposes
	// so n users will have a random Server attached to it to fetch the get requests
	// for now I can just ask the system to get me data through gcp ,
	// so need protobuf here for get services

	if _, ok := s.userMap[user]; !ok {

		n := s.cs.GetRandomConnectionName()
		fmt.Println("Random connection is ", n)
		s.userMap[user] = n
	}

	return s.cs.Get(s.userMap[user], key)

}

func init() {
	prometheus.MustRegister(requestsTotal)
}
