package consensus

import (
	"context"
	"errors"
	"time"

	"github.com/kdsama/kdb/config"
	"github.com/kdsama/kdb/logger"
	pb "github.com/kdsama/kdb/protodata"
)

type Nodes struct {
	leader    bool
	name      string
	con       *pb.ConsensusClient
	ticker    time.Ticker
	lastBeat  time.Time
	factor    int
	logger    *logger.Logger
	delete    bool
	dcCounter int
}

func NewNodes(name string, con *pb.ConsensusClient, factor int, logger *logger.Logger) *Nodes {

	t := time.Duration(factor) * time.Second
	ticker := *time.NewTicker(t)

	c := &Nodes{false, name, con, ticker, time.Now(), factor, logger, false, 0}

	go c.Schedule()
	return c
}

// scheduling will be done only by the leader
// so we cannot initiate it while creating a client object
func (c *Nodes) Schedule() {

	for {
		select {
		case <-c.ticker.C:
			{
				c.HearbeatCheck()
			}
		}
	}
}

func (c *Nodes) HearbeatCheck() {
	if !c.leader {
	}

	// if time.Since(c.lastBeat) > 3*time.Duration(c.factor)*time.Second {
	// 	// c.logger.Errorf("The leader is dead I shall ask other nodes about it, but this layer has no idea about other nodes, so Need to remove ticker from here")
	// }
}

func (c *Nodes) Hearbeat() error {

	if time.Since(c.lastBeat) > time.Duration(10*c.factor)*time.Second {
		c.logger.Errorf("Server %v is dead\n", c.name)
		// the server is not responsive
		// changing the ticker timing
		// remove it from the client map
		c.logger.Warnf("%v is dead as no hearbeat received for last 10 requests", c.name)

		// this is us informing the service to delete this
		c.delete = true
		return errors.New("the node is dead to me")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	nc := c.con

	_, err := (*nc).Ack(ctx, &pb.Hearbeat{Message: "i"})
	if err != nil {
		c.logger.Fatalf("Ouch, No heartbeat from %v because of %v \n", c.name, err)
		return err
	}
	c.lastBeat = time.Now()

	// c.ticker = *time.NewTicker(time.Duration(c.factor) * time.Second)

	return nil
}

func (c *Nodes) SendRecord(ctx context.Context, data *[]byte, state config.RecordState) error {
	// we should not send it to dead ones

	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)

	defer cancel()
	nc := c.con

	_, err := (*nc).SendRecord(ctx, &pb.WalEntry{Entry: *data, Status: int32(state)})

	if err != nil {
		// c.logger.Errorf("Ouch, Failed sending data to  %v\n", c.name, err)
		return err
	}

	// c.logger.Infof("data acknowledged by %v = %v", c.name, r)
	return nil
}

func (c *Nodes) GetRecord(ctx context.Context, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	nc := c.con
	res, err := (*nc).Get(ctx, &pb.GetKey{Key: key})
	if err != nil {
		c.logger.Errorf("Ouch, Failed to Get data to %s\n", c.name, err)
		return "", err
	}

	return res.Value, nil
}
