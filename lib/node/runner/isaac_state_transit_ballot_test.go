//	In this file, there are unittests assume that one node receive a message from validators,
//	and how the state of the node changes.

package runner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"boscoin.io/sebak/lib/ballot"
	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/consensus"
	"boscoin.io/sebak/lib/consensus/round"
)

//	TestStateTransitFromBallot indicates the following:
//		1. Proceed for one round.
//		2. The node is the proposer of this round.
//		3. There are 5 nodes and threshold is 4.
//		4. The node receives the SIGN, ACCEPT messages in order from the other four validator nodes.
//		5. The node receives a ballot that exceeds the threshold, and the block is confirmed.

func TestStateTransitFromBallot(t *testing.T) {
	conf := consensus.NewISAACConfiguration()
	nr, nodes, _ := createNodeRunnerForTesting(5, conf, nil)

	tx, txByte := GetTransaction(t)

	message := common.NetworkMessage{Type: common.TransactionMessage, Data: txByte}

	proposer := nr.localNode

	nr.Consensus().SetLatestConsensusedBlock(genesisBlock)
	latestBlock := nr.Consensus().LatestConfirmedBlock()
	require.Equal(t, uint64(1), latestBlock.Height)
	require.Equal(t, uint64(0), latestBlock.TotalTxs)

	var err error
	err = nr.handleTransaction(message)

	require.Nil(t, err)
	require.True(t, nr.Consensus().TransactionPool.Has(tx.GetHash()))

	err = nr.proposeNewBallot(0)
	require.Nil(t, err)

	round := round.Round{
		Number:      0,
		BlockHeight: latestBlock.Height,
		BlockHash:   latestBlock.Hash,
		TotalTxs:    latestBlock.TotalTxs,
	}
	runningRounds := nr.Consensus().RunningRounds

	// Check that the transaction is in RunningRounds
	rr := runningRounds[round.Hash()]
	require.NotNil(t, rr)
	txHashs := rr.Transactions[proposer.Address()]
	require.Equal(t, tx.GetHash(), txHashs[0])

	ballotSIGN1 := GenerateBallot(t, proposer, round, tx, ballot.StateSIGN, nodes[1])
	err = ReceiveBallot(t, nr, ballotSIGN1)
	require.Nil(t, err)

	ballotSIGN2 := GenerateBallot(t, proposer, round, tx, ballot.StateSIGN, nodes[2])
	err = ReceiveBallot(t, nr, ballotSIGN2)
	require.Nil(t, err)

	ballotSIGN3 := GenerateBallot(t, proposer, round, tx, ballot.StateSIGN, nodes[3])
	err = ReceiveBallot(t, nr, ballotSIGN3)
	require.Nil(t, err)

	ballotSIGN4 := GenerateBallot(t, proposer, round, tx, ballot.StateSIGN, nodes[4])
	err = ReceiveBallot(t, nr, ballotSIGN4)
	require.Nil(t, err)

	rr = runningRounds[round.Hash()]
	require.NotNil(t, rr)
	require.Equal(t, 4, len(rr.Voted[proposer.Address()].GetResult(ballot.StateSIGN)))

	ballotACCEPT1 := GenerateBallot(t, proposer, round, tx, ballot.StateACCEPT, nodes[1])
	err = ReceiveBallot(t, nr, ballotACCEPT1)
	require.Nil(t, err)

	ballotACCEPT2 := GenerateBallot(t, proposer, round, tx, ballot.StateACCEPT, nodes[2])
	err = ReceiveBallot(t, nr, ballotACCEPT2)
	require.Nil(t, err)

	ballotACCEPT3 := GenerateBallot(t, proposer, round, tx, ballot.StateACCEPT, nodes[3])
	err = ReceiveBallot(t, nr, ballotACCEPT3)
	require.Nil(t, err)

	ballotACCEPT4 := GenerateBallot(t, proposer, round, tx, ballot.StateACCEPT, nodes[4])
	err = ReceiveBallot(t, nr, ballotACCEPT4)
	require.EqualError(t, err, "ballot got consensus and will be stored")

	rr = runningRounds[round.Hash()]
	require.Nil(t, rr)

	lastConfirmedBlock := nr.Consensus().LatestConfirmedBlock()
	require.Equal(t, proposer.Address(), lastConfirmedBlock.Proposer)
	require.Equal(t, uint64(2), lastConfirmedBlock.Height)
	require.Equal(t, uint64(1), lastConfirmedBlock.TotalTxs)
	require.Equal(t, 1, len(lastConfirmedBlock.Transactions))
	require.Equal(t, tx.GetHash(), lastConfirmedBlock.Transactions[0])
	require.Equal(t, 0, len(nr.Consensus().TransactionPool.AvailableTransactions(1000)))
}

