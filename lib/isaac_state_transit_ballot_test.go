//	In this file, there are unittests assume that one node receive a message from validators,
//	and how the state of the node changes.

package sebak

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/node"
	"boscoin.io/sebak/lib/round"
)

//	TestStateTransitFromBallot indicates the following:
//		1. Proceed for one round.
//		2. The node is the proposer of this round.
//		3. There are 5 nodes and threshold is 4.
//		4. The node receives the SIGN, ACCEPT messages in order from the other four validator nodes.
//		5. The node receives a ballot that exceeds the threshold, and the block is confirmed.

func TestStateTransitFromBallot(t *testing.T) {
	nodeRunners := createTestNodeRunner(5)

	tx, txByte := GetTransaction(t)

	message := common.NetworkMessage{Type: common.TransactionMessage, Data: txByte}

	nodeRunner := nodeRunners[0]

	// `nodeRunner` is proposer's runner
	nodeRunner.SetProposerCalculator(SelfProposerCalculator{})
	proposer := nodeRunner.localNode

	nodeRunner.Consensus().SetLatestConsensusedBlock(genesisBlock)
	require.Equal(t, uint64(1), nodeRunner.Consensus().LatestConfirmedBlock.Height)
	require.Equal(t, uint64(0), nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs)

	var err error
	err = nodeRunner.handleTransaction(message)

	require.Nil(t, err)
	require.True(t, nodeRunner.Consensus().TransactionPool.Has(tx.GetHash()))

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
	rr := runningRounds[round.Hash()]
	require.NotNil(t, rr)
	txHashs := rr.Transactions[proposer.Address()]
	require.Equal(t, tx.GetHash(), txHashs[0])

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

	rr = runningRounds[round.Hash()]
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
	require.Equal(t, uint64(2), lastConfirmedBlock.Height)
	require.Equal(t, uint64(1), lastConfirmedBlock.TotalTxs)
	require.Equal(t, 1, len(lastConfirmedBlock.Transactions))
	require.Equal(t, tx.GetHash(), lastConfirmedBlock.Transactions[0])
	require.Equal(t, 0, len(nodeRunner.Consensus().TransactionPool.AvailableTransactions(NewISAACConfiguration())))
}

//	TestStateTransitWithNoVoting indicates the following:
//		1. Proceed for one round.
//		2. The node is the proposer of this round.
//		3. There are 5 nodes and threshold is 4.
//		4. The node receives the SIGN messages with NO in order from the other four validator nodes.
//		5. The node receives a SIGN ballot that exceeds the threshold, the node broadcast B(`ACCEPT`, `NO`).

