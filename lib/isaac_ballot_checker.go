package sebak

import (
	"errors"

	logging "github.com/inconshreveable/log15"

	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/error"
	"boscoin.io/sebak/lib/network"
	"boscoin.io/sebak/lib/node"
)

type BallotChecker struct {
	sebakcommon.DefaultChecker

	Consensus          *ISAAC
	LocalNode          *sebaknode.LocalNode
	NetworkID          []byte
	Message            sebaknetwork.Message
	IsNew              bool
	Ballot             Ballot
	VotingHole         sebakcommon.VotingHole
	RoundVote          *RoundVote
	Result             RoundVoteResult
	VotingFinished     bool
	FinishedVotingHole sebakcommon.VotingHole

	Log logging.Logger
}

// BallotUnmarshal makes `Ballot` from sebaknetwork.Message.
func BallotUnmarshal(c sebakcommon.Checker, args ...interface{}) (err error) {
	checker := c.(*BallotChecker)

	var ballot Ballot
	if ballot, err = NewBallotFromJSON(checker.Message.Data); err != nil {
		return
	}

	if err = ballot.IsWellFormed(checker.NetworkID); err != nil {
		return
	}

	checker.Ballot = ballot
	checker.Log = checker.Log.New(logging.Ctx{"ballot": checker.Ballot.GetHash(), "state": checker.Ballot.State()})
	checker.Log.Debug("message is verified")

	return
}

// BallotNotFromKnownValidators checks the incoming ballot
// is from the known validators.
func BallotNotFromKnownValidators(c sebakcommon.Checker, args ...interface{}) (err error) {
	checker := c.(*BallotChecker)
	if checker.LocalNode.HasValidators(checker.Ballot.Source()) {
		return
	}

	checker.Log.Debug(
		"ballot from unknown validator",
		"from", checker.Ballot.Source(),
	)

	err = sebakerror.ErrorBallotFromUnknownValidator
	return
}

// BallotAlreadyFinished checks the incoming ballot in
// valid round.
func BallotAlreadyFinished(c sebakcommon.Checker, args ...interface{}) (err error) {
	checker := c.(*BallotChecker)

	round := checker.Ballot.Round()
	if !checker.Consensus.IsAvailableRound(round) {
		err = sebakerror.ErrorBallotAlreadyFinished
		checker.Log.Debug("ballot already finished", "round", round)
		return
	}

	return
}

// BallotVote vote by incoming ballot; if the ballot is new
// and the round of ballot is not yet registered, this will make new
// `RunningRound`.
func BallotVote(c sebakcommon.Checker, args ...interface{}) (err error) {

	// err = sebakcommon.RunChecker(checker, nil)
	// if err != nil {
	// 	if stopped, ok := err.(sebakcommon.CheckerStop); ok {
	// 		is.log.Debug(
	// 			"stopped to handle ballot",
	// 			"state", baseChecker.Ballot.State(),
	// 			"reason", stopped.Error(),
	// 		)
	// 	} else {
	// 		is.log.Error("failed to handle ballot", "error", err, "state", baseChecker.Ballot.State())
	// 		return
	// 	}
	return
}

// BallotIsSameProposer checks the incoming ballot has the
// same proposer with the current `RunningRound`.
func BallotIsSameProposer(c sebakcommon.Checker, args ...interface{}) (err error) {
	checker := c.(*BallotChecker)

	if checker.Ballot.Proposer() == checker.LocalNode.Address() {
		return
	}

	var proposer string
	if proposer, err = checker.Consensus.GetVotedProposer(checker.Ballot); err == nil {
		if proposer != checker.Ballot.Proposer() {
			checker.VotingHole = sebakcommon.VotingNO
			checker.Log.Debug(
				"ballot has different proposer",
				"proposer", proposer,
				"proposed-proposer", checker.Ballot.Proposer(),
			)
		}
	}

	return
}

// BallotCheckResult checks the voting result.
func BallotCheckResult(c sebakcommon.Checker, args ...interface{}) (err error) {
	checker := c.(*BallotChecker)

	if !checker.Ballot.State().IsValidForVote() {
		return
	}

	result, votingHole, finished := checker.RoundVote.CanGetVotingResult(
		checker.Consensus.VotingThresholdPolicy,
		checker.Ballot.State(),
	)

	checker.Result = result
	checker.VotingFinished = finished
	checker.FinishedVotingHole = votingHole

	if checker.VotingFinished {
		checker.Log.Debug("get result", "finished VotingHole", checker.FinishedVotingHole, "result", checker.Result)
	}

	return
}

