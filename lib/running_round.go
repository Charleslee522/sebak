package sebak

import (
	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/error"
	"boscoin.io/sebak/lib/round"
	logging "github.com/inconshreveable/log15"
)

type RunningRound struct {
	sebakcommon.SafeLock

	Round        round.Round
	Proposer     string                              // LocalNode's `Proposer`
	Transactions map[ /* Proposer */ string][]string /* Transaction.Hash */
	Voted        map[ /* Proposer */ string]*RoundVote
	log          logging.Logger
}

func NewRunningRound(proposer string, ballot Ballot) (*RunningRound, error) {
	transactions := map[string][]string{
		ballot.Proposer(): ballot.Transactions(),
	}

	roundVote := NewRoundVote(ballot)
	voted := map[string]*RoundVote{
		ballot.Proposer(): roundVote,
	}
	return &RunningRound{
		Round:        ballot.Round(),
		Proposer:     proposer,
		Transactions: transactions,
		Voted:        voted,
		log:          log.New(logging.Ctx{"RunningRound": ballot.Round()}),
	}, nil
}

func (rr *RunningRound) RoundVote(proposer string) (rv *RoundVote, err error) {
	var found bool
	rv, found = rr.Voted[proposer]
	if !found {
		err = sebakerror.ErrorRoundVoteNotFound
		return
	}
	return
}

func (rr *RunningRound) IsVoted(ballot Ballot) bool {
	roundVote, err := rr.RoundVote(ballot.Proposer())
	if err != nil {
		return false
	}

	return roundVote.IsVoted(ballot)
}

func (rr *RunningRound) Vote(ballot Ballot, votingThresholdPolicy sebakcommon.VotingThresholdPolicy) {
	rr.Lock()
	defer rr.Unlock()

	if _, found := rr.Voted[ballot.Proposer()]; !found {
		rr.Voted[ballot.Proposer()] = NewRoundVote(ballot)
	} else {
		rr.Voted[ballot.Proposer()].Vote(ballot)
	}
	roundVote := rr.Voted[ballot.Proposer()]

	result, votingHole, finished := roundVote.CanGetVotingResult(
		votingThresholdPolicy,
		ballot.State(),
	)

	if finished {
		rr.log.Debug("get result", "finished VotingHole", votingHole, "result", result)
	}

	return
}
