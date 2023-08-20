package consensus

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/kdsama/kdb/config"
	"github.com/kdsama/kdb/logger"
)

// here I need to implement the acknowledgement first
// I need a maddr of grpc connections as well
// create a maddr of all the connections for all the servers

const (
	addressPort = ":50051"
)

var (
	errTransactionAborted = errors.New("the transaction was aborted, this can happen if enough acknowledgement arenot received")
	errTransactionBroken  = errors.New("couldnot reach quorum after getting enough acknowledements from other nodes")
)

// basically we will connect to all the servers from the leader
// send acknowledgement from the leader to the receivers
// so send ACK should be just done by the leader
// sending WAL Entries will also be done just by the leader
// So A addr will connect to all the addrs
// those connections should be persisted somewhere, like or an array
// on those connections, if leader, we send acknowlgements
// Leader Election needs to be checked, how it is done.

type stateLevel int32

// redundancy in state if I have a leader field as well
// its better to remove the leader field, and move this to configuration as well
// we will do this later though
const (
	Initializing stateLevel = iota
	Follower
	Candidate
	Leader
)

type ConsensusService struct {
	currLeader string // currentLeader
	name       string // name of our server
	logger     *logger.Logger
	ticker     *time.Ticker      // duration after which the leader sends the heartbeat
	recTicker  *time.Ticker      // duration after its checked whether a heartbeat is received
	clientMux  *sync.Mutex       // mutex Lock
	clients    map[string]*Nodes // recently connect clients
	wg         map[string]*sync.WaitGroup
	state      stateLevel // state is stored (Leader,Follower or Candidate)
	active     int        // active nodes
	lastBeat   time.Time  // Last time a heartbeat was received, if not a leader.
	term       int        // election term
	init       bool       // turns true when the service is initialized
}

func NewConsensusService(name string, logger *logger.Logger) *ConsensusService {
	return &ConsensusService{
		name:      name,
		logger:    logger,
		ticker:    nil,
		clientMux: &sync.Mutex{},
		clients:   map[string]*Nodes{},
		wg:        map[string]*sync.WaitGroup{},
		state:     Initializing,
		active:    0,
		lastBeat:  time.Now(),
		recTicker: nil,
		init:      false,
	}
}

// initializes and set our selves as a follower
func (cs *ConsensusService) Init() {
	cs.state = Follower
	cs.logger.Infof("Waiting for the first broadcast")
	// go cs.Schedule()
}

// schedules timer for heartbeats and heartbeat acknowledgements
func (cs *ConsensusService) Schedule() {

	for {
		select {

		case <-cs.recTicker.C:
			cs.lastHeatBeatCheck()
		case <-cs.ticker.C:
			cs.connectClients()

		}

	}
}

func (cs *ConsensusService) Get(cname string, key string) (string, error) {
	client, ok := cs.clients[cname]
	if !ok {
		return "", errors.New("client not found")
	}
	value, err := client.GetRecord(context.Background(), key)

	return value, err
}

// sends unconfirmed transactions to all the clients
func (cs *ConsensusService) SendTransaction(data []byte, TxnID string) error {
	d := data

	ctx := context.WithValue(context.Background(), "transaction-ID", TxnID)

	var (
		count    int
		errCount int
		quorum   = (len(cs.clients) - 1) / 2
		wg       sync.WaitGroup
		resultCh = make(chan error, len(cs.clients))
	)

	wg.Add(len(cs.clients))

	for _, client := range cs.clients {
		client := client
		go func() {
			defer wg.Done()

			err := client.SendRecord(ctx, &d, config.Acknowledge)

			resultCh <- err
		}()
	}

	wg.Wait()
	close(resultCh)

	for err := range resultCh {
		if err != nil {
			errCount++
		} else {
			count++
		}
	}
	// quorum check, proceed only if we have gotten atleast half acknowledgements on the transaction
	if count > quorum {

		return nil
	}
	return errTransactionAborted
}

// transaction confirmation
func (cs *ConsensusService) SendTransactionConfirmation(data []byte, TxnID string, state config.RecordState) error {

	d := data

	ctx := context.WithValue(context.Background(), "transaction-ID", TxnID)

	var (
		count    int
		errCount int
		quorum   = (len(cs.clients) - 1) / 2
		wg       sync.WaitGroup
		resultCh = make(chan error, len(cs.clients))
	)

	wg.Add(len(cs.clients))

	for _, client := range cs.clients {
		client := client
		go func() {
			defer wg.Done()

			err := client.SendRecord(ctx, &d, config.Commit)
			resultCh <- err
		}()
	}

	wg.Wait()
	close(resultCh)

	for err := range resultCh {
		if err != nil {
			errCount++
		} else {
			count++
		}
	}
	if count > quorum {
		return nil
	}
	return errTransactionBroken
}