//	TestStateTransitWithNoVoting indicates the following:
//		1. Proceed for one round.
//		2. The node is the proposer of this round.
//		3. There are 5 nodes and threshold is 4.
//		4. The node receives the SIGN messages with NO in order from the other four validator nodes.
//		5. The node receives a SIGN ballot that exceeds the threshold, the node broadcast B(`ACCEPT`, `NO`).

func TestStateTransitCauseNoBallot(t *testing.T) {
	conf := consensus.NewISAACConfiguration()
	conf.TimeoutINIT = time.Hour
	conf.TimeoutSIGN = time.Hour
	conf.TimeoutACCEPT = time.Hour

	recv := make(chan struct{})
	nr, nodes, cm := createNodeRunnerForTesting(5, conf, recv)

	tx, txByte := GetTransaction(t)

	message := common.NetworkMessage{Type: common.TransactionMessage, Data: txByte}

	nr.Consensus().SetLatestConsensusedBlock(genesisBlock)
	latestBlock := nr.Consensus().LatestConfirmedBlock()
	require.Equal(t, uint64(1), latestBlock.Height)
	require.Equal(t, uint64(0), latestBlock.TotalTxs)

	var err error
	err = nr.handleTransaction(message)
	<-recv

	require.Nil(t, err)
	require.True(t, nr.Consensus().TransactionPool.Has(tx.GetHash()))

	nr.StartStateManager()

	<-recv

	round := round.Round{
		Number:      0,
		BlockHeight: latestBlock.Height,
		BlockHash:   latestBlock.Hash,
		TotalTxs:    latestBlock.TotalTxs,
	}
	proposer := nr.localNode
	runningRounds := nr.Consensus().RunningRounds

	// Check that the transaction is in RunningRounds
	rr := runningRounds[round.Hash()]
	require.NotNil(t, rr)
	txHashs := rr.Transactions[proposer.Address()]
	require.Equal(t, tx.GetHash(), txHashs[0])

	ballotSIGN1 := GenerateBallot(t, proposer, round, tx, ballot.StateSIGN, nodes[1])
	ballotSIGN1.SetVote(ballot.StateSIGN, ballot.VotingNO)
	err = ReceiveBallot(t, nr, ballotSIGN1)
	require.Nil(t, err)

	ballotSIGN2 := GenerateBallot(t, proposer, round, tx, ballot.StateSIGN, nodes[2])
	ballotSIGN2.SetVote(ballot.StateSIGN, ballot.VotingNO)
	err = ReceiveBallot(t, nr, ballotSIGN2)
	require.Nil(t, err)

	ballotSIGN3 := GenerateBallot(t, proposer, round, tx, ballot.StateSIGN, nodes[3])
	ballotSIGN3.SetVote(ballot.StateSIGN, ballot.VotingNO)
	err = ReceiveBallot(t, nr, ballotSIGN3)
	require.Nil(t, err)

	ballotSIGN4 := GenerateBallot(t, proposer, round, tx, ballot.StateSIGN, nodes[4])
	ballotSIGN4.SetVote(ballot.StateSIGN, ballot.VotingNO)
	err = ReceiveBallot(t, nr, ballotSIGN4)
	require.Nil(t, err)

	<-recv

	require.Equal(t, 3, len(cm.Messages()))

	init, sign, accept := 0, 0, 0

	for _, message := range cm.Messages() {
		b, ok := message.(ballot.Ballot)
		if !ok {
			continue
		}
		switch b.State() {
		case ballot.StateINIT:
			init++
			require.Equal(t, ballot.VotingYES, b.Vote())
		case ballot.StateSIGN:
			sign++
		case ballot.StateACCEPT:
			require.Equal(t, ballot.VotingNO, b.Vote())
			accept++
		}
	}

	require.Equal(t, 1, init)
	require.Equal(t, 0, sign)
	require.Equal(t, 1, accept)
}

