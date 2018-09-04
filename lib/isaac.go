package sebak

import (
	"errors"
	"sort"

	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/error"
	"boscoin.io/sebak/lib/network"
	"boscoin.io/sebak/lib/node"
	"boscoin.io/sebak/lib/round"
	"boscoin.io/sebak/lib/storage"
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
	storage               *sebakstorage.LevelDBBackend

	BaseBallotCheckerFuncs []sebakcommon.CheckerFunc

	log logging.Logger
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

var BaseBallotCheckerFuncs = []sebakcommon.CheckerFunc{
	BallotUnmarshal,
	BallotNotFromKnownValidators,
	BallotAlreadyFinished,
}

var handleBallotTransactionCheckerFuncs = []sebakcommon.CheckerFunc{
	IsNew,
	GetMissingTransaction,
	BallotTransactionsSameSource,
	BallotTransactionsSourceCheck,
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

func (is *ISAAC) SetStorage(db *sebakstorage.LevelDBBackend) {
	is.storage = db
}

func (is *ISAAC) SetLatestConsensusedBlock(block Block) {
	is.LatestConfirmedBlock = block
}

func (is *ISAAC) SetLatestRound(round round.Round) {
	is.LatestRound = round
}

func (is *ISAAC) handleBallotMessage(message sebaknetwork.Message) (err error) {
	is.log.Debug("got ballot", "message", message.Head(50))

	baseChecker := &BallotChecker{
		DefaultChecker: sebakcommon.DefaultChecker{BaseBallotCheckerFuncs},
		Consensus:      is,
		LocalNode:      is.Node,
		NetworkID:      is.NetworkID,
		Message:        message,
		Log:            is.log,
		VotingHole:     sebakcommon.VotingNOTYET,
	}
	err = sebakcommon.RunChecker(baseChecker, nil)
	if err != nil {
		if _, ok := err.(sebakcommon.CheckerErrorStop); !ok {
			is.log.Error("failed to handle ballot", "error", err, "state", "base")
			return
		}
	}

	ballot := baseChecker.Ballot
	err = is.handleBallot(ballot)
	return
}

func (is *ISAAC) handleBallot(ballot Ballot) (err error) {
	{
		if is.IsVoted(ballot) {
			err = sebakerror.ErrorBallotAlreadyVoted
			return
		}

		var isNew bool
		if isNew, err = is.Vote(ballot); err != nil {
			err = errors.New("`RunningRound` not found")
		}

		var votingResult sebakcommon.VotingHole = sebakcommon.VotingNOTYET
		if ballot.Proposer() != is.Node.Address() {
			var proposer string
			if proposer, err = is.GetVotedProposer(ballot); err == nil {
				if proposer != ballot.Proposer() {
					votingResult = sebakcommon.VotingNO
					is.log.Debug(
						"ballot has different proposer",
						"proposer", proposer,
						"proposed-proposer", ballot.Proposer(),
					)
				}
			}
		}

		switch ballot.State() {
		case sebakcommon.BallotStateINIT:
			{
				// INITBallotValidateTransactions(checker, args)
				if !isNew || checker.VotingFinished {
					break
				}

				if checker.RoundVote.IsVotedByNode(ballot.State(), is.Node.Address()) {
					err = sebakerror.ErrorBallotAlreadyVoted
					break
				}

				if votingResult != sebakcommon.VotingNOTYET {
					break
				}

				if ballot.TransactionsLength() < 1 {
					votingResult = sebakcommon.VotingYES
					break
				}

				transactionsChecker := &BallotTransactionChecker{
					DefaultChecker: sebakcommon.DefaultChecker{handleBallotTransactionCheckerFuncs},
					Consensus:      is,
					LocalNode:      is.Node,
					NetworkID:      is.NetworkID,
					Transactions:   ballot.Transactions(),
					VotingHole:     sebakcommon.VotingNOTYET,
				}

				err = sebakcommon.RunChecker(transactionsChecker, sebakcommon.DefaultDeferFunc)
				if err != nil {
					if _, ok := err.(sebakcommon.CheckerErrorStop); !ok {
						votingResult = sebakcommon.VotingNO
						is.log.Debug("failed to handle transactions of ballot", "error", err)
						err = nil
						break
					}
					err = nil
				}

				if transactionsChecker.VotingHole == sebakcommon.VotingNO {
					votingResult = sebakcommon.VotingNO
				} else {
					votingResult = sebakcommon.VotingYES
				}
			}

			// SIGNBallotBroadcast(checker, args)
			{
				if !isNew {
					return
				}

				newBallot := ballot
				newBallot.SetSource(is.Node.Address())
				newBallot.SetVote(sebakcommon.BallotStateSIGN, votingResult)
				newBallot.Sign(is.Node.Keypair(), is.NetworkID)

				if found := is.VoteIfRunningRoundExists(newBallot); !found {
					err = errors.New("`RunningRound` not found")
				}

				is.connectionManager.Broadcast(newBallot)
				is.log.Debug("ballot will be broadcasted", "newBallot", newBallot)
			}
			// TransitStateToSIGN(checker, args)
		case sebakcommon.BallotStateSIGN:
			// BallotCheckResult(checker, args)
			{
				if !ballot.State().IsValidForVote() {
					return
				}

				result, votingHole, finished := checker.RoundVote.CanGetVotingResult(
					is.VotingThresholdPolicy,
					ballot.State(),
				)

				checker.Result = result
				checker.VotingFinished = finished
				checker.FinishedVotingHole = votingHole

				if checker.VotingFinished {
					is.log.Debug("get result", "finished VotingHole", checker.FinishedVotingHole, "result", checker.Result)
				}

			}
			// ACCEPTBallotBroadcast(checker, args)
			{
				if !checker.VotingFinished {
					return
				}

				newBallot := ballot
				newBallot.SetSource(is.Node.Address())
				newBallot.SetVote(sebakcommon.BallotStateACCEPT, checker.FinishedVotingHole)
				newBallot.Sign(is.Node.Keypair(), checker.NetworkID)

				if found := is.VoteIfRunningRoundExists(newBallot); !found {
					err = errors.New("RunningRound not found")
					return
				}

				is.connectionManager.Broadcast(newBallot)
				is.log.Debug("ballot will be broadcasted", "newBallot", newBallot)
			}
			// TransitStateToACCEPT(checker, args)
		case sebakcommon.BallotStateACCEPT:
			{
				// BallotCheckResult(checker, args)
				if !ballot.State().IsValidForVote() {
					return
				}

				result, votingHole, finished := checker.RoundVote.CanGetVotingResult(
					is.VotingThresholdPolicy,
					ballot.State(),
				)

				checker.Result = result
				checker.VotingFinished = finished
				checker.FinishedVotingHole = votingHole

				if checker.VotingFinished {
					is.log.Debug("get result", "finished VotingHole", checker.FinishedVotingHole, "result", checker.Result)
				}

			}
			{
				// FinishedBallotStore(checker, args)
				if !checker.VotingFinished {
					return
				}
				if checker.FinishedVotingHole == sebakcommon.VotingYES {
					var block Block
					block, err = FinishBallot(
						is.storage,
						ballot,
						is.TransactionPool,
					)
					if err != nil {
						return
					}

					is.SetLatestConsensusedBlock(block)
					is.log.Debug("ballot was stored", "block", block)

					err = NewCheckerStopCloseConsensus(checker, "ballot got consensus and will be stored")
				} else {
					err = NewCheckerStopCloseConsensus(checker, "ballot got consensus")
				}

				is.CloseConsensus(
					ballot.Proposer(),
					ballot.Round(),
					checker.FinishedVotingHole,
				)
				is.SetLatestRound(ballot.Round())
				// [TODO]nr.isaacStateManager.ResetRound()
			}
		}
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
		// }
	}
	return
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
