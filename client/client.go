package client

import (
	"github.com/kdsama/kdb/consensus"
	"github.com/kdsama/kdb/server/logger"
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
	go c.cs.SendTransaction(dat)

	return nil

}