func TestStateTransitCauseNoBallot(t *testing.T) {
	nodeRunners := createTestNodeRunner(5)

	tx, txByte := GetTransaction(t)

	message := common.NetworkMessage{Type: common.TransactionMessage, Data: txByte}

	nodeRunner := nodeRunners[0]

	conf := NewISAACConfiguration()
	conf.TimeoutINIT = time.Hour
	conf.TimeoutSIGN = time.Hour
	conf.TimeoutACCEPT = time.Hour
	conf.BlockTime = time.Hour

	nodeRunner.SetConf(conf)
	nodeRunner.SetProposerCalculator(SelfProposerCalculator{})

	nodeRunner.Consensus().SetLatestConsensusedBlock(genesisBlock)
	require.Equal(t, uint64(1), nodeRunner.Consensus().LatestConfirmedBlock.Height)
	require.Equal(t, uint64(0), nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs)

	b := NewTestBroadcaster(nil)
	nodeRunner.SetBroadcaster(b)

	var err error
	err = nodeRunner.handleTransaction(message)

	require.Nil(t, err)
	require.True(t, nodeRunner.Consensus().TransactionPool.Has(tx.GetHash()))

	nodeRunner.StartStateManager()
	time.Sleep(200 * time.Millisecond)

	require.Nil(t, err)
	round := round.Round{
		Number:      0,
		BlockHeight: nodeRunner.Consensus().LatestConfirmedBlock.Height,
		BlockHash:   nodeRunner.Consensus().LatestConfirmedBlock.Hash,
		TotalTxs:    nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs,
	}
	proposer := nodeRunner.localNode
	runningRounds := nodeRunner.Consensus().RunningRounds

	// Check that the transaction is in RunningRounds
	rr := runningRounds[round.Hash()]
	require.NotNil(t, rr)
	txHashs := rr.Transactions[proposer.Address()]
	require.Equal(t, tx.GetHash(), txHashs[0])

	ballotSIGN1 := GenerateBallot(t, proposer, round, tx, common.BallotStateSIGN, nodeRunners[1].localNode)
	ballotSIGN1.SetVote(common.BallotStateSIGN, common.VotingNO)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN1)
	require.Nil(t, err)

	ballotSIGN2 := GenerateBallot(t, proposer, round, tx, common.BallotStateSIGN, nodeRunners[2].localNode)
	ballotSIGN2.SetVote(common.BallotStateSIGN, common.VotingNO)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN2)
	require.Nil(t, err)

	ballotSIGN3 := GenerateBallot(t, proposer, round, tx, common.BallotStateSIGN, nodeRunners[3].localNode)
	ballotSIGN3.SetVote(common.BallotStateSIGN, common.VotingNO)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN3)
	require.Nil(t, err)

	ballotSIGN4 := GenerateBallot(t, proposer, round, tx, common.BallotStateSIGN, nodeRunners[4].localNode)
	ballotSIGN4.SetVote(common.BallotStateSIGN, common.VotingNO)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN4)
	require.Nil(t, err)

	time.Sleep(100 * time.Millisecond)

	require.Equal(t, 3, len(b.Messages))

	init, sign, accept := 0, 0, 0

	for _, message := range b.Messages {
		ballot, ok := message.(Ballot)
		if !ok {
			continue
		}
		switch ballot.State() {
		case common.BallotStateINIT:
			init++
			require.Equal(t, common.VotingYES, ballot.Vote())
		case common.BallotStateSIGN:
			sign++
		case common.BallotStateACCEPT:
			require.Equal(t, common.VotingNO, ballot.Vote())
			accept++
		}
	}

	require.Equal(t, 1, init)
	require.Equal(t, 0, sign)
	require.Equal(t, 1, accept)
}

//	TestStateTransitCauseEXPBallot indicates the following:
//		1. Proceed for one round.
//		2. The node is the proposer of this round.
//		3. There are 5 nodes and threshold is 4.
//		4. The node receives the SIGN messages with EXPIRED in order from the other four validator nodes.
//		5. The ballots which is included in the node cannot exceeds the threshold as `YES`, then the node broadcast ballot with `EXPIRED`

func TestStateTransitCauseEXPBallot(t *testing.T) {
	nodeRunners := createTestNodeRunner(5)

	tx, txByte := GetTransaction(t)

	message := common.NetworkMessage{Type: common.TransactionMessage, Data: txByte}

	nodeRunner := nodeRunners[0]

	conf := NewISAACConfiguration()
	conf.TimeoutINIT = time.Hour
	conf.TimeoutSIGN = time.Hour
	conf.TimeoutACCEPT = time.Hour
	conf.BlockTime = time.Hour

	nodeRunner.SetConf(conf)
	nodeRunner.SetProposerCalculator(SelfProposerCalculator{})

	nodeRunner.Consensus().SetLatestConsensusedBlock(genesisBlock)
	require.Equal(t, uint64(1), nodeRunner.Consensus().LatestConfirmedBlock.Height)
	require.Equal(t, uint64(0), nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs)

	b := NewTestBroadcaster(nil)
	nodeRunner.SetBroadcaster(b)

	var err error
	err = nodeRunner.handleTransaction(message)

	require.Nil(t, err)
	require.True(t, nodeRunner.Consensus().TransactionPool.Has(tx.GetHash()))

	nodeRunner.StartStateManager()
	time.Sleep(200 * time.Millisecond)

	require.Nil(t, err)
	round := round.Round{
		Number:      0,
		BlockHeight: nodeRunner.Consensus().LatestConfirmedBlock.Height,
		BlockHash:   nodeRunner.Consensus().LatestConfirmedBlock.Hash,
		TotalTxs:    nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs,
	}
	proposer := nodeRunner.localNode
	runningRounds := nodeRunner.Consensus().RunningRounds

	// Check that the transaction is in RunningRounds
	rr := runningRounds[round.Hash()]
	require.NotNil(t, rr)
	txHashs := rr.Transactions[proposer.Address()]
	require.Equal(t, tx.GetHash(), txHashs[0])

	ballotSIGN1 := GenerateBallot(t, proposer, round, tx, common.BallotStateSIGN, nodeRunners[1].localNode)
	ballotSIGN1.SetVote(common.BallotStateSIGN, common.VotingEXP)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN1)
	require.Nil(t, err)

	ballotSIGN2 := GenerateBallot(t, proposer, round, tx, common.BallotStateSIGN, nodeRunners[2].localNode)
	ballotSIGN2.SetVote(common.BallotStateSIGN, common.VotingEXP)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN2)
	require.Nil(t, err)

	ballotSIGN3 := GenerateBallot(t, proposer, round, tx, common.BallotStateSIGN, nodeRunners[3].localNode)
	ballotSIGN3.SetVote(common.BallotStateSIGN, common.VotingEXP)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN3)
	require.Nil(t, err)

	time.Sleep(100 * time.Millisecond)

	require.Equal(t, 3, len(b.Messages))

	init, sign, accept := 0, 0, 0

	for _, message := range b.Messages {
		ballot, ok := message.(Ballot)
		if !ok {
			continue
		}
		switch ballot.State() {
		case common.BallotStateINIT:
			init++
			require.Equal(t, common.VotingYES, ballot.Vote())
		case common.BallotStateSIGN:
			sign++
		case common.BallotStateACCEPT:
			require.Equal(t, common.VotingEXP, ballot.Vote())
			accept++
		}
	}

	require.Equal(t, 1, init)
	require.Equal(t, 0, sign)
	require.Equal(t, 1, accept)
}

