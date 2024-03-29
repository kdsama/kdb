package consensus

import (
	"context"
	"time"

	"github.com/kdsama/kdb/config"
	pb "github.com/kdsama/kdb/protodata"
	"go.uber.org/zap"
)

type Node struct {
	name     string              // name of the connected client
	con      *pb.ConsensusClient // connection
	ticker   time.Ticker         //
	lastBeat time.Time
	factor   int
	logger   *zap.SugaredLogger
	delete   bool
	init     bool
}

// new client object
func NewNodes(name string, con *pb.ConsensusClient, factor int, logger *zap.SugaredLogger) *Node {

	var (
		t      = time.Duration(factor) * time.Second
		ticker = *time.NewTicker(t)
	)

	c := &Node{
		name:     name,
		con:      con,
		ticker:   ticker,
		lastBeat: time.Now(),
		factor:   factor,
		logger:   logger,
		delete:   false,
		init:     false,
	}
	return c
}

// Hearbeat to this particular client
func (c *Node) Hearbeat() error {

	c.init = true
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	nc := c.con

	_, err := (*nc).Ack(ctx, &pb.Hearbeat{Message: "i"})
	if err != nil {
		c.logger.Infof("Error %v", err)
		if c.init && time.Since(c.lastBeat) > time.Duration(3*c.factor)*time.Second {
			c.logger.Warnf("%v No Heartbeat ", c.name)
			// this is us informing the service to delete this
			c.delete = true
		}
		return err
	}

	c.lastBeat = time.Now()
	c.delete = false

	return nil
}

// data sent to client (set transaction or set TransactionConfirmation)
func (c *Node) SendRecord(ctx context.Context, data *[]byte, state config.RecordState) error {
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

// client asked for data (get key)
func (c *Node) GetRecord(ctx context.Context, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
	// t := time.Now()

	defer cancel()
	nc := c.con
	res, err := (*nc).Get(ctx, &pb.GetKey{Key: key})
	if err != nil {
		c.logger.Errorf("Ouch, Failed to Get data to %s\n", c.name, err)
		return "", err
	}
	// c.logger.Infof("Time taken in ms ::: %d", time.Since(t).Milliseconds())
	return res.Value, nil
}
