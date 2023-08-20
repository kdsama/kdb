package consensus

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/kdsama/kdb/protodata"
)

// request for election . If the node is the only standing one, it elects itself as a leader
func (cs *ConsensusService) requestElection() {
	if len(cs.clients) == 0 {
		cs.state = Leader
		cs.term++
		cs.currLeader = cs.name
		return
	}
	cs.askForVote()

}

// election process, with a  random timeOut.
func (cs *ConsensusService) askForVote() {
	// Lock so it doesnot collide with a voting request that might be received at the same time.
	rand.Seed(time.Now().UnixMicro())
	time.Sleep(time.Duration(50+rand.Intn(150)) * time.Millisecond)
	cs.clientMux.Lock()
	cs.state = Candidate
	cs.currLeader = cs.name
	cs.term++
	cs.clientMux.Unlock()
	var (
		wg   = sync.WaitGroup{}
		done = false
		term = cs.term
		// leader string
	)
	count := 0 // we dont give ourselves vote as we dont have ourselves in the client list
	for key, _ := range cs.clients {
		wg.Add(1)
		key := key
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			conn := *cs.clients[key].con

			// Vote Request through RPC
			r, err := conn.Vote(ctx, &protodata.VoteNode{Term: int32(term), Leader: cs.currLeader, Votes: []string{}})
			cs.logger.Infof("%v this is the vote response", r)
			if err != nil || !r.Status {
				return
			}
			cs.clientMux.Lock()
			defer cs.clientMux.Unlock()
			count++

			if done || count < len(cs.clients)/2 {
				return
			}

			done = true
			// Once we are not a Candidate or the termState is different from when this function was executed means either we are leader or followe
			// but not a candidate anymore
			if cs.state != Candidate || cs.term != term {
				return
			}

			//  Leader state confirmed
			cs.state = Leader
			cs.currLeader = cs.name
		}()
	}
	wg.Wait()

}

func (cs *ConsensusService) Vote(term int, leader string, votes []string) (string, bool) {

	cs.clientMux.Lock()
	defer cs.clientMux.Unlock()
	if term > cs.term {
		cs.state = Follower
		cs.lastBeat = time.Now()
		cs.currLeader = leader
		return leader, true
	}
	return leader, false
}

func (cs *ConsensusService) LeaderInfo() (string, error) {
	if cs.currLeader == "" {
		return "", errors.New("there is not leader ")
	}
	return cs.currLeader, nil
}

func (cs *ConsensusService) askWhoIsTheLeader() {
	// we give ourselves vote first
	if cs.state == Leader {
		return
	}
	var (
		wg        = sync.WaitGroup{}
		leaderMap = map[string]int{}
		max       = -1
		leader    = ""
	)

	for key, _ := range cs.clients {
		wg.Add(1)
		key := key
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			conn := *cs.clients[key].con
			ldResponse, err := conn.LeaderInfo(ctx, &protodata.AskLeader{})

			if err == nil {

				cs.clientMux.Lock()
				defer cs.clientMux.Unlock()
				leaderMap[ldResponse.Leader]++
				if leaderMap[ldResponse.Leader] > max {
					max = leaderMap[ldResponse.Leader]
					leader = ldResponse.Leader
				}

			}

		}()
	}

	wg.Wait()
	// this will get us the latest leader
	cs.currLeader = leader
	cs.logger.Infof("Leader is %s", cs.currLeader)
	cs.state = Follower

}
