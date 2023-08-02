package consensus

import (
	"bufio"
	"log"
	"os"
	"time"

	"github.com/kdsama/kdb/server/logger"
)

// here I need to implement the acknowledgement first
// I need a map of grpc connections as well
// create a map of all the connections for all the servers

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
// So A node will connect to all the nodes
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

func (cs *ConsensusService) connectClients() {
	// get the list of address from the servers.txt
	// run the below function in different goroutines
	// lets start with just

	arr, err := ReadFromFile(cs.filepath)
	if err != nil {
		log.Fatal(err)
	}
	// The problem here is we need to setup a leader and do the connection afterwards
	// so it will be better if I put everything in a map first
	// and add new ones to the list by doing a cron call every n seconds
	// this needs to be re-written
	// leader should be set here before any connection
	// and each of them should have the information about the leader as well
	// need to sit and think this one through

	for _, node := range arr {
		node := node
		if _, ok := clients[node]; ok {
			continue
		}
		ap := "node" + node
		if ap == cs.name {
			continue
		}
		conn := connect(ap)
		nc := NewClient(ap, &conn, 5, cs.logger)
		if nc.leader {

			go nc.Schedule()
		}

	}

}