// //	TestStateTransitCauseEXPBallot indicates the following:
// //		1. Proceed for one round.
// //		2. The node is the proposer of this round.
// //		3. There are 5 nodes and threshold is 4.
// //		4. The node receives the SIGN messages with EXPIRED in order from the other four validator nodes.
// //		5. The ballots which is included in the node cannot exceeds the threshold as `YES`, then the node broadcast ballot with `EXPIRED`

// func TestStateTransitCauseEXPBallot(t *testing.T) {
// 	recv := make(chan struct{})
// 	nr, nodes, cm := createNodeRunnerForTesting(5, conf, recv)

// 	tx, txByte := GetTransaction(t)

// 	message := common.NetworkMessage{Type: common.TransactionMessage, Data: txByte}

// 	conf := consensus.NewISAACConfiguration()
// 	conf.TimeoutINIT = time.Hour
// 	conf.TimeoutSIGN = time.Hour
// 	conf.TimeoutACCEPT = time.Hour
// 	conf.BlockTime = time.Hour

// 	nr.Consensus().SetLatestConsensusedBlock(genesisBlock)
// 	require.Equal(t, uint64(1), nr.Consensus().LatestConfirmedBlock.Height)
// 	require.Equal(t, uint64(0), nr.Consensus().LatestConfirmedBlock.TotalTxs)

// 	var err error
// 	err = nr.handleTransaction(message)

// 	require.Nil(t, err)
// 	require.True(t, nr.Consensus().TransactionPool.Has(tx.GetHash()))

// 	nr.StartStateManager()
// 	time.Sleep(200 * time.Millisecond)

// 	require.Nil(t, err)
// 	round := round.Round{
// 		Number:      0,
// 		BlockHeight: nr.Consensus().LatestConfirmedBlock.Height,
// 		BlockHash:   nr.Consensus().LatestConfirmedBlock.Hash,
// 		TotalTxs:    nr.Consensus().LatestConfirmedBlock.TotalTxs,
// 	}
// 	proposer := nr.localNode
// 	runningRounds := nr.Consensus().RunningRounds

// 	// Check that the transaction is in RunningRounds
// 	rr := runningRounds[round.Hash()]
// 	require.NotNil(t, rr)
// 	txHashs := rr.Transactions[proposer.Address()]
// 	require.Equal(t, tx.GetHash(), txHashs[0])

// 	ballotSIGN1 := GenerateBallot(t, proposer, round, tx, ballot.StateSIGN, nodes[1])
// 	ballotSIGN1.SetVote(ballot.StateSIGN, ballot.VotingEXP)
// 	err = ReceiveBallot(t, nr, ballotSIGN1)
// 	require.Nil(t, err)

// 	ballotSIGN2 := GenerateBallot(t, proposer, round, tx, ballot.StateSIGN, nodes[2])
// 	ballotSIGN2.SetVote(ballot.StateSIGN, ballot.VotingEXP)
// 	err = ReceiveBallot(t, nr, ballotSIGN2)
// 	require.Nil(t, err)

// 	ballotSIGN3 := GenerateBallot(t, proposer, round, tx, ballot.StateSIGN, nodes[3])
// 	ballotSIGN3.SetVote(ballot.StateSIGN, ballot.VotingEXP)
// 	err = ReceiveBallot(t, nr, ballotSIGN3)
// 	require.Nil(t, err)

// 	time.Sleep(100 * time.Millisecond)

// 	require.Equal(t, 3, len(cm.Messages()))

// 	init, sign, accept := 0, 0, 0

// 	for _, message := range cm.Messages() {
// 		b, ok := message.(ballot.Ballot)
// 		if !ok {
// 			continue
// 		}
// 		switch b.State() {
// 		case ballot.StateINIT:
// 			init++
// 			require.Equal(t, ballot.VotingYES, b.Vote())
// 		case ballot.StateSIGN:
// 			sign++
// 		case ballot.StateACCEPT:
// 			require.Equal(t, ballot.VotingEXP, b.Vote())
// 			accept++
// 		}
// 	}

// 	require.Equal(t, 1, init)
// 	require.Equal(t, 0, sign)
// 	require.Equal(t, 1, accept)
// }

// func TestStateTransitSIGNTimeoutACCEPTBallotProposer(t *testing.T) {
// 	recv := make(chan struct{})
// 	nr, nodes, cm := createNodeRunnerForTesting(5, conf, recv)

// 	tx, txByte := GetTransaction(t)

