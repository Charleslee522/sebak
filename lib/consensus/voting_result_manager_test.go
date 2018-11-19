package consensus

import (
	"testing"

	"boscoin.io/sebak/lib/ballot"
	"boscoin.io/sebak/lib/voting"
	"github.com/stretchr/testify/require"
)

func update(hm *VotingResultManager, source string, height, round uint64) {
	hm.update(source, votingResult{height: height, round: round, state: ballot.StateACCEPT, votingHole: voting.YES})
}

func TestUpdate(t *testing.T) {
	hm := NewVotingResultManager()

	sources := []string{
		"nodeA",
		"nodeB",
		"nodeC",
		"nodeD",
		"nodeE",
	}

	update(hm, sources[0], 9, 0)
	update(hm, sources[0], 10, 0)
	update(hm, sources[1], 10, 0)

	require.Equal(t, 2, len(hm.recentVoting))
	require.Equal(t, 1, len(hm.votingResults))

	update(hm, sources[2], 10, 0)
	update(hm, sources[3], 10, 0)

	require.Equal(t, 4, len(hm.recentVoting))
	require.Equal(t, 1, len(hm.votingResults))

	update(hm, sources[2], 10, 1)
	update(hm, sources[3], 10, 1)

	require.Equal(t, 4, len(hm.recentVoting))
	require.Equal(t, 2, len(hm.votingResults))

	update(hm, sources[0], 10, 1)
	update(hm, sources[1], 10, 1)

	require.Equal(t, 1, len(hm.votingResults))

	update(hm, sources[0], 9, 10)
	require.Equal(t, 4, len(hm.recentVoting))
	require.Equal(t, 1, len(hm.votingResults))
}

func TestGetSyncInfoNormal(t *testing.T) {
	hm := NewVotingResultManager()

	valids := []string{
		"nodeA",
		"nodeB",
		"nodeC",
		"nodeE",
	}

	invalids := []string{
		"nodeD",
	}

	update(hm, valids[0], 10, 0)
	update(hm, valids[1], 10, 0)
	update(hm, valids[2], 10, 0)
	update(hm, invalids[0], 10, 0)
	update(hm, valids[3], 10, 0)

	b := *ballot.NewBallot(valids[3], valids[3], voting.Basis{Height: 10, Round: 0}, []string{})
	b.SetVote(ballot.StateACCEPT, voting.YES)

	height, nodeAddrs, err := hm.getSyncInfo(b, 4)
	require.NoError(t, err)
	require.Equal(t, uint64(10), height)
	require.True(t, contains(nodeAddrs, valids[0]))
	require.True(t, contains(nodeAddrs, valids[1]))
	require.True(t, contains(nodeAddrs, valids[2]))
	require.True(t, contains(nodeAddrs, valids[3]))
}

func TestGetSyncInfoBeforeFull(t *testing.T) {
	hm := NewVotingResultManager()

	valids := []string{
		"nodeA",
		"nodeB",
		"nodeC",
		"nodeE",
	}

	update(hm, valids[0], 10, 0)
	update(hm, valids[1], 10, 0)
	update(hm, valids[2], 10, 0)
	update(hm, valids[3], 10, 0)

	b := *ballot.NewBallot(valids[3], valids[3], voting.Basis{Height: 10, Round: 0}, []string{})
	b.SetVote(ballot.StateACCEPT, voting.YES)

	height, nodeAddrs, err := hm.getSyncInfo(b, 4)
	require.NoError(t, err)
	require.Equal(t, uint64(10), height)
	require.True(t, contains(nodeAddrs, valids[0]))
	require.True(t, contains(nodeAddrs, valids[1]))
	require.True(t, contains(nodeAddrs, valids[2]))
	require.True(t, contains(nodeAddrs, valids[3]))
}

func TestGetSyncInfoInvalidBallot(t *testing.T) {
	hm := NewVotingResultManager()

	valids := []string{
		"nodeA",
		"nodeB",
		"nodeC",
		"nodeD",
		"nodeE",
	}

	update(hm, valids[0], 10, 0)
	update(hm, valids[1], 10, 0)
	update(hm, valids[2], 10, 0)
	update(hm, valids[3], 10, 0)
	update(hm, valids[4], 10, 1)

	b := *ballot.NewBallot(valids[4], valids[4], voting.Basis{Height: 10, Round: 1}, []string{})
	b.SetVote(ballot.StateACCEPT, voting.YES)

	_, _, err := hm.getSyncInfo(b, 4)
	require.Error(t, err)
}

func TestGetSyncInfoFoundSmallestHeight(t *testing.T) {
	hm := NewVotingResultManager()
	nodes := []string{
		"nodeA",
		"nodeB",
		"nodeC",
		"nodeD",
		"nodeE",
	}

	update(hm, nodes[0], 32, 0)
	update(hm, nodes[1], 33, 0)
	update(hm, nodes[2], 33, 0)
	update(hm, nodes[3], 34, 0)
	update(hm, nodes[4], 35, 0)

	b := *ballot.NewBallot(nodes[4], nodes[4], voting.Basis{Height: 35, Round: 0}, []string{})
	b.SetVote(ballot.StateACCEPT, voting.YES)

	_, _, err := hm.getSyncInfo(b, 4)
	require.Error(t, err)

	update(hm, nodes[0], 36, 0)
	update(hm, nodes[1], 36, 0)
	update(hm, nodes[2], 36, 1)
	update(hm, nodes[3], 36, 1)
	update(hm, nodes[4], 36, 1)
	b = *ballot.NewBallot(nodes[4], nodes[4], voting.Basis{Height: 36, Round: 1}, []string{})
	b.SetVote(ballot.StateACCEPT, voting.YES)

	_, _, err = hm.getSyncInfo(b, 4)
	require.Error(t, err)

	update(hm, nodes[0], 36, 1)
	b = *ballot.NewBallot(nodes[4], nodes[4], voting.Basis{Height: 36, Round: 1}, []string{})
	b.SetVote(ballot.StateACCEPT, voting.YES)

	height, nodeAddrs, err := hm.getSyncInfo(b, 4)
	require.NoError(t, err)
	require.Equal(t, uint64(36), height)

	require.True(t, contains(nodeAddrs, nodes[0]))
	require.True(t, contains(nodeAddrs, nodes[2]))
	require.True(t, contains(nodeAddrs, nodes[3]))
	require.True(t, contains(nodeAddrs, nodes[4]))
}

func TestGetSyncInfoGenesis(t *testing.T) {
	hm := NewVotingResultManager()
	nodes := []string{
		"nodeA",
		"nodeB",
		"nodeC",
		"nodeE",
	}

	update(hm, nodes[0], 1, 0)
	update(hm, nodes[1], 1, 0)
	update(hm, nodes[2], 1, 0)
	update(hm, nodes[3], 1, 0)

	b := *ballot.NewBallot(nodes[3], nodes[3], voting.Basis{Height: 1, Round: 0}, []string{})
	b.SetVote(ballot.StateACCEPT, voting.YES)

	height, nodeAddrs, err := hm.getSyncInfo(b, 4)
	require.NoError(t, err)
	require.Equal(t, uint64(1), height)
	require.Equal(t, 4, len(nodeAddrs))
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
