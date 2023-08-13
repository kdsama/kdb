package consensus

import "fmt"

type result byte

const (
	won result = iota
	lost
	draw
)

func quorumElection(total, acquiredVotes int) result {

	// so here we know that its an election
	// so our leader is down.
	// remove that from leader
	// remove yourself as well
	// and then compare with acquired Votes

	total -= 1
	halfVote := float32(total) / 2
	fmt.Println(acquiredVotes, halfVote)
	res := float32(acquiredVotes) - halfVote
	if res > 0 {
		return won
	} else if res == 0 {
		return draw
	}
	return lost

}

func quorumOperation(total, acquiredVotes int) result {
	halfVote := float32(total) / 2
	res := float32(acquiredVotes) - halfVote
	if res > 0 {
		return won
	} else if res == 0 {
		return draw
	}
	return lost
}
