package clientdiscovery

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	pb "github.com/kdsama/kdb/protodata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// we need to have connections to the servers present here
// we would also need

type Nodes struct {
	leader bool
	addr   string
	name   string
	con    *pb.ConsensusClient
}

type service struct {
	clients   map[string]*Nodes
	leader    string
	addresses []string
}

func NewService() *service {
	return &service{}
}

func (s *service) addServer(addr string) error {
	_, ok := s.clients[addr]
	if !ok {
		client, err := connect(addr)
		if err != nil {
			return errors.New("cannot connect to the server")

		}
		s.clients[addr] = &Nodes{
			leader: false,
			addr:   addr,
			name:   addr,
			con:    client,
		}
		s.addresses = append(s.addresses, addr)
		if len(s.addresses)-1 == 0 {
			s.clients[addr].leader = true
			s.leader = addr
			// first server<-- make it the leader
		}
		s.broadcast(addr)

	} else {
		return errors.New("Already present, so not going to broadcast or make it a leader")
	}
	return nil
}

func connect(addr string) (*pb.ConsensusClient, error) {

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	// defer conn.Close()
	c := pb.NewConsensusClient(conn)

	return &c, nil
}

// we will broadcast and wont rely on the leader to tell everyone.
// once broadcast message is sent, everyone will update the information about all the nodes
func (s *service) broadcast(addr string) error {
	// broadcast about the incoming server to all the servers
	// broadcast message will include leader name and node added
	// we can verify leader name on each broadcast

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	wg := sync.WaitGroup{}
	for _, addr := range s.addresses {
		n := *s.clients[addr].con
		wg.Add(1)
		go func() {
			defer wg.Done()
			response, err := n.Broadcast(ctx, &pb.BroadcastNode{Addr: addr, Leader: s.leader})
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(response)
		}()
	}
	wg.Wait()

	return nil
}
