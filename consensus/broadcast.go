package consensus

import "time"

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
			cs.addresses = append(cs.addresses, addr)
		}
	}

	if cs.currLeader == "" {
		if len(cs.addresses) == 0 {
			cs.electMeAndBroadcast()
		} else {
			cs.state = Follower
			cs.askWhoIsTheLeader()
		}
	}
	if cs.state == Leader {
		if cs.recTicker != nil {
			cs.recTicker.Stop()
		}
	} else {
		if cs.recTicker == nil {
			cs.recTicker = time.NewTicker(3 * time.Second)
		}
	}
	if cs.ticker == nil {
		cs.ticker = time.NewTicker(5 * time.Second)
	}
	if !cs.init {
		cs.init = true
		go cs.Schedule()
	}
	return nil
}
