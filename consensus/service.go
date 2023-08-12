package consensus

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/kdsama/kdb/config"
	"github.com/kdsama/kdb/logger"
	pb "github.com/kdsama/kdb/protodata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	leader    bool
	name      string
	logger    *logger.Logger
	ticker    *time.Ticker
	clientMux *sync.Mutex
	clients   map[string]*Nodes
	wg        map[string]*sync.WaitGroup
	addresses []string
	state     stateLevel
	active    int
}

func NewConsensusService(name string, logger *logger.Logger) *ConsensusService {
	ticker := time.NewTicker(time.Duration(5) * time.Second)
	return &ConsensusService{
		leader:    false,
		name:      name,
		logger:    logger,
		ticker:    ticker,
		clientMux: &sync.Mutex{},
		clients:   map[string]*Nodes{},
		wg:        map[string]*sync.WaitGroup{},
		addresses: []string{},
		state:     Initializing,
		active:    0,
	}
}

func (cs *ConsensusService) Init() {
	cs.state = Follower
}

func (cs *ConsensusService) Schedule() {

	for {
		select {
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
			cs.leader = true
			cs.state = Leader
			cs.logger.Infof("I am god, destroyer of the world ")
			go cs.Schedule()
		} else if addr == cs.name {
			cs.state = Follower
		} else {
			cs.addresses = append(cs.addresses, addr)
		}
	}
	return nil
}

func (cs *ConsensusService) connectClients() {
	// get the list of address from the servers.txt
	// run the below function in different goroutines
	// lets start with just
	// now the other node information will be fetched from the client/discovery service

	// The problem here is we need to setup a leader and do the connection afterwards
	// so it will be better if I put everything in a maddr first
	// and add new ones to the list by doing a cron call every n seconds
	// this needs to be re-written
	// leader should be set here before any connection
	// and each of them should have the information about the leader as well
	// need to sit and think this one through

	cs.active = 0
	for _, addr := range cs.addresses {
		cs.active++
		addr := addr
		val, ok := cs.clients[addr]
		if addr == cs.name {

			continue
		}
		if ok {
			// has the client layer marked itself to be deleted ?
			if val.delete {
				cs.clientMux.Lock()
				delete(cs.clients, val.name)
				cs.clientMux.Unlock()
				cs.active--
			}

		} else {

			conn, err := connect(addr)

			if err != nil {
				cs.logger.Errorf("%v", err)
				cs.clientMux.Lock()
				delete(cs.clients, addr)
				cs.clientMux.Unlock()
				cs.active--
				continue

			}
			nc := NewNodes(addr, conn, 7, cs.logger)
			cs.clients[nc.name] = nc
		}
		// we are going to generate heartbeat from the server code instead of nodes.go
		cs.confirmHeartBeat()
	}
}

func (cs *ConsensusService) confirmHeartBeat() {
	// check for quorum
	if !cs.leader {
		return
	}
	if cs.active < len(cs.addresses)/2 {
		// might as well
		cs.logger.Warnf("Quorum is broken , we have %d active nodes out of %d", cs.active, len(cs.addresses))
		// no hard action for now
	}

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

			err := client.Hearbeat()

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

	if count < quorum {
		cs.logger.Errorf("Quorum is broken")
	}

}

func connect(addr string) (*pb.ConsensusClient, error) {
	addr += addressPort
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	// defer conn.Close()
	c := pb.NewConsensusClient(conn)

	return &c, nil
}
