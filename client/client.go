package clientdiscovery

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/kdsama/kdb/protodata"
	pb "github.com/kdsama/kdb/protodata"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "client_http_requests_total",
			Help: "Total number of requests of different type",
		},
		[]string{"reqtype"},
	)
	requestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "client_request_latency",
		Help:    "requests latency in milliseconds, checked in the service layer, not handler",
		Buckets: []float64{0.0, 2.0, 4.0, 6.0, 8.0, 10.0, 20.0, 40.0, 60.0, 80.0, 100.0, 160.0, 180.0, 200.0, 400.0, 800.0, 1000.0, 2000.0},
	}, []string{"reqtype"})
)

var letters = []rune("abcdefghijklmnopq")

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
	ticker    *time.Ticker
}

var (
	curr     = 0
	mcurr    = 0
	count    = 0
	errCount = 0
)

func NewService() *service {
	ticker := time.NewTicker(2 * time.Second)

	s := &service{
		clients:   map[string]*Nodes{},
		leader:    "",
		addresses: []string{},
		ticker:    ticker,
	}
	go s.schedule()
	return s
}

func (s *service) schedule() {
	for {
		select {
		case <-s.ticker.C:
			s.leaderCheck()
		}
	}
}
func (s *service) leaderCheck() {
	// we give ourselves vote first

	leaderMap := map[string]int{}
	max := -1
	leader := ""

	for key, _ := range s.clients {
		//s.Votefor Me()

		key := key

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		conn := *s.clients[key].con
		ldResponse, err := conn.LeaderInfo(ctx, &protodata.AskLeader{})

		if err == nil {

			leaderMap[ldResponse.Leader]++
			if leaderMap[ldResponse.Leader] > max {
				max = leaderMap[ldResponse.Leader]
				leader = ldResponse.Leader
			}

		}

	}

	s.leader = leader

}

func (s *service) addServer(addr string) error {
	_, ok := s.clients[addr]
	if !ok {
		client, err := connect(addr)
		if err != nil {
			fmt.Println("Connected ???", err)
			return errors.New("cannot connect to the server")

		}

		s.clients[addr] = &Nodes{
			leader: false,
			addr:   addr,
			name:   addr,
			con:    client,
		}
		s.addresses = append(s.addresses, addr)
		if len(s.addresses) == 1 {
			s.leader = addr
		}
		return s.broadcast()

	} else {
		return errors.New("Already present, so not going to broadcast or make it a leader")
	}
	return nil
}
func (s *service) shareAllAddressesWithNewNode(a string) {

	for _, addr := range s.addresses {
		if addr == a {
			continue
		}
		n := *s.clients[a].con
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()
		response, err := n.Broadcast(ctx, &pb.BroadcastNode{Addr: s.addresses})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(response)
	}
}

func (s *service) broadcast() error {
	// broadcast about the incoming server to all the servers
	// broadcast message will include leader name and node added
	// we can verify leader name on each broadcast

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	wg := sync.WaitGroup{}
	var e error
	for _, addr := range s.addresses {
		addr := addr
		n := *s.clients[addr].con
		wg.Add(1)
		go func() {

			defer wg.Done()

			response, err := n.Broadcast(ctx, &pb.BroadcastNode{Addr: s.addresses})
			if err != nil {

				e = err
			}
			fmt.Println(response)
		}()
	}
	wg.Wait()

	return e
}

func (s *service) set(key, val string) error {
	t := time.Now()

	n := *s.clients[s.leader].con
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	_, err := n.Set(ctx, &pb.SetKey{Key: key, Value: val})

	if err != nil {
		requestsTotal.WithLabelValues("SetError").Inc()
		return err
	}
	requestsTotal.WithLabelValues("Set").Inc()
	requestLatency.WithLabelValues("Set").Observe(float64(time.Since(t)) / 1000_000)
	return nil
}

func (s *service) automateSet(duration, requests, sleep string) (int, error) {
	rp, sp := 100000, 500

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
			time.Sleep(time.Duration(sp) * time.Microsecond)

			go func() {
				rand.Seed(time.Now().UnixNano())
				err := s.set("key"+randSeq(rand.Intn(3)), randSeq(rand.Intn(1000)))
				if err != nil {

					errCount++
				} else {

					count++
				}
			}()
		}
		curr = 0
	}()

	return curr, nil
}

func (s *service) automateGet(duration, requests, sleep string) (int, error) {
	rp, sp := 100000, 50

	if mcurr != 0 {
		return -1, errors.New("A request is still getting completed")
	}
	rand.Seed(time.Now().Unix())
	mcurr = rand.Intn(100000)
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

			time.Sleep(time.Duration(sp) * time.Microsecond)

			go func() {

				rand.Seed(time.Now().UnixNano())
				s.getRandom("key" + fmt.Sprint("key"+randSeq(rand.Intn(3))))

			}()
		}

		mcurr = 0
	}()

	return curr, nil
}

func (s *service) get(key string) (string, error) {
	t := time.Now()

	requestsTotal.WithLabelValues("Get").Inc()
	cl := s.leader

	n := *s.clients[cl].con
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	val, err := n.Get(ctx, &pb.GetKey{Key: key})
	requestLatency.WithLabelValues("Get").Observe(float64(time.Since(t)) / 1000_000)
	if err != nil {
		return "", err
	}

	return val.Value, nil
}

func (s *service) getRandom(key string) (string, error) {
	t := time.Now()
	requestsTotal.WithLabelValues("GetRandom").Inc()
	cl := s.getRandomClient()

	n := *s.clients[cl].con
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	val, err := n.Get(ctx, &pb.GetKey{Key: key})
	if err != nil {
		requestsTotal.WithLabelValues("GetRandomError").Inc()
		requestLatency.WithLabelValues("GetRandom").Observe(float64(time.Since(t)) / 1000_000)
		return "", nil
	}
	requestLatency.WithLabelValues("GetRandom").Observe(float64(time.Since(t)) / 1000_000)
	return val.Value, nil
}

func (s *service) getRandomClient() string {
	rand.Seed(time.Now().Unix())
	return s.addresses[rand.Intn(len(s.addresses))]
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

func init() {
	prometheus.MustRegister(requestsTotal, requestLatency)
}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
