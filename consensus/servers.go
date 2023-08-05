package consensus

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	pb "github.com/kdsama/kdb/consensus/protodata"
	"github.com/kdsama/kdb/server/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// here I need to implement the acknowledgement first
// I need a maddr of grpc connections as well
// create a maddr of all the connections for all the servers
const (
	addressPort = ":50051"
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

	clientMux sync.Mutex
	clients   map[string]*Client
}

func NewConsensusService(leader bool, name string, filepath string, logger *logger.Logger) *ConsensusService {
	ticker := time.NewTicker(time.Duration(5) * time.Second)
	return &ConsensusService{
		leader:    leader,
		name:      name,
		filepath:  filepath,
		logger:    logger,
		ticker:    ticker,
		clientMux: sync.Mutex{},
		clients:   map[string]*Client{},
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

func (cs *ConsensusService) SendTransaction(data []byte, TxnID string) {
	// send transactions to all the clients
	// when quorum is reached
	// send TransactionConfirmation To all the clients
	// what I am thinking of doing is
	// if quorum is reached we add a value to a transaction map
	// and change its value from 0 to 1
	// on the client layer we implement a lock interface
	// and do spinup locks that when the value is set to 1
	// we can proceed to get into acknowledgement
	// but what if the value doesnot reach 1 ? or it reaches -1 ?
	// in that case we need to abort transaction
	// should we be doing all this in the client layer or here in the transaction layer itself ?
	// i think it should be done here in the transaction layer
	// so need to figure out the steps for that
	// the client shouldn't even know that there is a transaction acknowledgement
	d := data
	type Response struct {
		err  error
		name string
	}
	resp := make(chan Response)
	ctx := context.WithValue(context.Background(), "transaction-ID", TxnID)
	count := 0
	errCount := 0
	quorum := len(cs.clients) / 2
	for _, client := range cs.clients {
		client := client
		// we are not going to delete the client from multiple location
		if client.name == cs.name || client.delete {
			continue
		}
		// create a channel
		// send a value to the channel when we receive non-error
		// open a switch case in for-select context
		// whenever we receive something in the channel
		// we do a counter++
		// and once the counter reaches the quorum consistency
		// we go ahead with sending COnfirmation record to all the servers
		// the method shared about is okay for single context, but we have multiple contexts
		// what will be the termination condition for the context

		go func() {
			err := client.SendRecord(ctx, &d)
			resp <- Response{err, client.name}
		}()

	}
	for {
		select {
		case <-ctx.Done():
			// does this mean the parent context is finished ?
		case res := <-resp:
			{
				if res.err != nil {
					errCount++
				} else {
					count++
				}
				if count >= quorum {
					ctx.Done()
				}

			}

		}
	}
}

func (cs *ConsensusService) SendTransactionConfirmation(data []byte) {
	// send transactions to all the clients
	// when quorum is reached
	// send TransactionConfirmation To all the clients
	d := data
	for _, client := range cs.clients {
		client.SendRecord(&d)
	}
}

func (cs *ConsensusService) connectClients() {
	// get the list of address from the servers.txt
	// run the below function in different goroutines
	// lets start with just

	addresses, err := ReadFromFile(cs.filepath)
	if err != nil {
		log.Fatal(err)
	}

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

		nc := NewClient(addr, conn, 5, cs.logger)
		if cs.leader {
			go nc.Schedule()
		}

	}

}
func connect(addr string) *pb.ConsensusClient {
	flag.Parse()
	addr += addressPort
	fmt.Println("addr", addr)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	// defer conn.Close()
	c := pb.NewConsensusClient(conn)

	return &c
}
