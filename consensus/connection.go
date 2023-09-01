package consensus

import (
	pb "github.com/kdsama/kdb/protodata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// connect to all the clients
func (cs *ConsensusService) connectClients() {
	for addr, _ := range cs.clients {
		// dont need to add our selves in the list
		if addr == cs.name {
			continue
		}
		addr := addr
		val, ok := cs.clients[addr]

		if ok {
			// has the client layer marked itself to be deleted ?
			if val.delete {
				cs.clientMux.Lock()
				delete(cs.clients, val.name)
				cs.clientMux.Unlock()

			}

		} else {

			conn, err := connect(addr)

			if err != nil {
				cs.logger.Errorf("%v", err)
				cs.clientMux.Lock()
				delete(cs.clients, addr)
				cs.clientMux.Unlock()

				continue

			}
			nc := NewNodes(addr, conn, 7, cs.logger)
			cs.clients[nc.name] = nc
		}
		// we are going to generate heartbeat from the server code instead of nodes.go

	}
	if cs.state == Leader {
		cs.logger.Infof("Leader %s ::Sending Heartbeat", cs.name)
		cs.checkHeartbeatOnNodes()
	}
}

// connects to the rpc Servers
func connect(addr string) (*pb.ConsensusClient, error) {
	// addr += addressPort
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	// defer conn.Close()
	c := pb.NewConsensusClient(conn)

	return &c, nil
}
