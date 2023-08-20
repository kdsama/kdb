package consensus

import (
	"os"
	"sync"
	"syscall"
	"time"
)

// Leader sending heartbeats to all the client. Only to be invoked by the leader in the cluster
func (cs *ConsensusService) checkHeartbeatOnNodes() {

	// check for quorum
	if cs.state != Leader {
		return
	}
	cs.logger.Infof("Checking heartbeat")
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
		if won == quorumOperation(len(cs.clients), count) {
			return
		}
	}
	cs.logger.Errorf("Leader with a broken Quorum")
	pid := os.Getpid()
	syscall.Kill(pid, syscall.SIGINT)

}

// HeartBeat Acknowledgement.
func (cs *ConsensusService) HeartbeatAck() {
	if cs.state == Leader {
		cs.logger.Errorf("What does this guy think he is , sending leader a heartbeat")
	}
	cs.state = Follower
	cs.lastBeat = time.Now()
	cs.logger.Infof("Received heartbeat from %s", cs.currLeader)
}

// last heart beat check , this will help us decide the candidacy
func (cs *ConsensusService) lastHeatBeatCheck() {
	if cs.state == Initializing {
		return
	}
	if cs.state == Leader {
		return
	}

	if time.Since(cs.lastBeat) > 3*time.Second {
		cs.logger.Infof("Request election ")
		cs.requestElection()
	}

}