// INITBallotValidateTransactions validates the
// transactions of newly added ballot.
func INITBallotValidateTransactions(c sebakcommon.Checker, args ...interface{}) (err error) {
	checker := c.(*BallotChecker)

	if !checker.IsNew || checker.VotingFinished {
		return
	}

	if checker.RoundVote.IsVotedByNode(checker.Ballot.State(), checker.LocalNode.Address()) {
		err = sebakerror.ErrorBallotAlreadyVoted
		return
	}

	if checker.VotingHole != sebakcommon.VotingNOTYET {
		return
	}

	if checker.Ballot.TransactionsLength() < 1 {
		checker.VotingHole = sebakcommon.VotingYES
		return
	}

	transactionsChecker := &BallotTransactionChecker{
		DefaultChecker: sebakcommon.DefaultChecker{handleBallotTransactionCheckerFuncs},
		Consensus:      checker.Consensus,
		LocalNode:      checker.LocalNode,
		NetworkID:      checker.NetworkID,
		Transactions:   checker.Ballot.Transactions(),
		VotingHole:     sebakcommon.VotingNOTYET,
	}

	err = sebakcommon.RunChecker(transactionsChecker, sebakcommon.DefaultDeferFunc)
	if err != nil {
		if _, ok := err.(sebakcommon.CheckerErrorStop); !ok {
			checker.VotingHole = sebakcommon.VotingNO
			checker.Log.Debug("failed to handle transactions of ballot", "error", err)
			err = nil
			return
		}
		err = nil
	}

	if transactionsChecker.VotingHole == sebakcommon.VotingNO {
		checker.VotingHole = sebakcommon.VotingNO
	} else {
		checker.VotingHole = sebakcommon.VotingYES
	}

	return
}

// SIGNBallotBroadcast will broadcast the validated SIGN ballot.
func SIGNBallotBroadcast(c sebakcommon.Checker, args ...interface{}) (err error) {
	checker := c.(*BallotChecker)
	if !checker.IsNew {
		return
	}

	newBallot := checker.Ballot
	newBallot.SetSource(checker.LocalNode.Address())
	newBallot.SetVote(sebakcommon.BallotStateSIGN, checker.VotingHole)
	newBallot.Sign(checker.LocalNode.Keypair(), checker.NetworkID)

	if found := checker.Consensus.VoteIfRunningRoundExists(newBallot); !found {
		err = errors.New("`RunningRound` not found")
	}

	checker.Consensus.connectionManager.Broadcast(newBallot)
	checker.Log.Debug("ballot will be broadcasted", "newBallot", newBallot)

	return
}

func TransitStateToSIGN(c sebakcommon.Checker, args ...interface{}) (err error) {
	checker := c.(*BallotChecker)
	if !checker.IsNew {
		return
	}
	// [TODO]checker.Consensus.TransitIsaacState(checker.Consensus.IsaacStateManager.State().round, sebakcommon.BallotStateSIGN)

	return
}

// ACCEPTBallotBroadcast will broadcast the confirmed ACCEPT
// ballot.
func ACCEPTBallotBroadcast(c sebakcommon.Checker, args ...interface{}) (err error) {
	checker := c.(*BallotChecker)
	if !checker.VotingFinished {
		return
	}

	newBallot := checker.Ballot
	newBallot.SetSource(checker.LocalNode.Address())
	newBallot.SetVote(sebakcommon.BallotStateACCEPT, checker.FinishedVotingHole)
	newBallot.Sign(checker.LocalNode.Keypair(), checker.NetworkID)

	if found := checker.Consensus.VoteIfRunningRoundExists(newBallot); !found {
		err = errors.New("RunningRound not found")
		return
	}

	checker.Consensus.connectionManager.Broadcast(newBallot)
	checker.Log.Debug("ballot will be broadcasted", "newBallot", newBallot)

	return
}

func TransitStateToACCEPT(c sebakcommon.Checker, args ...interface{}) (err error) {
	checker := c.(*BallotChecker)
	if !checker.VotingFinished {
		return
	}
	// [TODO]checker.NodeRunner.TransitIsaacState(checker.NodeRunner.isaacStateManager.State().round, sebakcommon.BallotStateACCEPT)

	return
}

// FinishedBallotStore will store the confirmed ballot to
// `Block`.
func FinishedBallotStore(c sebakcommon.Checker, args ...interface{}) (err error) {
	checker := c.(*BallotChecker)

	if !checker.VotingFinished {
		return
	}
	if checker.FinishedVotingHole == sebakcommon.VotingYES {
		var block Block
		block, err = FinishBallot(
			checker.Consensus.storage,
			checker.Ballot,
			checker.Consensus.TransactionPool,
		)
		if err != nil {
			return
		}

		checker.Consensus.SetLatestConsensusedBlock(block)
		checker.Log.Debug("ballot was stored", "block", block)

		err = NewCheckerStopCloseConsensus(checker, "ballot got consensus and will be stored")
	} else {
		err = NewCheckerStopCloseConsensus(checker, "ballot got consensus")
	}

	checker.Consensus.CloseConsensus(
		checker.Ballot.Proposer(),
		checker.Ballot.Round(),
		checker.FinishedVotingHole,
	)
	checker.Consensus.SetLatestRound(checker.Ballot.Round())

	// [TODO]nr.isaacStateManager.ResetRound()

	return
}
