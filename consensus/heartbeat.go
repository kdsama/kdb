package consensus

import (
	"os"
	"sync"
	"syscall"
	"time"
)

// SO this is where we are
// we have sending heartbeat and receiving heartbeat scenarion
// when we send we store information about it (lastBeat) from that particular node
// if it exceeds a time, we pronounce it dead
// if we have a lot of dead nodes , we cant commit our data (yet)

// then there is the receiver part
// which actually goes through the service layer
// Technically the nodes should not be having the ticker
// the service should have the ticker
// there should be two tickers, one to send heartbeat and the other one to check if they have received heartbeat or not
// we dont need to have this ticker running for the leader
// but make sure if we become a leader afterwards , we need to destroy the ticker
// so the heartbeats file will have send and receive heartbeats both the parts
// how the functions will be named
// so there is a function to call heartbeat in node file
// the name of that function will be checkHeartbeatOnNodes
// there will be a function on server file called receive heartbeat

// this function is only accessible by the leader. It will connect to new clients and send heartbeat

func (cs *ConsensusService) checkHeartbeatOnNodes() {

	// check for quorum
	if cs.state != Leader {
		return
	}

	if len(cs.clients) == 0 {
		cs.logger.Infof("No clients found but myself so no need to check heartbeat %v  :::: %v", cs.name, cs.clients)
		return
	}

	var (
		count    int
		errCount int
		wg       sync.WaitGroup
		resultCh = make(chan error, len(cs.clients))
	)

	for _, client := range cs.clients {
		if time.Since(client.lastBeat) < 50*time.Millisecond {
			resultCh <- nil
			continue
		}
		client := client
		wg.Add(1)
		if client.delete {
			wg.Done()
			continue
		}

		go func() {
			defer wg.Done()

			err := client.Hearbeat()

			resultCh <- err
		}()
	}

	wg.Wait()
	close(resultCh)

	for err := range resultCh {
		if err != nil {
			errCount++
		} else {
			count++
		}
		if won == quorumOperation(len(cs.addresses), count) {
			return
		}
	}
	cs.logger.Errorf("Leader with a broken Quorum")
	pid := os.Getpid()
	syscall.Kill(pid, syscall.SIGINT)

}

// this is the receiving part of the heartbeat, received by the followers
func (cs *ConsensusService) HeartbeatAck() {
	if cs.state == Leader {
		cs.logger.Errorf("What does this guy think he is , sending leader a heartbeat")
	}

	cs.lastBeat = time.Now()

}

// last heart beat check , this will help us decide the candidacy
func (cs *ConsensusService) lastHeatBeatCheck() {
	if cs.state == Initializing {
		return
	}
	if cs.state == Leader {

		return
	}
	// so what if the term changed before you get to become a leader ??
	// need to compare values
	// or previous leader
	// THis is where some synchronisation is important .
	// How can we do it ?
	// need a better idea for this one
	if time.Since(cs.lastBeat) > 10*time.Second {
		cs.electMeAndBroadcast()
	}

}
