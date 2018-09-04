package sebak

import (
	"errors"
	"sort"

	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/network"
	"boscoin.io/sebak/lib/node"
	"boscoin.io/sebak/lib/round"
	logging "github.com/inconshreveable/log15"
)

type ISAAC struct {
	sebakcommon.SafeLock

	NetworkID             []byte
	Node                  *sebaknode.LocalNode
	VotingThresholdPolicy sebakcommon.VotingThresholdPolicy
	TransactionPool       *TransactionPool
	runningRounds         map[ /* Round.Hash() */ string]*RunningRound
	LatestConfirmedBlock  Block
	LatestRound           round.Round
	proposerCalculator    ProposerCalculator
	connectionManager     *sebaknetwork.ConnectionManager
	log                   logging.Logger
}

func NewISAAC(networkID []byte, node *sebaknode.LocalNode, votingThresholdPolicy sebakcommon.VotingThresholdPolicy) (is *ISAAC, err error) {
	is = &ISAAC{
		NetworkID: networkID,
		Node:      node,
		VotingThresholdPolicy: votingThresholdPolicy,
		TransactionPool:       NewTransactionPool(),
		runningRounds:         map[string]*RunningRound{},
	}
	is.log = log.New(logging.Ctx{"isaac": is.Node.Alias()})

	return
}

func (is *ISAAC) CloseConsensus(proposer string, round round.Round, vh sebakcommon.VotingHole) (err error) {
	is.Lock()
	defer is.Unlock()

	if vh == sebakcommon.VotingNOTYET {
		err = errors.New("invalid VotingHole, `VotingNOTYET`")
		return
	}

	roundHash := round.Hash()
	rr, found := is.runningRounds[roundHash]
	if !found {
		return
	}

	if vh == sebakcommon.VotingNO {
		delete(rr.Transactions, proposer)
		delete(rr.Voted, proposer)

		return
	}

	is.TransactionPool.Remove(rr.Transactions[proposer]...)

	delete(is.runningRounds, roundHash)

	// remove all the same rounds
	for hash, runningRound := range is.runningRounds {
		if runningRound.Round.BlockHeight > round.BlockHeight {
			continue
		}
		delete(is.runningRounds, hash)
	}

	return
}

func (is *ISAAC) SetConnectionManager(cm *sebaknetwork.ConnectionManager) {
	is.connectionManager = cm
}

func (is *ISAAC) SetLatestConsensusedBlock(block Block) {
	is.LatestConfirmedBlock = block
}

func (is *ISAAC) SetLatestRound(round round.Round) {
	is.LatestRound = round
}

func (is *ISAAC) IsAvailableRound(round round.Round) bool {
	// check current round is from InitRound
	if is.LatestRound.BlockHash == "" {
		return true
	}

	if round.BlockHeight < is.LatestConfirmedBlock.Height {
		return false
	} else if round.BlockHeight == is.LatestConfirmedBlock.Height {
		if round.BlockHash != is.LatestConfirmedBlock.Hash {
			return false
		}
	} else {
		// TODO if incoming round.BlockHeight is bigger than
		// LatestConfirmedBlock.Height and this round confirmed successfully,
		// this node will get into catchup status
	}

	if round.BlockHeight == is.LatestRound.BlockHeight {
		if round.Number <= is.LatestRound.Number {
			return false
		}
	}

	return true
}

func (is *ISAAC) IsVoted(ballot Ballot) bool {
	rr := is.runningRounds

	var found bool
	var runningRound *RunningRound
	if runningRound, found = rr[ballot.Round().Hash()]; !found {
		return false
	}

	return runningRound.IsVoted(ballot)
}

func (is *ISAAC) GetVotedProposer(ballot Ballot) (string, error) {
	rr := is.runningRounds
	var runningRound *RunningRound
	var found bool
	if runningRound, found = rr[ballot.Round().Hash()]; !found {
		err := errors.New("`RunningRound` not found")
		return "", err
	}
	return runningRound.Proposer, nil
}

func (is *ISAAC) VoteIfRunningRoundExists(ballot Ballot) (found bool) {
	if !ballot.State().IsValidForVote() {
		return
	}
	rr := is.runningRounds

	var runningRound *RunningRound
	if runningRound, found = rr[ballot.Round().Hash()]; !found {
		return
	}

	runningRound.Vote(ballot, is.VotingThresholdPolicy)

	return
}

func (is *ISAAC) Vote(ballot Ballot) (isNew bool, err error) {
	if !ballot.State().IsValidForVote() {
		return
	}
	roundHash := ballot.Round().Hash()
	rr := is.runningRounds

	var found bool
	var runningRound *RunningRound
	if runningRound, found = rr[roundHash]; !found {
		proposer := is.CalculateProposer(
			ballot.Round().BlockHeight,
			ballot.Round().Number,
		)

		runningRound, err = NewRunningRound(proposer, ballot)
		if err != nil {
			return
		}

		rr[roundHash] = runningRound
		isNew = true
	} else {
		if _, found = runningRound.Voted[ballot.Proposer()]; !found {
			isNew = true
		}

		runningRound.Vote(ballot, is.VotingThresholdPolicy)
	}

	return
}

func (is *ISAAC) SetProposerCalculator(c ProposerCalculator) {
	is.proposerCalculator = c
}

func (is *ISAAC) CalculateProposer(blockHeight uint64, roundNumber uint64) string {
	return is.proposerCalculator.Calculate(is.connectionManager, blockHeight, roundNumber)
}

type ProposerCalculator interface {
	Calculate(cm *sebaknetwork.ConnectionManager, blockHeight uint64, roundNumber uint64) string
}

type SimpleProposerCalculator struct{}

func (c SimpleProposerCalculator) Calculate(cm *sebaknetwork.ConnectionManager, blockHeight uint64, roundNumber uint64) string {
	candidates := sort.StringSlice(cm.AllValidators())
	candidates.Sort()

	return candidates[(blockHeight+roundNumber)%uint64(len(candidates))]
}