// 	message := common.NetworkMessage{Type: common.TransactionMessage, Data: txByte}

// 	conf := consensus.NewISAACConfiguration()
// 	conf.TimeoutINIT = time.Hour
// 	conf.TimeoutSIGN = time.Millisecond
// 	conf.TimeoutACCEPT = time.Hour
// 	conf.BlockTime = 200 * time.Millisecond

// 	// `nr` is proposer's runner
// 	proposer := nr.localNode

// 	nr.Consensus().SetLatestConsensusedBlock(genesisBlock)
// 	require.Equal(t, uint64(1), nr.Consensus().LatestConfirmedBlock.Height)
// 	require.Equal(t, uint64(0), nr.Consensus().LatestConfirmedBlock.TotalTxs)

// 	var err error
// 	err = nr.handleTransaction(message)

// 	require.Nil(t, err)
// 	require.True(t, nr.Consensus().TransactionPool.Has(tx.GetHash()))

// 	nr.StartStateManager()

// 	time.Sleep(200 * time.Millisecond)

// 	init, sign, accept, ballots := 0, 0, 0, 0
// 	for _, message := range cm.Messages() {
// 		b, ok := message.(ballot.Ballot)
// 		if !ok {
// 			continue
// 		}
// 		ballots++
// 		switch b.State() {
// 		case ballot.StateINIT:
// 			init++
// 			require.Equal(t, ballot.VotingYES, b.Vote())
// 		case ballot.StateSIGN:
// 			sign++
// 		case ballot.StateACCEPT:
// 			accept++
// 			require.Equal(t, ballot.VotingEXP, b.Vote())
// 		}
// 	}
// 	require.Equal(t, 2, ballots)
// 	require.Equal(t, 1, init)
// 	require.Equal(t, 0, sign)
// 	require.Equal(t, 1, accept)

// 	round := round.Round{
// 		Number:      0,
// 		BlockHeight: nr.Consensus().LatestConfirmedBlock.Height,
// 		BlockHash:   nr.Consensus().LatestConfirmedBlock.Hash,
// 		TotalTxs:    nr.Consensus().LatestConfirmedBlock.TotalTxs,
// 	}
// 	runningRounds := nr.Consensus().RunningRounds

// 	// Check that the transaction is in RunningRounds
// 	ballotACCEPT1 := GenerateBallot(t, proposer, round, tx, ballot.StateACCEPT, nodes[1])
// 	err = ReceiveBallot(t, nr, ballotACCEPT1)
// 	require.Nil(t, err)

// 	ballotACCEPT2 := GenerateBallot(t, proposer, round, tx, ballot.StateACCEPT, nodes[2])
// 	err = ReceiveBallot(t, nr, ballotACCEPT2)
// 	require.Nil(t, err)

// 	ballotACCEPT3 := GenerateBallot(t, proposer, round, tx, ballot.StateACCEPT, nodes[3])
// 	err = ReceiveBallot(t, nr, ballotACCEPT3)
// 	require.Nil(t, err)

// 	ballotACCEPT4 := GenerateBallot(t, proposer, round, tx, ballot.StateACCEPT, nodes[4])
// 	err = ReceiveBallot(t, nr, ballotACCEPT4)
// 	require.EqualError(t, err, "ballot got consensus and will be stored")

// 	rr := runningRounds[round.Hash()]
// 	require.Nil(t, rr)

// 	lastConfirmedBlock := nr.Consensus().LatestConfirmedBlock
// 	require.Equal(t, proposer.Address(), lastConfirmedBlock.Proposer)
// 	require.Equal(t, uint64(2), lastConfirmedBlock.Height)
// 	require.Equal(t, uint64(1), lastConfirmedBlock.TotalTxs)
// 	require.Equal(t, 1, len(lastConfirmedBlock.Transactions))
// 	require.Equal(t, tx.GetHash(), lastConfirmedBlock.Transactions[0])
// }

// func TestStateTransitSIGNBallotACCEPTTimeoutProposer(t *testing.T) {
// 	recv := make(chan struct{})
// 	nr, nodes, cm := createNodeRunnerForTesting(5, conf, recv)

// 	tx, txByte := GetTransaction(t)

// 	message := common.NetworkMessage{Type: common.TransactionMessage, Data: txByte}

