// We can see the ISAAC state change during multiple heights.

package sebak

import (
	"testing"

	common "boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/round"
	"github.com/stretchr/testify/require"
)

//	TestISAACStateNewHeight indicates the following:
//		1. The node is the proposer in every round.
//		2. There are 5 nodes and threshold is 4.
//		3. The node receives the SIGN, ACCEPT messages in order from the other four validator nodes.
//		4. The node receives a ballot that exceeds the threshold, and the block is confirmed.
//		5. Proceed '1.' ~ '4.' for 5 heights.

func TestISAACStateNewHeight(t *testing.T) {
	nodeRunners := createTestNodeRunner(5)

	messages, txs := GetNMessages(t, 5)
	nodeRunner := nodeRunners[0]

	// `nodeRunner` is proposer's runner
	nodeRunner.SetProposerCalculator(SelfProposerCalculator{})

	nodeRunner.Consensus().SetLatestConsensusedBlock(genesisBlock)
	require.Equal(t, uint64(1), nodeRunner.Consensus().LatestConfirmedBlock.Height)
	require.Equal(t, uint64(0), nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs)

	runHeight(t, nodeRunners, messages[0], txs[0], 2, 1)
	runHeight(t, nodeRunners, messages[1], txs[1], 3, 2)
	runHeight(t, nodeRunners, messages[2], txs[2], 4, 3)
	runHeight(t, nodeRunners, messages[3], txs[3], 5, 4)
	runHeight(t, nodeRunners, messages[4], txs[4], 6, 5)
}

func runHeight(t *testing.T, nodeRunners []*NodeRunner, message common.NetworkMessage, tx Transaction, height uint64, totalTxs uint64) {
	var err error
	nodeRunner := nodeRunners[0]
	err = nodeRunner.handleTransaction(message)
	require.Nil(t, err)

	require.Equal(t, 1, nodeRunner.Consensus().TransactionPool.Len())
	proposer := nodeRunner.localNode

	err = nodeRunner.proposeNewBallot(0)
	require.Nil(t, err)

	round := round.Round{
		Number:      0,
		BlockHeight: nodeRunner.Consensus().LatestConfirmedBlock.Height,
		BlockHash:   nodeRunner.Consensus().LatestConfirmedBlock.Hash,
		TotalTxs:    nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs,
	}
	runningRounds := nodeRunner.Consensus().RunningRounds

	// Check that the transaction is in RunningRounds
	ballotSIGN1 := GenerateBallot(t, proposer, round, tx, common.BallotStateSIGN, nodeRunners[1].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN1)
	require.Nil(t, err)

	ballotSIGN2 := GenerateBallot(t, proposer, round, tx, common.BallotStateSIGN, nodeRunners[2].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN2)
	require.Nil(t, err)

	ballotSIGN3 := GenerateBallot(t, proposer, round, tx, common.BallotStateSIGN, nodeRunners[3].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN3)
	require.Nil(t, err)

	ballotSIGN4 := GenerateBallot(t, proposer, round, tx, common.BallotStateSIGN, nodeRunners[4].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN4)
	require.Nil(t, err)

	rr := runningRounds[round.Hash()]
	require.NotNil(t, rr)
	require.Equal(t, 4, len(rr.Voted[proposer.Address()].GetResult(common.BallotStateSIGN)))

	ballotACCEPT1 := GenerateBallot(t, proposer, round, tx, common.BallotStateACCEPT, nodeRunners[1].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotACCEPT1)
	require.Nil(t, err)

	ballotACCEPT2 := GenerateBallot(t, proposer, round, tx, common.BallotStateACCEPT, nodeRunners[2].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotACCEPT2)
	require.Nil(t, err)

	ballotACCEPT3 := GenerateBallot(t, proposer, round, tx, common.BallotStateACCEPT, nodeRunners[3].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotACCEPT3)
	require.Nil(t, err)

	ballotACCEPT4 := GenerateBallot(t, proposer, round, tx, common.BallotStateACCEPT, nodeRunners[4].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotACCEPT4)
	require.EqualError(t, err, "ballot got consensus and will be stored")

	rr = runningRounds[round.Hash()]
	require.Nil(t, rr)

	lastConfirmedBlock := nodeRunner.Consensus().LatestConfirmedBlock
	require.Equal(t, proposer.Address(), lastConfirmedBlock.Proposer)
	require.Equal(t, height, lastConfirmedBlock.Height)
	require.Equal(t, totalTxs, lastConfirmedBlock.TotalTxs)
	require.Equal(t, tx.GetHash(), lastConfirmedBlock.Transactions[0])
	require.Equal(t, 0, len(nodeRunner.Consensus().TransactionPool.AvailableTransactions(NewISAACConfiguration())))
}
