package sebak

import (
	"testing"

	"github.com/stellar/go/keypair"
	"github.com/stretchr/testify/assert"

	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/network"
	"boscoin.io/sebak/lib/node"
)

func TestIsaacSimulationProposer(t *testing.T) {
	nodeRunners := CreateTestNodeRunner(5)

	tx, txByte := getTransaction(t)

	message := sebaknetwork.Message{Type: sebaknetwork.MessageFromClient, Data: txByte}

	nodeRunner := nodeRunners[0]

	// `nodeRunner` is proposer's runner
	nodeRunner.SetProposerCalculator(SelfProposerCalculator{})
	proposer := nodeRunner.localNode

	var err error
	err = nodeRunner.handleMessageFromClient(message)
	assert.Nil(t, err)
	assert.True(t, nodeRunner.Consensus().TransactionPool.Has(tx.GetHash()))

	// Generate proposed ballot in nodeRunner
	err = nodeRunner.ProposeNewBallot(0)
	assert.Nil(t, err)
	runningRounds := nodeRunner.Consensus().RunningRounds

	round := Round{
		Number:      0,
		BlockHeight: 0,
		BlockHash:   "",
		TotalTxs:    0,
	}

	// Check that the transaction is in RunningRounds
	rr := runningRounds[round.Hash()]
	txHashs := rr.Transactions[proposer.Address()]
	assert.Equal(t, tx.GetHash(), txHashs[0])

	ballotSIGN1 := generateBallot(t, proposer, round, tx, sebakcommon.BallotStateSIGN, nodeRunners[1].localNode)
	err = receiveBallot(t, nodeRunner, ballotSIGN1)
	assert.Nil(t, err)

	ballotSIGN2 := generateBallot(t, proposer, round, tx, sebakcommon.BallotStateSIGN, nodeRunners[2].localNode)
	err = receiveBallot(t, nodeRunner, ballotSIGN2)
	assert.Nil(t, err)

	ballotSIGN3 := generateBallot(t, proposer, round, tx, sebakcommon.BallotStateSIGN, nodeRunners[3].localNode)
	err = receiveBallot(t, nodeRunner, ballotSIGN3)
	assert.Nil(t, err)

	ballotSIGN4 := generateBallot(t, proposer, round, tx, sebakcommon.BallotStateSIGN, nodeRunners[4].localNode)
	err = receiveBallot(t, nodeRunner, ballotSIGN4)
	assert.Nil(t, err)

	assert.Equal(t, 4, len(rr.Voted[proposer.Address()].GetResult(sebakcommon.BallotStateSIGN)))

	ballotACCEPT1 := generateBallot(t, proposer, round, tx, sebakcommon.BallotStateACCEPT, nodeRunners[1].localNode)
	err = receiveBallot(t, nodeRunner, ballotACCEPT1)
	assert.Nil(t, err)

	ballotACCEPT2 := generateBallot(t, proposer, round, tx, sebakcommon.BallotStateACCEPT, nodeRunners[2].localNode)
	err = receiveBallot(t, nodeRunner, ballotACCEPT2)
	assert.Nil(t, err)

	ballotACCEPT3 := generateBallot(t, proposer, round, tx, sebakcommon.BallotStateACCEPT, nodeRunners[3].localNode)
	err = receiveBallot(t, nodeRunner, ballotACCEPT3)
	assert.Nil(t, err)

	ballotACCEPT4 := generateBallot(t, proposer, round, tx, sebakcommon.BallotStateACCEPT, nodeRunners[4].localNode)
	err = receiveBallot(t, nodeRunner, ballotACCEPT4)
	assert.EqualError(t, err, "stop checker and return: ballot got consensus and will be stored")

	assert.Equal(t, 4, len(rr.Voted[proposer.Address()].GetResult(sebakcommon.BallotStateACCEPT)))

	block := nodeRunner.Consensus().LatestConfirmedBlock
	assert.Equal(t, proposer.Address(), block.Proposer)
	assert.Equal(t, 1, len(block.Transactions))
	assert.Equal(t, tx.GetHash(), block.Transactions[0])
}

func getTransaction(t *testing.T) (tx Transaction, txByte []byte) {
	initialBalance := sebakcommon.Amount(1)
	kpNewAccount, _ := keypair.Random()

	tx = makeTransactionCreateAccount(kp, kpNewAccount.Address(), initialBalance)
	tx.B.Checkpoint = account.Checkpoint
	tx.Sign(kp, networkID)

	var err error

	txByte, err = tx.Serialize()
	assert.Nil(t, err)

	return
}

func generateBallot(t *testing.T, proposer *sebaknode.LocalNode, round Round, tx Transaction, ballotState sebakcommon.BallotState, sender *sebaknode.LocalNode) *Ballot {
	ballot := NewBallot(proposer, round, []string{tx.GetHash()})
	ballot.SetVote(sebakcommon.BallotStateINIT, VotingYES)
	ballot.Sign(proposer.Keypair(), networkID)

	ballot.SetSource(sender.Address())
	ballot.SetVote(ballotState, VotingYES)
	ballot.Sign(sender.Keypair(), networkID)

	err := ballot.IsWellFormed(networkID)
	assert.Nil(t, err)

	return ballot
}

func receiveBallot(t *testing.T, nodeRunner *NodeRunner, ballot *Ballot) error {
	data, err := ballot.Serialize()
	assert.Nil(t, err)

	ballotMessage := sebaknetwork.Message{Type: sebaknetwork.BallotMessage, Data: data}
	err = nodeRunner.handleBallotMessage(ballotMessage)
	return err
}
