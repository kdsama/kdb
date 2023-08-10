package clientdiscovery

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/kdsama/kdb/consensus"
	"github.com/kdsama/kdb/logger"
	"github.com/kdsama/kdb/server/store"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ps_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"reqtype"},
	)
)

// all of this should be done in the consensus folder for now
// no time to create another layer
// we will call these functions using gRPC
// that means The task of renaming files have started as well.
// maybe I would still need to create another layer
// it should not be inside the consensus but it should call the consensus
// So a Receiver package that communicates with the ClientDiscovery
// How the restructuring of the the server folder wil go ?
// so there is a sender (leader) and there is receivers
// each receiver receives stuff . I have named it ClientDiscovery
// a better name would be a connection or a serverNodes
// so selfNode will send data to other nodes
// so we can rename ourself as selfNodes
// and others as Nodes ?
// selfNode is a leader or not ? can name it like that ?
// Still need a better name for it
// for the receiver file, it has rpc receiver functions
// What should be its name ?
// so it can be considered as functions for rpcHandlers ?

type ClientDiscovery struct {
	kv      *store.KVService
	cs      *consensus.ConsensusService
	logger  *logger.Logger
	userMap map[string]string
}

func New(kv *store.KVService, cs *consensus.ConsensusService, logger *logger.Logger) *ClientDiscovery {

	return &ClientDiscovery{
		kv:      kv,
		cs:      cs,
		logger:  logger,
		userMap: map[string]string{},
	}
}

func (c *ClientDiscovery) Add(key, value string) error {
	requestsTotal.WithLabelValues("Set Key").Inc()

	entry, err := c.kv.Add(key, value)
	if err != nil {
		c.logger.Errorf("%v", err)
		return err
	}

	// when we get the entry we should send the entry to the consensus service
	// should we implement spinning locks for this ?
	// what will be the criteria for that ?
	dat, err := c.kv.SerializeRecord(&entry)
	if err != nil {
		c.logger.Errorf("%v", err)
		return err
	}
	err = c.cs.SendTransaction(dat, entry.TxnID)
	if err != nil {
		c.logger.Fatalf("%v", err)
		return err
	}
	err = c.kv.SetRecord(&dat)
	if err != nil {
		c.logger.Fatalf("%v", err)
	}
	return c.cs.SendTransactionConfirmation(dat, entry.TxnID, consensus.Commit)

}

func (c *ClientDiscovery) Get(user string, key string) (string, error) {
	requestsTotal.WithLabelValues("Get Key").Inc()
	// here what we can do is
	// set a ClientDiscovery connection to a particular ClientDiscovery for the current user for reading purposes
	// so n users will have a random ClientDiscovery attached to it to fetch the get requests
	// for now I can just ask the system to get me data through gcp ,
	// so need protobuf here for get services

	if _, ok := c.userMap[user]; !ok {

		n := c.cs.GetRandomConnectionName()
		fmt.Println("Random connection is ", n)
		c.userMap[user] = n
	}

	return c.cs.Get(c.userMap[user], key)

}

func (c *ClientDiscovery) AutomateGet() {
	time.Sleep(50 * time.Second)
	for {
		// time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		time.Sleep(10 * time.Millisecond)
		_, err := c.Get("user"+fmt.Sprint(rand.Int31n(100)), "key"+fmt.Sprint(rand.Int31n(10000)))
		if err != nil {
			c.logger.Errorf("%v", err)
		}

	}

}

func (c *ClientDiscovery) AutomateSet() {
	time.Sleep(20 * time.Second)
	c.BulkAdd("val")
}

func (c *ClientDiscovery) BulkAdd(value string) {
	rand.Seed(time.Now().UnixNano())
	for {
		// time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		time.Sleep(10 * time.Millisecond)
		err := c.Add("key"+fmt.Sprint(rand.Int31n(10000)), fmt.Sprint(rand.Int31n(10000)))
		if err != nil {
			fmt.Println(err)
		}

	}
}

func init() {
	prometheus.MustRegister(requestsTotal)
}
