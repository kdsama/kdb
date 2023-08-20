package consensus

import (
	"time"
)

// broadcas
func (cs *ConsensusService) Broadcast(addresses []string) error {
	// add the node if it doesnot exist already
	for _, addr := range addresses {
		if addr == cs.name {
			continue
		}
		if _, ok := cs.clients[addr]; !ok {

			client, err := connect(addr)
			if err != nil {
				cs.logger.Fatalf("%v", err)
				return err
			}
			cs.clients[addr] = NewNodes(addr, client, 3, cs.logger)

		}
	}

	if cs.currLeader == "" {
		if len(cs.clients) == 0 {
			cs.requestElection()
		} else {

			cs.askWhoIsTheLeader()
		}
	}

	if cs.recTicker == nil {
		cs.recTicker = time.NewTicker(700 * time.Millisecond)
		cs.lastBeat = time.Now()
	}

	if cs.ticker == nil {
		cs.logger.Infof("New normal ticker ")
		cs.ticker = time.NewTicker(1 * time.Second)
	}
	if !cs.init {
		cs.logger.Infof("Scheduling the scheduler")
		cs.init = true
		go cs.Schedule()
	}
	return nil
}