// 	conf := consensus.NewISAACConfiguration()
// 	conf.TimeoutINIT = time.Hour
// 	conf.TimeoutSIGN = time.Hour
// 	conf.TimeoutACCEPT = time.Millisecond
// 	conf.BlockTime = 200 * time.Millisecond

// 	// `nr` is proposer's runner
// 	proposer := nr.localNode

// 	nr.Consensus().SetLatestConsensusedBlock(genesisBlock)
// 	require.Equal(t, uint64(1), nr.Consensus().LatestConfirmedBlock.Height)
// 	require.Equal(t, uint64(0), nr.Consensus().LatestConfirmedBlock.TotalTxs)

// 	var err error
// 	err = nr.handleTransaction(message)

// 	require.Nil(t, err)
// 	require.True(t, nr.Consensus().TransactionPool.Has(tx.GetHash()))

// 	nr.StartStateManager()

// 	time.Sleep(200 * time.Millisecond)
// 	// Generate proposed ballot in nr
// 	r := round.Round{
// 		Number:      0,
// 		BlockHeight: nr.Consensus().LatestConfirmedBlock.Height,
// 		BlockHash:   nr.Consensus().LatestConfirmedBlock.Hash,
// 		TotalTxs:    nr.Consensus().LatestConfirmedBlock.TotalTxs,
// 	}

// 	runningRounds := nr.Consensus().RunningRounds
// 	require.Equal(t, 1, len(runningRounds))

// 	ballotSIGN1 := GenerateBallot(t, proposer, r, tx, ballot.StateSIGN, nodes[1])
// 	err = ReceiveBallot(t, nr, ballotSIGN1)
// 	require.Nil(t, err)

// 	ballotSIGN2 := GenerateBallot(t, proposer, r, tx, ballot.StateSIGN, nodes[2])
// 	err = ReceiveBallot(t, nr, ballotSIGN2)
// 	require.Nil(t, err)

// 	ballotSIGN3 := GenerateBallot(t, proposer, r, tx, ballot.StateSIGN, nodes[3])
// 	err = ReceiveBallot(t, nr, ballotSIGN3)
// 	require.Nil(t, err)

// 	runningRounds = nr.Consensus().RunningRounds
// 	require.Equal(t, 1, len(runningRounds))

// 	time.Sleep(200 * time.Millisecond)
// 	require.Equal(t, 2, len(runningRounds))

// 	rr, ok := runningRounds[r.Hash()]
// 	require.True(t, ok)
// 	require.NotNil(t, rr)
// 	require.Equal(t, 3, len(rr.Voted[proposer.Address()].GetResult(ballot.StateSIGN)))

// 	ballotSIGN4 := GenerateBallot(t, proposer, r, tx, ballot.StateSIGN, nodes[4])
// 	err = ReceiveBallot(t, nr, ballotSIGN4)
// 	require.Nil(t, err)

// 	rr, ok = runningRounds[r.Hash()]
// 	require.True(t, ok)
// 	require.NotNil(t, rr)
// 	require.Equal(t, 4, len(rr.Voted[proposer.Address()].GetResult(ballot.StateSIGN)))

// 	time.Sleep(200 * time.Millisecond)
// 	require.Equal(t, 2, len(runningRounds))

// 	nextRound := round.Round{
// 		Number:      1,
// 		BlockHeight: nr.Consensus().LatestConfirmedBlock.Height,
// 		BlockHash:   nr.Consensus().LatestConfirmedBlock.Hash,
// 		TotalTxs:    nr.Consensus().LatestConfirmedBlock.TotalTxs,
// 	}

// 	nextRr, ok := runningRounds[nextRound.Hash()]
// 	require.True(t, ok)
// 	require.NotNil(t, nextRr)
// 	require.Equal(t, 0, len(nextRr.Voted[proposer.Address()].GetResult(ballot.StateSIGN)))

// 	time.Sleep(200 * time.Millisecond)

// 	init, sign, accept, ballots := 0, 0, 0, 0
// 	for _, message := range cm.Messages() {
// 		b, ok := message.(ballot.Ballot)
// 		if !ok {
// 			continue
// 		}
// 		ballots++
// 		switch b.State() {
// 		case ballot.StateINIT:
// 			init++
// 			require.Equal(t, ballot.VotingYES, b.Vote())
// 		case ballot.StateSIGN:
// 			sign++
// 		case ballot.StateACCEPT:
// 			accept++
// 		}
// 	}
// 	require.Equal(t, 4, ballots)
// 	require.Equal(t, 2, init) // INIT(Send B(`INIT`, `YES`) cause Proposer itself) -> SIGN(Receive 4 Ballots) -> ACCEPT(Timeout) -> INIT(Send B(`INIT`, `YES`)
// 	require.Equal(t, 0, sign)
// 	require.Equal(t, 2, accept)
// }

