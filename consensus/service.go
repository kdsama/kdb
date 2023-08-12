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
	currLeader string
	name       string
	logger     *logger.Logger
	ticker     *time.Ticker
	recTicker  *time.Ticker
	clientMux  *sync.Mutex
	clients    map[string]*Nodes
	wg         map[string]*sync.WaitGroup
	addresses  []string
	state      stateLevel
	active     int // active nodes
	lastBeat   time.Time
	term       int
}

func NewConsensusService(name string, logger *logger.Logger) *ConsensusService {
	ticker := time.NewTicker(time.Duration(5) * time.Second)
	recTicker := time.NewTicker(time.Duration(3) * time.Second)
	return &ConsensusService{
		name:      name,
		logger:    logger,
		ticker:    ticker,
		clientMux: &sync.Mutex{},
		clients:   map[string]*Nodes{},
		wg:        map[string]*sync.WaitGroup{},
		addresses: []string{},
		state:     Initializing,
		active:    0,
		lastBeat:  time.Now(),
		recTicker: recTicker,
	}
}

func (cs *ConsensusService) Init() {
	cs.state = Follower
	cs.Schedule()
}

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

		if client.name == cs.name || client.delete {
			wg.Done()
			continue
		}

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

	if count > quorum {

		return nil
	}
	return errTransactionAborted
}

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

		if client.name == cs.name || client.delete {
			wg.Done()
			continue
		}

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

func (cs *ConsensusService) Broadcast(addr, leader string) error {
	// add the node if it doesnot exist already
	cs.logger.Infof("%s v %v", addr, leader)
	if _, ok := cs.clients[addr]; !ok {

		client, err := connect(addr)
		if err != nil {
			return err
		}
		// if leader and ticker not running , run it
		cs.clients[addr] = NewNodes(addr, client, 3, cs.logger)
		if addr == leader {
			cs.logger.Infof("Leadership confirmed")

			cs.state = Leader
			cs.logger.Infof("I am god, destroyer of the world ")

		} else if addr == cs.name {
			cs.state = Follower
		}
		cs.addresses = append(cs.addresses, addr)

	}
	if len(cs.addresses) == 1 {
		// means I am the daddy caddy
		//  I am gonna become the leader. At
		cs.electMeAndBroadcast()
	}
	return nil
}
