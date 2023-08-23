package consensus

type result byte

const (
	won result = iota
	lost
	draw
)

// checking for quorun election, not being used as of now
func quorumElection(total, acquiredVotes int) result {

	// total does not include ourselves
	// so no need to remove it from total count
	var (
		halfVote = float32(total) / 2
		res      = float32(acquiredVotes) - halfVote
	)
	if res > 0 {
		return won
	} else if res == 0 {
		return draw
	}
	return lost

}

// quorum check for Any transaction operation
func quorumOperation(total, acquiredVotes int) result {
	var (
		halfVote = float32(total) / 2
		res      = float32(acquiredVotes) - halfVote
	)
	if res > 0 {
		return won
	} else if res == 0 {
		return draw
	}
	return lost
}