// func TestStateTransitTwoBlocks(t *testing.T) {
// 	conf := consensus.NewISAACConfiguration()
// 	conf.TimeoutINIT = time.Hour
// 	conf.TimeoutSIGN = time.Hour
// 	conf.TimeoutACCEPT = time.Hour
// 	conf.BlockTime = 1 * time.Millisecond

// 	recv := make(chan struct{})

// 	nr, nodes, cm := createNodeRunnerForTesting(5, conf, recv)

// 	tx, txByte := GetTransaction(t)

// 	message := common.NetworkMessage{Type: common.TransactionMessage, Data: txByte}

// 	// `nr` is proposer's runner
// 	nr.SetProposerCalculator(&SelfProposerThenNotProposer{})
// 	proposer := nr.localNode

// 	nr.Consensus().SetLatestConsensusedBlock(genesisBlock)
// 	require.Equal(t, uint64(1), nr.Consensus().LatestConfirmedBlock.Height)
// 	require.Equal(t, uint64(0), nr.Consensus().LatestConfirmedBlock.TotalTxs)

// 	var err error
// 	err = nr.handleTransaction(message)

// 	require.Nil(t, err)
// 	require.True(t, nr.Consensus().TransactionPool.Has(tx.GetHash()))

// 	nr.StartStateManager()

// 	time.Sleep(200 * time.Millisecond)

// 	init, sign, accept, ballots := 0, 0, 0, 0
// 	for _, message := range cm.Messages() {
// 		b, ok := message.(ballot.Ballot)
// 		if !ok {
// 			continue
// 		}
// 		ballots++
// 		switch b.State() {
// 		case ballot.StateINIT:
// 			init++
// 			require.Equal(t, ballot.VotingYES, b.Vote())
// 		case ballot.StateSIGN:
// 			sign++
// 		case ballot.StateACCEPT:
// 			accept++
// 		}
// 	}
// 	require.Equal(t, 1, ballots)
// 	require.Equal(t, 1, init)
// 	require.Equal(t, 0, sign)
// 	require.Equal(t, 0, accept)

// 	r := round.Round{
// 		Number:      0,
// 		BlockHeight: nr.Consensus().LatestConfirmedBlock.Height,
// 		BlockHash:   nr.Consensus().LatestConfirmedBlock.Hash,
// 		TotalTxs:    nr.Consensus().LatestConfirmedBlock.TotalTxs,
// 	}
// 	runningRounds := nr.Consensus().RunningRounds

// 	// Check that the transaction is in RunningRounds
// 	ballotACCEPT1 := GenerateBallot(t, proposer, r, tx, ballot.StateACCEPT, nodes[1])
// 	err = ReceiveBallot(t, nr, ballotACCEPT1)
// 	require.Nil(t, err)

// 	ballotACCEPT2 := GenerateBallot(t, proposer, r, tx, ballot.StateACCEPT, nodes[2])
// 	err = ReceiveBallot(t, nr, ballotACCEPT2)
// 	require.Nil(t, err)

// 	ballotACCEPT3 := GenerateBallot(t, proposer, r, tx, ballot.StateACCEPT, nodes[3])
// 	err = ReceiveBallot(t, nr, ballotACCEPT3)
// 	require.Nil(t, err)

// 	ballotACCEPT4 := GenerateBallot(t, proposer, r, tx, ballot.StateACCEPT, nodes[4])
// 	err = ReceiveBallot(t, nr, ballotACCEPT4)
// 	require.EqualError(t, err, "ballot got consensus and will be stored")

// 	rr := runningRounds[r.Hash()]
// 	require.Nil(t, rr)

// 	lastConfirmedBlock := nr.Consensus().LatestConfirmedBlock
// 	require.Equal(t, proposer.Address(), lastConfirmedBlock.Proposer)
// 	require.Equal(t, uint64(2), lastConfirmedBlock.Height)
// 	require.Equal(t, uint64(1), lastConfirmedBlock.TotalTxs)
// 	require.Equal(t, 1, len(lastConfirmedBlock.Transactions))
// 	require.Equal(t, tx.GetHash(), lastConfirmedBlock.Transactions[0])
// 	require.Equal(t, 0, len(nr.Consensus().TransactionPool.AvailableTransactions(NewISAACConfiguration())))

