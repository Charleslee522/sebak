package sebak

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/round"
)

func TestISAACNewBallotWithEmptyTransaction(t *testing.T) {
	nodeRunners := createTestNodeRunner(1)

	nodeRunner := nodeRunners[0]

	// `nodeRunner` is proposer's runner
	nodeRunner.SetProposerCalculator(SelfProposerCalculator{})

	nodeRunner.Consensus().SetLatestConsensusedBlock(genesisBlock)

	round := round.Round{
		Number:      0,
		BlockHeight: nodeRunner.Consensus().LatestConfirmedBlock.Height,
		BlockHash:   nodeRunner.Consensus().LatestConfirmedBlock.Hash,
		TotalTxs:    nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs,
	}

	ballot := NewBallot(nodeRunner.localNode, round, []string{})
	assert.Equal(t, ballot.B.Vote, common.VotingYES)

}

func TestISAACBallotWithEmptyTransactionVoting(t *testing.T) {
	nodeRunners := createTestNodeRunner(5)

	nodeRunner := nodeRunners[0]

	// `nodeRunner` is proposer's runner
	nodeRunner.SetProposerCalculator(SelfProposerCalculator{})
	proposer := nodeRunner.localNode

	nodeRunner.Consensus().SetLatestConsensusedBlock(genesisBlock)
	assert.Equal(t, uint64(1), nodeRunner.Consensus().LatestConfirmedBlock.Height)
	assert.Equal(t, uint64(0), nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs)

	// Generate proposed ballot in nodeRunner
	var err error

	err = nodeRunner.proposeNewBallot(0)

	round := round.Round{
		Number:      0,
		BlockHeight: nodeRunner.Consensus().LatestConfirmedBlock.Height,
		BlockHash:   nodeRunner.Consensus().LatestConfirmedBlock.Hash,
		TotalTxs:    nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs,
	}

	ballot := NewBallot(nodeRunner.localNode, round, []string{})
	ballot.SetVote(common.BallotStateINIT, common.VotingYES)

	ballotSIGN1 := GenerateEmptyTxBallot(t, proposer, round, common.BallotStateSIGN, nodeRunners[1].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN1)
	assert.Nil(t, err)

	ballotSIGN2 := GenerateEmptyTxBallot(t, proposer, round, common.BallotStateSIGN, nodeRunners[2].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN2)
	assert.Nil(t, err)

	ballotSIGN3 := GenerateEmptyTxBallot(t, proposer, round, common.BallotStateSIGN, nodeRunners[3].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN3)
	assert.Nil(t, err)

	runningRounds := nodeRunner.Consensus().RunningRounds

	// Check that the transaction is in RunningRounds
	rr := runningRounds[round.Hash()]
	assert.NotNil(t, rr)

	result := rr.Voted[proposer.Address()].GetResult(common.BallotStateSIGN)
	assert.Equal(t, 3, len(result))

	ballotACCEPT1 := GenerateEmptyTxBallot(t, proposer, round, common.BallotStateACCEPT, nodeRunners[1].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotACCEPT1)
	assert.Nil(t, err)

	ballotACCEPT2 := GenerateEmptyTxBallot(t, proposer, round, common.BallotStateACCEPT, nodeRunners[2].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotACCEPT2)
	assert.Nil(t, err)

	ballotACCEPT3 := GenerateEmptyTxBallot(t, proposer, round, common.BallotStateACCEPT, nodeRunners[3].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotACCEPT3)
	assert.Nil(t, err)

	ballotACCEPT4 := GenerateEmptyTxBallot(t, proposer, round, common.BallotStateACCEPT, nodeRunners[4].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotACCEPT4)
	assert.EqualError(t, err, "ballot got consensus and will be stored")

	assert.Equal(t, 4, len(rr.Voted[proposer.Address()].GetResult(common.BallotStateACCEPT)))

	lastConfirmedBlock := nodeRunner.Consensus().LatestConfirmedBlock
	assert.Equal(t, proposer.Address(), lastConfirmedBlock.Proposer)
	assert.Equal(t, uint64(2), lastConfirmedBlock.Height)
	assert.Equal(t, uint64(0), lastConfirmedBlock.TotalTxs)
	assert.Equal(t, 0, len(lastConfirmedBlock.Transactions))
}
