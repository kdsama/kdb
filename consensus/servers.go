package consensus

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
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
}

func NewConsensusService(leader bool, name string, filepath string, logger *logger.Logger) *ConsensusService {
	ticker := time.NewTicker(time.Duration(5) * time.Second)
	return &ConsensusService{leader, name, filepath, logger, ticker}
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

func (cs *ConsensusService) SendTransaction(data []byte) {
	// send transactions to all the clients
	// when quorum is reached
	// send TransactionConfirmation To all the clients
	d := data
	for _, client := range clients {
		client := client
		if client.name == cs.name {
			continue
		}
		go func() {
			client.SendRecord(&d)
		}()

	}
}

func (cs *ConsensusService) SendTransactionConfirmation(data []byte) {
	// send transactions to all the clients
	// when quorum is reached
	// send TransactionConfirmation To all the clients
	d := data
	for _, client := range clients {
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

		if _, ok := clients[addr]; ok {

			continue
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
