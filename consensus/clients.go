package consensus

import (
	"context"
	"time"

	pb "github.com/kdsama/kdb/consensus/protodata"
	"github.com/kdsama/kdb/server/logger"
)

type Client struct {
	name     string
	con      *pb.ConsensusClient
	ticker   time.Ticker
	lastBeat time.Time
	factor   int
	logger   *logger.Logger
}

var clients = map[string]*Client{}

func NewClient(name string, con *pb.ConsensusClient, factor int, logger *logger.Logger) *Client {
	if val, ok := clients[name]; ok {
		return val
	}
	t := time.Duration(factor) * time.Second
	ticker := *time.NewTicker(t)

	c := &Client{name, con, ticker, time.Now(), factor, logger}
	clients[name] = c

	// go c.Schedule()
	return c
}

// scheduling will be done only by the leader
// so we cannot initiate it while creating a client object
func (c *Client) Schedule() {

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

	if time.Since(c.lastBeat) > time.Duration(10*c.factor)*time.Second {
		c.logger.Errorf("Server %v is dead\n", c.name)
		// the server is not responsive
		// changing the ticker timing
		// remove it from the client map
		c.logger.WARNf("%v is dead as no hearbeat received for last 10 requests", c.name)
		delete(clients, c.name)
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	nc := c.con
	r, err := (*nc).Ack(ctx, &pb.Hearbeat{Message: "i"})
	if err != nil {
		c.logger.Errorf("Ouch, No heartbeat from %v\n", c.name, err)
		return
	}
	c.lastBeat = time.Now()
	// c.ticker = *time.NewTicker(time.Duration(c.factor) * time.Second)
	c.logger.Infof("Greeting: from %s %s", c.name, r.GetMessage())
}

func (c *Client) SendRecord(data *[]byte) {
	// we should not send it to dead ones
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	nc := c.con
	r, err := (*nc).SendRecord(ctx, &pb.WalEntry{Entry: *data})
	if err != nil {
		c.logger.Errorf("Ouch, Failed sending data to  %v\n", c.name, err)
		return
	}
	c.logger.Infof("data acknowledged by %v = %v", c.name, r)
}
