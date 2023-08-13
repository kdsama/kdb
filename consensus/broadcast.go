package consensus

func (cs *ConsensusService) Broadcast(addresses []string) error {
	// add the node if it doesnot exist already
	cs.logger.Infof("Addresses are %v and the leader sent is %v and my name is %v ", addresses, cs.name)

	for _, addr := range addresses {
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
		if len(cs.addresses) == 1 {
			cs.electMeAndBroadcast()
		} else {
			cs.state = Follower
			cs.askWhoIsTheLeader()
		}
	}

	return nil
}
