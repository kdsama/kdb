package consensus

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/kdsama/kdb/protodata"
)

type Term struct {
	ID     int32
	Leader string
	Votes  []string
}

var terms = []Term{}

func NewTerm(id int32, leader string) (*Term, error) {

	t := &Term{
		ID:     id,
		Leader: leader,
		Votes:  []string{leader},
	}
	// check for existing term
	if len(terms) == 1 {
		if t.ID == terms[0].ID {
			// there is already an election that has started
			// dont share the vote in this case
			return nil, errors.New("no broadcast")
		}
	}
	return t, nil
}

func (cs *ConsensusService) electMeAndBroadcast() {

	// need a way to return from here or not call this functon
	if len(cs.addresses) == 0 {
		cs.currLeader = cs.name
		cs.state = Leader
		t, _ := NewTerm(int32(0), cs.name)
		cs.term = t

		return
		// dont do nothing
		// just return
		// you are not going to ask for vote from nobody
	}
	rand.Seed(time.Now().UnixMicro())
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)

	id := cs.term.ID
	id++
	term, err := NewTerm(id, cs.name)
	if err != nil {
		// no need to broadcast yourself
		return
	}
	terms = append(terms, *term)
	cs.term = term
	cs.logger.Infof("Term %d", cs.term.ID)
	cs.askForVote()
}

func (cs *ConsensusService) askForVote() {
	// we give ourselves vote first

	var (
		wg = sync.WaitGroup{}
		// leader string
	)

	for key, _ := range cs.clients {
		//cs.Votefor Me()
		if key == cs.currLeader {
			continue
		}
		wg.Add(1)
		key := key
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			conn := *cs.clients[key].con

			r, err := conn.Vote(ctx, &protodata.VoteNode{Term: cs.term.ID, Leader: cs.term.Leader, Votes: cs.term.Votes})
			if err != nil {
				cs.logger.Infof("This is what it is %v", err)
				return

			}
			if r.Leader == cs.name {
				terms[0].Votes = append(terms[0].Votes, key)
			}
			// now here we will put the logic
			// we will return from vote acknowledgement the other peoples response
			// They will say Yes or NO
			// But first what I need to say is, if the election is a new one or an old one
			// if it is a new one , just tell the guy they are the leader
			// else if election is not a new one, we respond with term already started, restart a new one.
			// But if you send me who has elected themselves as well, we both should be able to compare the number of votes we have to start a re-election
			// Now comes the dead node part. Lets say we have 4 nodes
			// 1 dies
			// 2 of them go for an election for a new term
			// One will reach the 3rd node later
			// once its request is sent the 3rd node will say I already have made this new guy a leader
			// 2nd node will check its quorum . Out of 4 nodes , he has 1 vote. It will make itself a follower
			// It will broadcast LeaderInformation and make the new leader his own leader
			// and then convert himself to the follower
			// Once this happens , once we have a new leader , we can delete our latest term
			// If they say no , they can mention who they voted for
			// They can also share who else voted for that leader if they have that information
			// WHat if both have the value one ?
		}()
	}

	// implement a channel so that , I can check for won scenario proactively
	//
	result := quorumElection(len(cs.clients), len(terms[0].Votes))
	switch result {
	case won:
		cs.state = Leader
		cs.currLeader = cs.name
		cs.voteTime = time.Now()

	case lost:
		cs.logger.Infof("We are going to broadcast and ask who is the leader as we lost this one bitch")
		cs.askWhoIsTheLeader()
	case draw:
		terms = []Term{}
		cs.electMeAndBroadcast()
	}

	terms = []Term{}
}

func (cs *ConsensusService) Vote(term int, leader string, votes []string) (string, bool) {
	// now here we will put the logic
	// we will return from vote acknowledgement the other peoples response
	// They will say Yes or NO
	// But first what I need to say is, if the election is a new one or an old one
	// if it is a new one , just tell the guy they are the leader
	// else if election is not a new one, we respond with term already started, restart a new one.
	// But if you send me who has elected themselves as well, we both should be able to compare the number of votes we have to start a re-election
	// Now comes the dead node part. Lets say we have 4 nodes
	// 1 dies
	// 2 of them go for an election for a new term
	// One will reach the 3rd node later
	// once its request is sent the 3rd node will say I already have made this new guy a leader
	// 2nd node will check its quorum . Out of 4 nodes , he has 1 vote. It will make itself a follower
	// It will broadcast LeaderInformation and make the new leader his own leader
	// and then convert himself to the follower
	// Once this happens , once we have a new leader , we can delete our latest term
	// If they say no , they can mention who they voted for
	// They can also share who else voted for that leader if they have that information
	// WHat if both have the value one ?

	// check if currentTerm < asked term
	if cs.term.ID < int32(term) {
		cs.currLeader = leader
		t, _ := NewTerm(int32(term), leader)

		terms = []Term{}
		terms = append(terms, *t)

		return cs.currLeader, true

	}
	return cs.currLeader, false
}

func (cs *ConsensusService) LeaderInfo() (string, error) {
	if cs.currLeader == "" {
		return "", errors.New("there is not leader ")
	}
	return cs.currLeader, nil
}

func (cs *ConsensusService) askWhoIsTheLeader() {
	// we give ourselves vote first

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
				cs.logger.Infof("leader is %s", ldResponse.Leader)
				cs.clientMux.Lock()

				leaderMap[ldResponse.Leader]++
				if leaderMap[ldResponse.Leader] > max {
					max = leaderMap[ldResponse.Leader]
					leader = ldResponse.Leader
				}
				cs.clientMux.Unlock()

			}

		}()
	}

	wg.Wait()

	if leader == "" || leaderMap[leader] < len(cs.clients)/2 {
		cs.logger.Infof("Leader %s, leaderMap Count %d , total clients %d", leader, leaderMap[leader], len(cs.clients))
		cs.logger.Fatalf("WTF this is not what I was expecting")
	}

	cs.currLeader = leader
	cs.voteTime = time.Now()
	if cs.term == nil {
		t, _ := NewTerm(int32(0), leader)
		cs.term = t
	}

}
