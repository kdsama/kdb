package client

import (
	"fmt"
	"time"

	"github.com/kdsama/kdb/consensus"
	"github.com/kdsama/kdb/logger"
	"github.com/kdsama/kdb/server/store"
)

type Client struct {
	kv     *store.KVService
	cs     *consensus.ConsensusService
	logger *logger.Logger
}

func New(kv *store.KVService, cs *consensus.ConsensusService, logger *logger.Logger) *Client {
	return &Client{
		kv:     kv,
		cs:     cs,
		logger: logger,
	}
}

func (c *Client) Add(key, value string) error {
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
	c.cs.SendTransaction(dat, entry.TxnID)

	return nil

}

func (c *Client) AutomateGet(key string) error {
	// this should be a bit different
	// we should not getting data from a random client thats not us
	// we should add delay to it as well probably
	// and we should be doing bulk reads

	return nil
}

func (c *Client) AutomateSet() error {
	time.Sleep(20 * time.Second)

	go c.BulkAdd("val")
	// the below one is to check if
	// we are getting another line in the persistance layer
	c.BulkAdd("pre")
	return nil
}

func (c *Client) BulkAdd(value string) {
	for i := 0; i < 10000; i++ {

		i := i
		time.Sleep(1 * time.Millisecond)

		c.Add("key"+fmt.Sprint(i), value)

	}
}
