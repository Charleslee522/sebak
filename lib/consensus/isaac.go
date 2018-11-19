package consensus

import (
	"context"
	"sync"

	logging "github.com/inconshreveable/log15"

	"boscoin.io/sebak/lib/ballot"
	"boscoin.io/sebak/lib/block"
	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/network"
	"boscoin.io/sebak/lib/node"
	"boscoin.io/sebak/lib/storage"
	"boscoin.io/sebak/lib/voting"
)

type SyncController interface {
	SetSyncTargetBlock(ctx context.Context, height uint64, nodeAddrs []string) error
}

type ISAAC struct {
	sync.RWMutex

	latestBlock         block.Block
	connectionManager   network.ConnectionManager
	storage             *storage.LevelDBBackend
	proposerSelector    ProposerSelector
	log                 logging.Logger
	policy              voting.ThresholdPolicy
	votingResultManager *VotingResultManager
	syncer              SyncController
	latestReqSyncHeight uint64

	LatestBallot      ballot.Ballot
	NetworkID         []byte
	Node              *node.LocalNode
	LatestVotingBasis voting.Basis
	Conf              common.Config
}

// ISAAC should know network.ConnectionManager
// because the ISAAC uses connected validators when calculating proposer
func NewISAAC(node *node.LocalNode, p voting.ThresholdPolicy,
	cm network.ConnectionManager, st *storage.LevelDBBackend, conf common.Config, syncer SyncController) (is *ISAAC, err error) {

	is = &ISAAC{
		NetworkID:           conf.NetworkID,
		Node:                node,
		connectionManager:   cm,
		storage:             st,
		proposerSelector:    SequentialSelector{cm},
		Conf:                conf,
		log:                 log.New(logging.Ctx{"node": node.Alias()}),
		votingResultManager: NewVotingResultManager(p),
		syncer:              syncer,
		LatestBallot:        ballot.Ballot{},
	}

	return
}

func (is *ISAAC) SetLatestVotingBasis(basis voting.Basis) {
	is.LatestVotingBasis = basis
}

func (is *ISAAC) SetProposerSelector(p ProposerSelector) {
	is.proposerSelector = p
}

func (is *ISAAC) ConnectionManager() network.ConnectionManager {
	return is.connectionManager
}

func (is *ISAAC) SelectProposer(blockHeight uint64, round uint64) string {
	return is.proposerSelector.Select(blockHeight, round)
}

func (is *ISAAC) IsValidVotingBasis(basis voting.Basis, latestBlock block.Block) bool {
	if basis.Height == latestBlock.Height {
		if is.isInitRound(basis) {
			return true
		}

		if basis.BlockHash != latestBlock.Hash {
			return false
		}

		if basis.Height == is.LatestVotingBasis.Height {
			if basis.Round <= is.LatestVotingBasis.Round {
				return false
			}
		}
		return true
	}

	return false
}

func (is *ISAAC) isInitRound(basis voting.Basis) bool {
	return is.LatestVotingBasis.BlockHash == "" && basis.Height == common.GenesisBlockHeight
}

func (is *ISAAC) StartSync(height uint64, nodeAddrs []string) {
	is.log.Debug("begin ISAAC.StartSync")
	if is.syncer == nil || len(nodeAddrs) < 1 || is.latestReqSyncHeight >= height {
		return
	}
	if is.Node.State() != node.StateSYNC {
		is.log.Info("node state transits to sync", "height", height)
		is.Node.SetSync()
	}
	is.latestReqSyncHeight = height
	if err := is.syncer.SetSyncTargetBlock(context.Background(), height, nodeAddrs); err != nil {
		is.log.Error("syncer.SetSyncTargetBlock", "err", err, "height", height)
	}

	return
}

func (is *ISAAC) GetSyncInfo(b ballot.Ballot) (uint64, []string, error) {
	is.log.Debug("begin ISAAC.GetSyncInfo")
	is.RLock()
	defer is.RUnlock()

	return is.votingResultManager.getSyncInfo(b)
}

func (is *ISAAC) IsVoted(b ballot.Ballot) bool {
	is.RLock()
	defer is.RUnlock()
	return is.votingResultManager.IsVoted(b)
}

func (is *ISAAC) Vote(b ballot.Ballot) (isNew bool, err error) {
	is.RLock()
	defer is.RUnlock()

	isNew = is.votingResultManager.IsNew(b)
	is.votingResultManager.Vote(b)

	return
}

func (is *ISAAC) CanGetVotingResult(b ballot.Ballot) (RoundVoteResult, voting.Hole, bool) {
	is.RLock()
	defer is.RUnlock()
	return is.votingResultManager.CanGetVotingResult(b)
}

func (is *ISAAC) IsVotedByNode(b ballot.Ballot, node string) (bool, error) {
	// is.RLock()
	// defer is.RUnlock()
	// runningRound, _ := is.RunningRounds[b.VotingBasis().Index()]
	// if roundVote, err := runningRound.RoundVote(b.Proposer()); err == nil {
	// 	return roundVote.IsVotedByNode(b.State(), node), nil
	// } else {
	// 	return false, err
	// }
	return true, nil
}

func (is *ISAAC) HasRunningRound(basisIndex string) bool {
	// is.RLock()
	// defer is.RUnlock()
	// _, found := is.RunningRounds[basisIndex]
	// return found
	return true
}

func (is *ISAAC) HasSameProposer(b ballot.Ballot) bool {
	// is.RLock()
	// defer is.RUnlock()
	// if runningRound, found := is.RunningRounds[b.VotingBasis().Index()]; found {
	// 	return runningRound.Proposer == b.Proposer()
	// }

	return false
}

func (is *ISAAC) LatestBlock() block.Block {
	return block.GetLatestBlock(is.storage)
}

func (is *ISAAC) RemoveRunningRoundsWithSameHeight(height uint64) {
	// for hash, runningRound := range is.RunningRounds {
	// 	if runningRound.VotingBasis.Height > height {
	// 		continue
	// 	}

	// 	delete(runningRound.Voted, runningRound.Proposer)
	// 	delete(is.RunningRounds, hash)
	// }
}
