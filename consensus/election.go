package consensus

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/kdsama/kdb/protodata"
)

// now we have to talk about election

// so a node will wait for a certain amount of time
// and then send broadcast about its candidacy
// Other nodes would have to acknowledge these two things
// first that they need to check if thats the case with them as well
// if yes , then they have to agree on the voting.
// so in latstHeatbeat check , at a random interval of 150-300 ms I will become a candidate and send the candidacy request to all of the nodes
// If the receiver hasn't voted yet in this term, then it votes for the candidate
// so we have to define a term as well ???
// if it doesnot get a majority of vote, a new term will start
// it will wait again for a random time and send its candidacy
// So if one becomes a leader on majority approval
// That node needs to send a broadcast that it is the new leader
// it can only send broadcast after majority vote has been reached.
// first thing is , we need to have a common election term, throughout all the servers
// once a leader is selected , on the initial process of startup
// we set electionTerm++ and then select a leader
// how to have an election term.
// we should have a separate composite struct for the same ?
// what data-structure should be used or we should use the same struct ?
// and how to do the init leader election
// Lets start at the beginning
// Client will send a broadcast
// that a server is added
// we will check if there is > 1 addresses
// then we will check do we have leader information
// if no leader information + addresses > 1 we put ourselves as candidate and start an election
// election term++
// send leadership request to others
// if others have election term size thats less. we just vote for the guy
// if the election term size is the same + vote has been casted (leader value is different) we send back false, else we just send true
// first we need to make sure we connect all servers with each other.

func (cs *ConsensusService) electMeAndBroadcast() {
	cs.term++
	cs.askForVote()
}

func (cs *ConsensusService) askForVote() {
	// we give ourselves vote first
	voteCount := 1
	wg := sync.WaitGroup{}
	wg.Add(len(cs.clients))

	for key, _ := range cs.clients {
		//cs.Votefor Me()
		key := key
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			conn := *cs.clients[key].con
			conn.Vote(ctx, &protodata.VoteNode{Term: int32(cs.term), Leader: cs.name})
			voteCount++
		}()
	}
	if voteCount > len(cs.clients)/2 {
		// I am the new leader now
		// stop my receiver ticker for heartbeat
		// my ticker for heartbeat is already

		cs.state = Leader
		cs.recTicker.Stop()
	} else if voteCount == len(cs.clients)/2 {
		cs.term++
		rand.Seed(time.Now().UnixMicro())
		time.Sleep(time.Duration(100 + rand.Intn(150)))
		cs.electMeAndBroadcast()

	} else {
		cs.currLeader = "whatever new leader we receive"
		cs.state = Follower
	}
	// now there are two cases here
	// where somebody else have gotten a lot of votes
	// or somebody else has gotten equal votes as you
	// how to figure this out ??
	// lets say we also return the leader in case they are not voting for you .
	// if the count is same as yours , we increase the election term value and do another vote
	// else we make the received value as the new leader
	// need new messages for rpc

}

func (cs *ConsensusService) Vote(term int, leader string) (string, bool) {
	if term > cs.term {
		// we are not persisting the information that who is the leader as of now
		// maybe we will put it on the node side
		// but that means we dont instantly have that information
		// so need to make changes about that
		cs.term = term
		cs.currLeader = leader
		return cs.currLeader, true
	}
	return cs.currLeader, false
}