// 	// New Height with empty transactions
// 	r = round.Round{
// 		Number:      0,
// 		BlockHeight: nr.Consensus().LatestConfirmedBlock.Height,
// 		BlockHash:   nr.Consensus().LatestConfirmedBlock.Hash,
// 		TotalTxs:    nr.Consensus().LatestConfirmedBlock.TotalTxs,
// 	}

// 	require.Equal(t, uint64(2), r.BlockHeight)

// 	calculator := SelfProposerThenNotProposer{}
// 	nextProposerAddress := calculator.Calculate(nr, 2, 0)

// 	var nextProposer *node.LocalNode = proposer
// 	for _, nr := range nodeRunners {
// 		if nr.localNode.Address() == nextProposerAddress {
// 			nextProposer = nr.localNode
// 			continue
// 		}
// 	}

// 	ballotProposeINIT := GenerateEmptyTxBallot(t, nextProposer, r, ballot.StateINIT, nextProposer)
// 	require.Equal(t, ballot.StateINIT, nr.isaacStateManager.State().ballotState)
// 	err = ReceiveBallot(t, nr, ballotProposeINIT)
// 	require.Nil(t, err)

// 	time.Sleep(100 * time.Millisecond)

// 	require.Equal(t, ballot.StateSIGN, nr.isaacStateManager.State().ballotState)

// 	ballotSIGN1 := GenerateEmptyTxBallot(t, nextProposer, r, ballot.StateSIGN, nodes[1])
// 	err = ReceiveBallot(t, nr, ballotSIGN1)
// 	require.Nil(t, err)

// 	ballotSIGN2 := GenerateEmptyTxBallot(t, nextProposer, r, ballot.StateSIGN, nodes[2])
// 	err = ReceiveBallot(t, nr, ballotSIGN2)
// 	require.Nil(t, err)

// 	ballotSIGN3 := GenerateEmptyTxBallot(t, nextProposer, r, ballot.StateSIGN, nodes[3])
// 	err = ReceiveBallot(t, nr, ballotSIGN3)
// 	require.Nil(t, err)

// 	ballotSIGN4 := GenerateEmptyTxBallot(t, nextProposer, r, ballot.StateSIGN, nodes[4])
// 	err = ReceiveBallot(t, nr, ballotSIGN4)
// 	require.Nil(t, err)

// 	time.Sleep(200 * time.Millisecond)

// 	ballotACCEPT1 = GenerateEmptyTxBallot(t, nextProposer, r, ballot.StateACCEPT, nodes[1])
// 	err = ReceiveBallot(t, nr, ballotACCEPT1)
// 	require.Nil(t, err)

// 	ballotACCEPT2 = GenerateEmptyTxBallot(t, nextProposer, r, ballot.StateACCEPT, nodes[2])
// 	err = ReceiveBallot(t, nr, ballotACCEPT2)
// 	require.Nil(t, err)

// 	ballotACCEPT3 = GenerateEmptyTxBallot(t, nextProposer, r, ballot.StateACCEPT, nodes[3])
// 	err = ReceiveBallot(t, nr, ballotACCEPT3)
// 	require.Nil(t, err)

// 	ballotACCEPT4 = GenerateEmptyTxBallot(t, nextProposer, r, ballot.StateACCEPT, nodes[4])
// 	err = ReceiveBallot(t, nr, ballotACCEPT4)
// 	require.EqualError(t, err, "ballot got consensus and will be stored")

// 	time.Sleep(200 * time.Millisecond)

// 	rr = runningRounds[r.Hash()]
// 	require.Nil(t, rr)

// 	lastConfirmedBlock = nr.Consensus().LatestConfirmedBlock
// 	require.Equal(t, uint64(3), lastConfirmedBlock.Height)
// 	require.Equal(t, uint64(1), lastConfirmedBlock.TotalTxs)
// 	require.Equal(t, 0, len(lastConfirmedBlock.Transactions))
// 	require.Equal(t, 0, len(nr.Consensus().TransactionPool.AvailableTransactions(NewISAACConfiguration())))
// }
