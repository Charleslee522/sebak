package consensus

import (
	"fmt"

	"boscoin.io/sebak/lib/ballot"
	"boscoin.io/sebak/lib/voting"
)

type VotingResultManager struct {
	recentVoting  map[ /* Node.Address() */ string]votingResult
	votingResults map[votingResult]set
	policy        voting.ThresholdPolicy
}

type votingResult struct {
	height     uint64
	round      uint64
	state      ballot.State
	votingHole voting.Hole
}

func NewVotingResult(b ballot.Ballot) votingResult {
	return votingResult{
		height:     b.VotingBasis().Height,
		round:      b.VotingBasis().Round,
		state:      b.State(), // ACCEPT always
		votingHole: b.Vote(),  // YES always
	}
}

type set map[ /* Node.Address() */ string]struct{}

func NewVotingResultManager(p voting.ThresholdPolicy) *VotingResultManager {
	return &VotingResultManager{
		recentVoting:  map[string]votingResult{},
		votingResults: map[votingResult]set{},
		policy:        p,
	}
}

func (current votingResult) isLaterThan(target votingResult) bool {
	if current.height == target.height {
		return current.round > target.round
	} else {
		return current.height > target.height
	}
}

func (vrm *VotingResultManager) IsNew(b ballot.Ballot) bool {
	if sources, ok := vrm.votingResults[NewVotingResult(b)]; ok {
		_, ok := sources[b.Source()]
		return ok
	}
	return false
}

func (vrm *VotingResultManager) IsVoted(b ballot.Ballot) bool {
	return !vrm.IsNew(b)
}

func (vrm *VotingResultManager) Vote(b ballot.Ballot) {
	vr := NewVotingResult(b)
	source := b.Source()

	current := vrm.recentVoting[source]
	if current.isLaterThan(vr) {
		return
	}

	delete(vrm.votingResults[current], source)

	if len(vrm.votingResults[current]) == 0 {
		delete(vrm.votingResults, current)
	}

	vrm.recentVoting[source] = vr
	if vrm.votingResults[vr] == nil {
		vrm.votingResults[vr] = set{}
	}
	vrm.votingResults[vr][source] = struct{}{}
}

func (vrm *VotingResultManager) CanGetVotingResult(b ballot.Ballot) (RoundVoteResult, voting.Hole, bool) {
	threshold := vrm.policy.Threshold()
	if threshold < 1 {
		return RoundVoteResult{}, voting.NOTYET, false
	}

	return RoundVoteResult{}, voting.NOTYET, false
}

// getSyncInfo gets the height it needs to sync.
// It returns height, node list and error.
// The height is the smallest height above the threshold.
// The node list is the nodes that sent the ballot when the threshold is exceeded.
func (vrm *VotingResultManager) getSyncInfo(b ballot.Ballot) (uint64, []string, error) {
	vrm.Vote(b)
	threshold := vrm.policy.Threshold()

	if len(vrm.recentVoting) < threshold {
		return 1, []string{}, fmt.Errorf("could not find enough nodes (threshold=%d) above", threshold)
	}

	if !vrm.exceed(b, threshold) {
		return 1, []string{}, fmt.Errorf("this ballot did not exceed the threshold(=%d)", threshold)
	}

	for height, sourceSet := range vrm.votingResults {
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

func (vrm *VotingResultManager) exceed(b ballot.Ballot, threshold int) bool {
	vr := votingResult{
		height:     b.VotingBasis().Height,
		round:      b.VotingBasis().Round,
		state:      b.State(), // ACCEPT always
		votingHole: b.Vote(),  // YES always
	}
	return len(vrm.votingResults[vr]) >= threshold
}
