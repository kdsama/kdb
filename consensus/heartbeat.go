package consensus

import (
	"sync"
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

	if len(cs.clients) == 1 {
		cs.logger.Infof("No clients found but myself so no need to check heartbeat")
		return
	}
	cs.logger.Infof("Multiple clients are present, we will start checking the heartbeat now ")
	// if cs.active < len(cs.addresses)/2 {
	// 	// might as well
	// 	cs.logger.Warnf("Quorum is broken , we have %d active nodes out of %d", cs.active, len(cs.addresses))
	// 	// no hard action for now
	// }

	var (
		count    int
		errCount int
		quorum   = (len(cs.clients) - 1) / 2
		wg       sync.WaitGroup
		resultCh = make(chan error, len(cs.clients))
	)

	wg.Add(len(cs.clients))

	for _, client := range cs.clients {
		client := client

		if client.name == cs.name || client.delete {
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
	}

	if count < quorum {
		cs.logger.Errorf("Quorum is broken")
	}

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

	if time.Since(cs.lastBeat) > 10*time.Second {
		cs.logger.Errorf("%s died", cs.currLeader)
		cs.electMeAndBroadcast()
	}

}
