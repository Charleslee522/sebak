package sebak

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/node"
)

func TestIsaacNilBallotFromEmptyTransactions(t *testing.T) {
	nodeRunners := createTestNodeRunner(1)

	nodeRunner := nodeRunners[0]

	// `nodeRunner` is proposer's runner
	nodeRunner.SetProposerCalculator(SelfProposerCalculator{})

	nodeRunner.Consensus().SetLatestConsensusedBlock(genesisBlock)

	roundNumber := uint64(0)

	round := Round{
		Number:      roundNumber,
		BlockHeight: nodeRunner.Consensus().LatestConfirmedBlock.Height,
		BlockHash:   nodeRunner.Consensus().LatestConfirmedBlock.Hash,
		TotalTxs:    nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs,
	}

	ballotNOTYET := NewBallot(nodeRunner.localNode, round, []string{"HASH"})
	assert.Equal(t, ballotNOTYET.B.Vote, sebakcommon.VotingNOTYET)

	ballotNIL := NewBallot(nodeRunner.localNode, round, []string{})
	assert.Equal(t, ballotNIL.B.Vote, sebakcommon.VotingEXPIRED)

}

func TestIsaacNilVoting(t *testing.T) {
	nodeRunners := createTestNodeRunner(5)

	nodeRunner := nodeRunners[0]

	// `nodeRunner` is proposer's runner
	nodeRunner.SetProposerCalculator(SelfProposerCalculator{})
	proposer := nodeRunner.localNode

	nodeRunner.Consensus().SetLatestConsensusedBlock(genesisBlock)

	// Generate proposed ballot in nodeRunner
	roundNumber := uint64(0)
	var err error

	err = nodeRunner.ProposeNewBallot(roundNumber)

	round := Round{
		Number:      roundNumber,
		BlockHeight: nodeRunner.Consensus().LatestConfirmedBlock.Height,
		BlockHash:   nodeRunner.Consensus().LatestConfirmedBlock.Hash,
		TotalTxs:    nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs,
	}

	ballotNIL := NewBallot(nodeRunner.localNode, round, []string{})
	ballotNIL.SetVote(sebakcommon.BallotStateINIT, sebakcommon.VotingEXPIRED)

	ballotSIGN1 := generateNilBallot(t, proposer, round, sebakcommon.BallotStateSIGN, nodeRunners[1].localNode)
	err = receiveBallot(t, nodeRunner, ballotSIGN1)
	assert.Nil(t, err)

	runningRounds := nodeRunner.Consensus().RunningRounds

	// Check that the transaction is in RunningRounds
	rr := runningRounds[round.Hash()]
	assert.NotNil(t, rr)

	result := rr.Voted[proposer.Address()].GetResult(sebakcommon.BallotStateSIGN)
	assert.Equal(t, 2, len(result))

	ballotACCEPT1 := generateNilBallot(t, proposer, round, sebakcommon.BallotStateACCEPT, nodeRunners[1].localNode)
	err = receiveBallot(t, nodeRunner, ballotACCEPT1)
	assert.Nil(t, err)

	ballotACCEPT2 := generateNilBallot(t, proposer, round, sebakcommon.BallotStateACCEPT, nodeRunners[2].localNode)
	err = receiveBallot(t, nodeRunner, ballotACCEPT2)
	assert.Nil(t, err)

	ballotACCEPT3 := generateNilBallot(t, proposer, round, sebakcommon.BallotStateACCEPT, nodeRunners[3].localNode)
	err = receiveBallot(t, nodeRunner, ballotACCEPT3)
	assert.EqualError(t, err, "stop checker and return: ballot got consensus and will be stored")

	assert.Equal(t, 3, len(rr.Voted[proposer.Address()].GetResult(sebakcommon.BallotStateACCEPT)))

	block := nodeRunner.Consensus().LatestConfirmedBlock
	assert.Equal(t, proposer.Address(), block.Proposer)
	assert.Equal(t, 0, len(block.Transactions))
}

func generateNilBallot(t *testing.T, proposer *sebaknode.LocalNode, round Round, ballotState sebakcommon.BallotState, sender *sebaknode.LocalNode) *Ballot {
	ballot := NewBallot(proposer, round, []string{})
	ballot.SetVote(sebakcommon.BallotStateINIT, sebakcommon.VotingEXPIRED)
	ballot.Sign(proposer.Keypair(), networkID)

	ballot.SetSource(sender.Address())
	ballot.SetVote(ballotState, sebakcommon.VotingEXPIRED)
	ballot.Sign(sender.Keypair(), networkID)

	err := ballot.IsWellFormed(networkID)
	assert.Nil(t, err)

	return ballot
}
