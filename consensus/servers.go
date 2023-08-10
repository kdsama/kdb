package consensus

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	pb "github.com/kdsama/kdb/consensus/protodata"
	"github.com/kdsama/kdb/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// here I need to implement the acknowledgement first
// I need a maddr of grpc connections as well
// create a maddr of all the connections for all the servers
type recordState int32

const (
	addressPort             = ":50051"
	Acknowledge recordState = iota
	Commit
	Abort
)

var (
	errTransactionAborted = errors.New("the transaction was aborted, this can happen if enough acknowledgement arenot received")
	errTransactionBroken  = errors.New("couldnot reach quorum after getting enough acknowledements from other nodes")
)

func ReadFromFile(filepath string) ([]string, error) {
	t := []string{}
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		t = append(t, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return t, nil
}

// basically we will connect to all the servers from the leader
// send acknowledgement from the leader to the receivers
// so send ACK should be just done by the leader
// sending WAL Entries will also be done just by the leader
// So A addr will connect to all the addrs
// those connections should be persisted somewhere, like or an array
// on those connections, if leader, we send acknowlgements
// Leader Election needs to be checked, how it is done.

type ConsensusService struct {
	leader   bool
	name     string
	filepath string
	logger   *logger.Logger
	ticker   *time.Ticker

	clientMux *sync.Mutex
	clients   map[string]*Client

	wg        map[string]*sync.WaitGroup
	addresses []string
}

func NewConsensusService(leader bool, name string, filepath string, logger *logger.Logger) *ConsensusService {
	ticker := time.NewTicker(time.Duration(5) * time.Second)
	return &ConsensusService{
		leader:    leader,
		name:      name,
		filepath:  filepath,
		logger:    logger,
		ticker:    ticker,
		clientMux: &sync.Mutex{},
		clients:   map[string]*Client{},
		wg:        map[string]*sync.WaitGroup{},
		addresses: []string{},
	}
}

func (cs *ConsensusService) Init() {

	go cs.Schedule()
}
func (cs *ConsensusService) Schedule() {

	for {
		select {
		case <-cs.ticker.C:
			cs.connectClients()
		}
	}
}

func (cs *ConsensusService) GetRandomConnectionName() string {
	rand.Seed(time.Now().Unix())
	mut := sync.Mutex{}
	ll := len(cs.addresses)

	randomServer := cs.addresses[rand.Intn(ll)]
	mut.Lock()
	if randomServer == cs.name {
		return cs.GetRandomConnectionName()
	}
	return randomServer

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

			err := client.SendRecord(ctx, &d, Acknowledge)
			resultCh <- err
		}()
	}

	wg.Wait()
	close(resultCh)

	for err := range resultCh {
		if err != nil {
			fmt.Println("???????????????????????????????????????????????")
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

func (cs *ConsensusService) SendTransactionConfirmation(data []byte, TxnID string, state recordState) error {

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

			err := client.SendRecord(ctx, &d, Commit)
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

func (cs *ConsensusService) connectClients() {
	// get the list of address from the servers.txt
	// run the below function in different goroutines
	// lets start with just

	addresses, err := ReadFromFile(cs.filepath)
	if err != nil {
		log.Fatal(err)
	}
	l := sync.Mutex{}
	l.Lock()
	cs.addresses = addresses
	l.Unlock()
	// The problem here is we need to setup a leader and do the connection afterwards
	// so it will be better if I put everything in a maddr first
	// and add new ones to the list by doing a cron call every n seconds
	// this needs to be re-written
	// leader should be set here before any connection
	// and each of them should have the information about the leader as well
	// need to sit and think this one through

	for _, addr := range addresses {
		addr := addr
		val, ok := cs.clients[addr]
		if ok {
			// has the client layer marked itself to be deleted ?
			if val.delete {
				cs.clientMux.Lock()
				delete(cs.clients, val.name)
				cs.clientMux.Unlock()
			} else {

				continue
			}

		}
		if addr == cs.name {
			continue
		}

		conn := connect(addr)

		nc := NewClient(addr, conn, 7, cs.logger)
		cs.clients[nc.name] = nc
		if cs.leader {
			go nc.Schedule()
		}

	}

}
func connect(addr string) *pb.ConsensusClient {
	flag.Parse()
	addr += addressPort
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	// defer conn.Close()
	c := pb.NewConsensusClient(conn)

	return &c
}