func TestStateTransitSIGNTimeoutACCEPTBallotProposer(t *testing.T) {
	nodeRunners := createTestNodeRunner(5)

	tx, txByte := GetTransaction(t)

	message := common.NetworkMessage{Type: common.TransactionMessage, Data: txByte}

	nodeRunner := nodeRunners[0]

	b := NewTestBroadcaster(nil)
	nodeRunner.SetBroadcaster(b)

	conf := NewISAACConfiguration()
	conf.TimeoutINIT = time.Hour
	conf.TimeoutSIGN = time.Millisecond
	conf.TimeoutACCEPT = time.Hour
	conf.BlockTime = 200 * time.Millisecond

	nodeRunner.SetConf(conf)

	// `nodeRunner` is proposer's runner
	nodeRunner.SetProposerCalculator(SelfProposerCalculator{})
	proposer := nodeRunner.localNode

	nodeRunner.Consensus().SetLatestConsensusedBlock(genesisBlock)
	require.Equal(t, uint64(1), nodeRunner.Consensus().LatestConfirmedBlock.Height)
	require.Equal(t, uint64(0), nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs)

	var err error
	err = nodeRunner.handleTransaction(message)

	require.Nil(t, err)
	require.True(t, nodeRunner.Consensus().TransactionPool.Has(tx.GetHash()))

	nodeRunner.StartStateManager()

	time.Sleep(200 * time.Millisecond)

	init, sign, accept, ballots := 0, 0, 0, 0
	for _, message := range b.Messages {
		ballot, ok := message.(Ballot)
		if !ok {
			continue
		}
		ballots++
		switch ballot.State() {
		case common.BallotStateINIT:
			init++
			require.Equal(t, common.VotingYES, ballot.Vote())
		case common.BallotStateSIGN:
			sign++
		case common.BallotStateACCEPT:
			accept++
			require.Equal(t, common.VotingEXP, ballot.Vote())
		}
	}
	require.Equal(t, 2, ballots)
	require.Equal(t, 1, init)
	require.Equal(t, 0, sign)
	require.Equal(t, 1, accept)

	round := round.Round{
		Number:      0,
		BlockHeight: nodeRunner.Consensus().LatestConfirmedBlock.Height,
		BlockHash:   nodeRunner.Consensus().LatestConfirmedBlock.Hash,
		TotalTxs:    nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs,
	}
	runningRounds := nodeRunner.Consensus().RunningRounds

	// Check that the transaction is in RunningRounds
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

	rr := runningRounds[round.Hash()]
	require.Nil(t, rr)

	lastConfirmedBlock := nodeRunner.Consensus().LatestConfirmedBlock
	require.Equal(t, proposer.Address(), lastConfirmedBlock.Proposer)
	require.Equal(t, uint64(2), lastConfirmedBlock.Height)
	require.Equal(t, uint64(1), lastConfirmedBlock.TotalTxs)
	require.Equal(t, 1, len(lastConfirmedBlock.Transactions))
	require.Equal(t, tx.GetHash(), lastConfirmedBlock.Transactions[0])
}

