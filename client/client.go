package client

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
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path"},
	)
)

type Client struct {
	kv      *store.KVService
	cs      *consensus.ConsensusService
	logger  *logger.Logger
	userMap map[string]string
}

func New(kv *store.KVService, cs *consensus.ConsensusService, logger *logger.Logger) *Client {

	return &Client{
		kv:      kv,
		cs:      cs,
		logger:  logger,
		userMap: map[string]string{},
	}
}

func (c *Client) Add(key, value string) error {
	requestsTotal.WithLabelValues("Type", "Setter").Inc()

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

func (c *Client) Get(user string, key string) (string, error) {
	requestsTotal.WithLabelValues("Type", "Getter").Inc()
	// here what we can do is
	// set a client connection to a particular client for the current user for reading purposes
	// so n users will have a random client attached to it to fetch the get requests
	// for now I can just ask the system to get me data through gcp ,
	// so need protobuf here for get services

	if _, ok := c.userMap[user]; !ok {
		n := c.cs.GetRandomConnectionName()
		fmt.Println("Random connection is ", n)
		c.userMap[user] = n
	}

	return c.cs.Get(c.userMap[user], key)

}

func (c *Client) AutomateGet() error {
	time.Sleep(21 * time.Second)
	for i := 0; i < 100000; i++ {
		time.Sleep(1 * time.Millisecond)

		_, err := c.Get("user", "key"+fmt.Sprint(rand.Int31n(100)))
		if err != nil {
			c.logger.Errorf("%v", err)
		}

	}
	return nil
}

func (c *Client) AutomateSet() error {
	time.Sleep(20 * time.Second)

	go c.BulkAdd("val")
	// the below one is to check if
	// we are getting another line in the persistance layer
	go c.BulkAdd("preval")
	return nil
}

func (c *Client) BulkAdd(value string) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 10000; i++ {
		time.Sleep(10 * time.Millisecond)

		err := c.Add("key"+fmt.Sprint(rand.Int31n(100)), fmt.Sprint(rand.Int31n(10000)))
		if err != nil {
			fmt.Println(err)
		}

	}
}

func init() {
	fmt.Println("Is this happening or not ???????")
	prometheus.MustRegister(requestsTotal)
}
