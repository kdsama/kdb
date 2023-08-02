package consensus

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/kdsama/kdb/consensus/protodata"
	"github.com/kdsama/kdb/server/logger"
)

type Client struct {
	name     string
	con      *pb.ConsensusClient
	leader   bool
	ticker   time.Ticker
	lastBeat time.Time
	factor   int
	logger   *logger.Logger
}

var clients = map[string]*Client{}

func NewClient(name string, con *pb.ConsensusClient, factor int, logger *logger.Logger) *Client {
	t := time.Duration(factor) * time.Second
	ticker := *time.NewTicker(t)

	c := &Client{name, con, false, ticker, time.Now(), factor, logger}
	clients[name] = c

	// go c.Schedule()
	return c
}

// scheduling will be done only by the leader
// so we cannot initiate it while creating a client object
func (c *Client) Schedule() {
	c.logger.Infof("Scheduler started for server ::", c.name)
	for {
		select {
		case <-c.ticker.C:
			{
				c.Hearbeat()
			}
		}
	}
}

func (c *Client) Hearbeat() {

	if time.Since(c.lastBeat) > time.Duration(3*c.factor)*time.Second {
		c.factor = 3 * c.factor
		// the server is not responsive
		// changing the ticker timing
		t := time.Duration(3*c.factor) * time.Second
		c.logger.Infof("duration changed for server %v, now heartbeats will be sent at an interval of %v seconds", c.name, c.factor)
		c.ticker = *time.NewTicker(t)
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	nc := c.con
	r, err := (*nc).Ack(ctx, &pb.Hearbeat{Message: (fmt.Sprint(time.Now()))})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: from %s %s", c.name, r.GetMessage())
}