func TestStateTransitSIGNBallotACCEPTTimeoutProposer(t *testing.T) {
	nodeRunners := createTestNodeRunner(5)

	tx, txByte := GetTransaction(t)

	message := common.NetworkMessage{Type: common.TransactionMessage, Data: txByte}

	nodeRunner := nodeRunners[0]

	b := NewTestBroadcaster(nil)
	nodeRunner.SetBroadcaster(b)

	conf := NewISAACConfiguration()
	conf.TimeoutINIT = time.Hour
	conf.TimeoutSIGN = time.Hour
	conf.TimeoutACCEPT = time.Millisecond
	conf.BlockTime = 200 * time.Millisecond

	nodeRunner.SetConf(conf)

	// `nodeRunner` is proposer's runner
	nodeRunner.SetProposerCalculator(SelfProposerCalculator{})
	proposer := nodeRunner.localNode

	nodeRunner.Consensus().SetLatestConsensusedBlock(genesisBlock)
	require.Equal(t, uint64(1), nodeRunner.Consensus().LatestConfirmedBlock.Height)
	require.Equal(t, uint64(0), nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs)

	var err error
	err = nodeRunner.handleTransaction(message)

	require.Nil(t, err)
	require.True(t, nodeRunner.Consensus().TransactionPool.Has(tx.GetHash()))

	nodeRunner.StartStateManager()

	time.Sleep(200 * time.Millisecond)
	// Generate proposed ballot in nodeRunner
	r := round.Round{
		Number:      0,
		BlockHeight: nodeRunner.Consensus().LatestConfirmedBlock.Height,
		BlockHash:   nodeRunner.Consensus().LatestConfirmedBlock.Hash,
		TotalTxs:    nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs,
	}

	runningRounds := nodeRunner.Consensus().RunningRounds
	require.Equal(t, 1, len(runningRounds))

	ballotSIGN1 := GenerateBallot(t, proposer, r, tx, common.BallotStateSIGN, nodeRunners[1].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN1)
	require.Nil(t, err)

	ballotSIGN2 := GenerateBallot(t, proposer, r, tx, common.BallotStateSIGN, nodeRunners[2].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN2)
	require.Nil(t, err)

	ballotSIGN3 := GenerateBallot(t, proposer, r, tx, common.BallotStateSIGN, nodeRunners[3].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN3)
	require.Nil(t, err)

	runningRounds = nodeRunner.Consensus().RunningRounds
	require.Equal(t, 1, len(runningRounds))

	time.Sleep(200 * time.Millisecond)
	require.Equal(t, 2, len(runningRounds))

	rr, ok := runningRounds[r.Hash()]
	require.True(t, ok)
	require.NotNil(t, rr)
	require.Equal(t, 3, len(rr.Voted[proposer.Address()].GetResult(common.BallotStateSIGN)))

	ballotSIGN4 := GenerateBallot(t, proposer, r, tx, common.BallotStateSIGN, nodeRunners[4].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN4)
	require.Nil(t, err)

	rr, ok = runningRounds[r.Hash()]
	require.True(t, ok)
	require.NotNil(t, rr)
	require.Equal(t, 4, len(rr.Voted[proposer.Address()].GetResult(common.BallotStateSIGN)))

	time.Sleep(200 * time.Millisecond)
	require.Equal(t, 2, len(runningRounds))

	nextRound := round.Round{
		Number:      1,
		BlockHeight: nodeRunner.Consensus().LatestConfirmedBlock.Height,
		BlockHash:   nodeRunner.Consensus().LatestConfirmedBlock.Hash,
		TotalTxs:    nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs,
	}

	nextRr, ok := runningRounds[nextRound.Hash()]
	require.True(t, ok)
	require.NotNil(t, nextRr)
	require.Equal(t, 0, len(nextRr.Voted[proposer.Address()].GetResult(common.BallotStateSIGN)))

	time.Sleep(200 * time.Millisecond)

	init, sign, accept, ballots := 0, 0, 0, 0
	for _, message := range b.Messages {
		ballot, ok := message.(Ballot)
		if !ok {
			continue
		}
		ballots++
		switch ballot.State() {
		case common.BallotStateINIT:
			init++
			require.Equal(t, common.VotingYES, ballot.Vote())
		case common.BallotStateSIGN:
			sign++
		case common.BallotStateACCEPT:
			accept++
		}
	}
	require.Equal(t, 4, ballots)
	require.Equal(t, 2, init) // INIT(Send B(`INIT`, `YES`) cause Proposer itself) -> SIGN(Receive 4 Ballots) -> ACCEPT(Timeout) -> INIT(Send B(`INIT`, `YES`)
	require.Equal(t, 0, sign)
	require.Equal(t, 2, accept)
}

