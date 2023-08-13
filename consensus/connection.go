package consensus

import (
	pb "github.com/kdsama/kdb/protodata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (cs *ConsensusService) connectClients() {
	// get the list of address from the servers.txt
	// run the below function in different goroutines
	// lets start with just
	// now the other node information will be fetched from the client/discovery service

	// The problem here is we need to setup a leader and do the connection afterwards
	// so it will be better if I put everything in a maddr first
	// and add new ones to the list by doing a cron call every n seconds
	// this needs to be re-written
	// leader should be set here before any connection
	// and each of them should have the information about the leader as well
	// need to sit and think this one through

	cs.active = 0
	for addr, _ := range cs.clients {
		cs.active++

		addr := addr
		val, ok := cs.clients[addr]
		if addr == cs.name {

			continue
		}
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

		cs.checkHeartbeatOnNodes()
	}
}

// connects to the rpc Servers
func connect(addr string) (*pb.ConsensusClient, error) {
	addr += addressPort
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	// defer conn.Close()
	c := pb.NewConsensusClient(conn)

	return &c, nil
}
