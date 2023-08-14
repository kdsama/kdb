package consensus

import (
	"fmt"
	"time"
)

func (cs *ConsensusService) Broadcast(addresses []string) error {
	// add the node if it doesnot exist already
	fmt.Println("Addresses whike my address is ", addresses)
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
			cs.addresses = append(cs.addresses, addr)
		}
	}
	fmt.Println("Addresses ", cs.addresses)

	if cs.currLeader == "" {
		if len(cs.addresses) == 0 {
			cs.electMeAndBroadcast()
		} else {
			cs.state = Follower
			cs.askWhoIsTheLeader()
		}
	}

	// need to implement a scenario
	// when I become a leader I need to reset the values for all the clients of the lastBeat
	if cs.recTicker == nil {
		cs.recTicker = time.NewTicker(3 * time.Second)
		cs.lastBeat = time.Now()
	}

	if cs.ticker == nil {
		cs.logger.Infof("New normal ticker ")
		cs.ticker = time.NewTicker(5 * time.Second)
	}
	if !cs.init {
		cs.logger.Infof("Scheduling the scheduler")
		cs.init = true
		go cs.Schedule()
	}
	return nil
}
