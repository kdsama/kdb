package clientdiscovery

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
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

var (
	curr     = 0
	count    = 0
	errCount = 0
)

func NewService() *service {
	return &service{
		clients:   map[string]*Nodes{},
		leader:    "",
		addresses: []string{},
	}
}

func (s *service) addServer(addr string) error {
	_, ok := s.clients[addr]
	if !ok {
		client, err := connect(addr)
		if err != nil {
			return errors.New("cannot connect to the server")

		}
		fmt.Println("New client is ", client)
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

	conn, err := grpc.Dial(addr+":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func (s *service) set(key, val string) error {
	requestsTotal.WithLabelValues("Set").Inc()
	n := *s.clients[s.leader].con
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	_, err := n.Set(ctx, &pb.SetKey{Key: key, Value: val})
	if err != nil {
		return err
	}
	return nil
}

func (s *service) automateSet(duration, requests, sleep string) (int, error) {
	rp, sp := 10000, 10

	if curr != 0 {
		return -1, errors.New("A request is still getting completed")
	}
	rand.Seed(time.Now().Unix())
	curr = rand.Intn(100000)
	if requests != "" {
		rb, err := strconv.Atoi(requests)
		if err == nil {
			rp = rb
		}
	}
	if sleep != "" {
		sb, err := strconv.Atoi(sleep)
		if err == nil {
			sp = sb
		}
	}

	go func() {

		for i := 0; i < rp; i++ {

			time.Sleep(time.Duration(sp) * time.Millisecond)
			n := *s.clients[s.leader].con
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()
			rand.Seed(time.Now().Unix())
			_, err := n.Set(ctx, &pb.SetKey{Key: "key" + fmt.Sprint(rand.Intn(10000)), Value: "val" + fmt.Sprint(rand.Intn(10000))})
			if err != nil {
				errCount++
			} else {
				count++
			}
		}

		curr = 0
	}()

	return curr, nil
}

func (s *service) get(key string) (string, error) {
	cl := s.leader

	n := *s.clients[cl].con
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	val, err := n.Get(ctx, &pb.GetKey{Key: key})
	if err != nil {
		return "", err
	}
	return val.Value, nil
}

func (s *service) getRandom(key string) (string, error) {
	cl := s.getRandomClient()
	fmt.Println("Random client we are taking is  <><><><><><>", cl)
	n := *s.clients[cl].con
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	val, err := n.Get(ctx, &pb.GetKey{Key: key})
	if err != nil {
		return "", err
	}
	return val.Value, nil
}

func (s *service) getRandomClient() string {
	rand.Seed(time.Now().Unix())
	return s.addresses[rand.Intn(len(s.addresses))]
}
