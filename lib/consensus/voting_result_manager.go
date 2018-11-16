package consensus

import (
	"fmt"

	"boscoin.io/sebak/lib/ballot"
)

type VotingResultManager struct {
	sourceHeight map[ /* Node.Address() */ string]heightRound
	heightSource map[heightRound]set
}

type set map[ /* Node.Address() */ string]struct{}

func NewVotingResultManager() *VotingResultManager {
	return &VotingResultManager{
		sourceHeight: map[string]heightRound{},
		heightSource: map[heightRound]set{},
	}
}

type heightRound struct {
	height uint64
	round  uint64
}

func (current heightRound) isLaterThan(target heightRound) bool {
	if current.height == target.height {
		return current.round > target.round
	} else {
		return current.height > target.height
	}
}

func (hm *VotingResultManager) update(source string, hr heightRound) {
	current := hm.sourceHeight[source]
	if current.isLaterThan(hr) {
		return
	}

	delete(hm.heightSource[current], source)

	if len(hm.heightSource[current]) == 0 {
		delete(hm.heightSource, current)
	}

	hm.sourceHeight[source] = hr
	if hm.heightSource[hr] == nil {
		hm.heightSource[hr] = set{}
	}
	hm.heightSource[hr][source] = struct{}{}
}

// getSyncInfo gets the height it needs to sync.
// It returns height, node list and error.
// The height is the smallest height above the threshold.
// The node list is the nodes that sent the ballot when the threshold is exceeded.
func (hm *VotingResultManager) getSyncInfo(b ballot.Ballot, threshold int) (uint64, []string, error) {
	hr := heightRound{height: b.VotingBasis().Height, round: b.VotingBasis().Round}
	hm.update(b.Source(), hr)

	if len(hm.sourceHeight) < threshold {
		return 1, []string{}, fmt.Errorf("could not find enough nodes (threshold=%d) above", threshold)
	}

	if len(hm.heightSource[hr]) < threshold {
		return 1, []string{}, fmt.Errorf("this ballot did not exceed the threshold(=%d)", threshold)
	}

	for height, sourceSet := range hm.heightSource {
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
