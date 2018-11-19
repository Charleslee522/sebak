package consensus

import (
	"fmt"

	"boscoin.io/sebak/lib/ballot"
	"boscoin.io/sebak/lib/voting"
)

type VotingResultManager struct {
	recentVoting  map[ /* Node.Address() */ string]votingResult
	votingResults map[votingResult]set
}

type votingResult struct {
	height     uint64
	round      uint64
	state      ballot.State
	votingHole voting.Hole
}

type set map[ /* Node.Address() */ string]struct{}

func NewVotingResultManager() *VotingResultManager {
	return &VotingResultManager{
		recentVoting:  map[string]votingResult{},
		votingResults: map[votingResult]set{},
	}
}

func (current votingResult) isLaterThan(target votingResult) bool {
	if current.height == target.height {
		return current.round > target.round
	} else {
		return current.height > target.height
	}
}

func (hm *VotingResultManager) update(source string, hr votingResult) {
	current := hm.recentVoting[source]
	if current.isLaterThan(hr) {
		return
	}

	delete(hm.votingResults[current], source)

	if len(hm.votingResults[current]) == 0 {
		delete(hm.votingResults, current)
	}

	hm.recentVoting[source] = hr
	if hm.votingResults[hr] == nil {
		hm.votingResults[hr] = set{}
	}
	hm.votingResults[hr][source] = struct{}{}
}

// getSyncInfo gets the height it needs to sync.
// It returns height, node list and error.
// The height is the smallest height above the threshold.
// The node list is the nodes that sent the ballot when the threshold is exceeded.
func (hm *VotingResultManager) getSyncInfo(b ballot.Ballot, threshold int) (uint64, []string, error) {
	hr := votingResult{
		height:     b.VotingBasis().Height,
		round:      b.VotingBasis().Round,
		state:      b.State(), // ACCEPT always
		votingHole: b.Vote(),  // YES always
	}
	hm.update(b.Source(), hr)

	if len(hm.recentVoting) < threshold {
		return 1, []string{}, fmt.Errorf("could not find enough nodes (threshold=%d) above", threshold)
	}

	if len(hm.votingResults[hr]) < threshold {
		return 1, []string{}, fmt.Errorf("this ballot did not exceed the threshold(=%d)", threshold)
	}

	for height, sourceSet := range hm.votingResults {
		if len(sourceSet) >= threshold {
			sources := []string{}
			for source := range sourceSet {
				sources = append(sources, source)
			}
			return height.height, sources, nil
		}
	}

	return 1, []string{}, fmt.Errorf("could not find enough nodes (threshold=%d) above", threshold)
}