func TestStateTransitTwoBlocks(t *testing.T) {
	nodeRunners := createTestNodeRunner(5)

	tx, txByte := GetTransaction(t)

	message := common.NetworkMessage{Type: common.TransactionMessage, Data: txByte}

	nodeRunner := nodeRunners[0]

	b := NewTestBroadcaster(nil)
	nodeRunner.SetBroadcaster(b)

	conf := NewISAACConfiguration()
	conf.TimeoutINIT = time.Hour
	conf.TimeoutSIGN = time.Hour
	conf.TimeoutACCEPT = time.Hour
	conf.BlockTime = 1 * time.Millisecond

	nodeRunner.SetConf(conf)

	// `nodeRunner` is proposer's runner
	nodeRunner.SetProposerCalculator(&SelfProposerThenNotProposer{})
	proposer := nodeRunner.localNode

	nodeRunner.Consensus().SetLatestConsensusedBlock(genesisBlock)
	require.Equal(t, uint64(1), nodeRunner.Consensus().LatestConfirmedBlock.Height)
	require.Equal(t, uint64(0), nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs)

	var err error
	err = nodeRunner.handleTransaction(message)

	require.Nil(t, err)
	require.True(t, nodeRunner.Consensus().TransactionPool.Has(tx.GetHash()))

	nodeRunner.StartStateManager()

	time.Sleep(200 * time.Millisecond)

	init, sign, accept, ballots := 0, 0, 0, 0
	for _, message := range b.Messages {
		ballot, ok := message.(Ballot)
		if !ok {
			continue
		}
		ballots++
		switch ballot.State() {
		case common.BallotStateINIT:
			init++
			require.Equal(t, common.VotingYES, ballot.Vote())
		case common.BallotStateSIGN:
			sign++
		case common.BallotStateACCEPT:
			accept++
		}
	}
	require.Equal(t, 1, ballots)
	require.Equal(t, 1, init)
	require.Equal(t, 0, sign)
	require.Equal(t, 0, accept)

	r := round.Round{
		Number:      0,
		BlockHeight: nodeRunner.Consensus().LatestConfirmedBlock.Height,
		BlockHash:   nodeRunner.Consensus().LatestConfirmedBlock.Hash,
		TotalTxs:    nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs,
	}
	runningRounds := nodeRunner.Consensus().RunningRounds

	// Check that the transaction is in RunningRounds
	ballotACCEPT1 := GenerateBallot(t, proposer, r, tx, common.BallotStateACCEPT, nodeRunners[1].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotACCEPT1)
	require.Nil(t, err)

	ballotACCEPT2 := GenerateBallot(t, proposer, r, tx, common.BallotStateACCEPT, nodeRunners[2].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotACCEPT2)
	require.Nil(t, err)

	ballotACCEPT3 := GenerateBallot(t, proposer, r, tx, common.BallotStateACCEPT, nodeRunners[3].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotACCEPT3)
	require.Nil(t, err)

	ballotACCEPT4 := GenerateBallot(t, proposer, r, tx, common.BallotStateACCEPT, nodeRunners[4].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotACCEPT4)
	require.EqualError(t, err, "ballot got consensus and will be stored")

	rr := runningRounds[r.Hash()]
	require.Nil(t, rr)

	lastConfirmedBlock := nodeRunner.Consensus().LatestConfirmedBlock
	require.Equal(t, proposer.Address(), lastConfirmedBlock.Proposer)
	require.Equal(t, uint64(2), lastConfirmedBlock.Height)
	require.Equal(t, uint64(1), lastConfirmedBlock.TotalTxs)
	require.Equal(t, 1, len(lastConfirmedBlock.Transactions))
	require.Equal(t, tx.GetHash(), lastConfirmedBlock.Transactions[0])
	require.Equal(t, 0, len(nodeRunner.Consensus().TransactionPool.AvailableTransactions(NewISAACConfiguration())))

	// New Height with empty transactions
	r = round.Round{
		Number:      0,
		BlockHeight: nodeRunner.Consensus().LatestConfirmedBlock.Height,
		BlockHash:   nodeRunner.Consensus().LatestConfirmedBlock.Hash,
		TotalTxs:    nodeRunner.Consensus().LatestConfirmedBlock.TotalTxs,
	}

	require.Equal(t, uint64(2), r.BlockHeight)

	calculator := SelfProposerThenNotProposer{}
	nextProposerAddress := calculator.Calculate(nodeRunner, 2, 0)

	var nextProposer *node.LocalNode = proposer
	for _, nodeRunner := range nodeRunners {
		if nodeRunner.localNode.Address() == nextProposerAddress {
			nextProposer = nodeRunner.localNode
			continue
		}
	}

	ballotProposeINIT := GenerateEmptyTxBallot(t, nextProposer, r, common.BallotStateINIT, nextProposer)
	require.Equal(t, common.BallotStateINIT, nodeRunner.isaacStateManager.State().ballotState)
	err = ReceiveBallot(t, nodeRunner, ballotProposeINIT)
	require.Nil(t, err)

	time.Sleep(100 * time.Millisecond)

	require.Equal(t, common.BallotStateSIGN, nodeRunner.isaacStateManager.State().ballotState)

	ballotSIGN1 := GenerateEmptyTxBallot(t, nextProposer, r, common.BallotStateSIGN, nodeRunners[1].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN1)
	require.Nil(t, err)

	ballotSIGN2 := GenerateEmptyTxBallot(t, nextProposer, r, common.BallotStateSIGN, nodeRunners[2].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN2)
	require.Nil(t, err)

	ballotSIGN3 := GenerateEmptyTxBallot(t, nextProposer, r, common.BallotStateSIGN, nodeRunners[3].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN3)
	require.Nil(t, err)

	ballotSIGN4 := GenerateEmptyTxBallot(t, nextProposer, r, common.BallotStateSIGN, nodeRunners[4].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotSIGN4)
	require.Nil(t, err)

	time.Sleep(200 * time.Millisecond)

	ballotACCEPT1 = GenerateEmptyTxBallot(t, nextProposer, r, common.BallotStateACCEPT, nodeRunners[1].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotACCEPT1)
	require.Nil(t, err)

	ballotACCEPT2 = GenerateEmptyTxBallot(t, nextProposer, r, common.BallotStateACCEPT, nodeRunners[2].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotACCEPT2)
	require.Nil(t, err)

	ballotACCEPT3 = GenerateEmptyTxBallot(t, nextProposer, r, common.BallotStateACCEPT, nodeRunners[3].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotACCEPT3)
	require.Nil(t, err)

	ballotACCEPT4 = GenerateEmptyTxBallot(t, nextProposer, r, common.BallotStateACCEPT, nodeRunners[4].localNode)
	err = ReceiveBallot(t, nodeRunner, ballotACCEPT4)
	require.EqualError(t, err, "ballot got consensus and will be stored")

	time.Sleep(200 * time.Millisecond)

	rr = runningRounds[r.Hash()]
	require.Nil(t, rr)

	lastConfirmedBlock = nodeRunner.Consensus().LatestConfirmedBlock
	require.Equal(t, uint64(3), lastConfirmedBlock.Height)
	require.Equal(t, uint64(1), lastConfirmedBlock.TotalTxs)
	require.Equal(t, 0, len(lastConfirmedBlock.Transactions))
	require.Equal(t, 0, len(nodeRunner.Consensus().TransactionPool.AvailableTransactions(NewISAACConfiguration())))
}
