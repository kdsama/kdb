package consensus

func (cs *ConsensusService) Broadcast(addresses []string, leader string) error {
	// add the node if it doesnot exist already

	for _, addr := range addresses {
		if _, ok := cs.clients[addr]; !ok {

			client, err := connect(addr)
			if err != nil {
				return err
			}
			cs.clients[addr] = NewNodes(addr, client, 3, cs.logger)
			cs.addresses = append(cs.addresses, addr)
		}
	}
	if cs.name != leader {
		cs.state = Follower
	}
	if cs.currLeader == "" {
		if len(cs.addresses) == 1 {
			cs.electMeAndBroadcast()
		} else {
			cs.askWhoIsTheLeader()
		}
	}

	return nil
}
