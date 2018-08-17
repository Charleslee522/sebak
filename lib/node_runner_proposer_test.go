package sebak

import (
	"reflect"
	"testing"

	"boscoin.io/sebak/lib/network"
)

func TestNodeRunnersHaveSameProposers(t *testing.T) {
	defer sebaknetwork.CleanUpMemoryNetwork()

	numberOfNodes := 3
	nodeRunners := createTestNodeRunnerWithReady(numberOfNodes)

	nr0 := nodeRunners[0]
	nr1 := nodeRunners[1]
	nr2 := nodeRunners[2]

	var maximumBlockHeight uint64 = 3
	var maximumRoundNumber uint64 = 3

	proposers0 := make([]string, maximumBlockHeight*maximumRoundNumber)
	proposers1 := make([]string, maximumBlockHeight*maximumRoundNumber)
	proposers2 := make([]string, maximumBlockHeight*maximumRoundNumber)

	for i := uint64(0); i < maximumBlockHeight; i++ {
		for j := uint64(0); j < maximumRoundNumber; j++ {
			proposers0[i*maximumRoundNumber+j] = nr0.CalculateProposer(i, j)
			proposers1[i*maximumRoundNumber+j] = nr1.CalculateProposer(i, j)
			proposers2[i*maximumRoundNumber+j] = nr2.CalculateProposer(i, j)
		}
	}

	if !reflect.DeepEqual(proposers0, proposers1) {
		t.Error("failed to have same proposers. nr0, nr1.")
	}
	if !reflect.DeepEqual(proposers0, proposers2) {
		t.Error("failed to have same proposers. nr0, nr2.")
	}
	if !reflect.DeepEqual(proposers1, proposers2) {
		t.Error("failed to have same proposers. nr1, nr2.")
	}

	for _, nr := range nodeRunners {
		nr.Stop()
	}
}

func TestNodeRunnerHasEvenProposers(t *testing.T) {
	defer sebaknetwork.CleanUpMemoryNetwork()

	numberOfNodes := 3
	nodeRunners := createTestNodeRunnerWithReady(numberOfNodes)

	nr0 := nodeRunners[0]
	nr1 := nodeRunners[1]
	nr2 := nodeRunners[2]

	var maximumBlockHeight uint64 = 3
	var maximumRoundNumber uint64 = 10

	proposers0 := make([]string, maximumBlockHeight*maximumRoundNumber)

	for i := uint64(0); i < maximumBlockHeight; i++ {
		for j := uint64(0); j < maximumRoundNumber; j++ {
			proposers0[i*maximumRoundNumber+j] = nr0.CalculateProposer(i, j)
		}
	}

	numN0inProposers := 0
	numN1inProposers := 0
	numN2inProposers := 0

	for _, p := range proposers0 {
		if p == nr0.localNode.Address() {
			numN0inProposers++
		} else if p == nr1.localNode.Address() {
			numN1inProposers++
		} else if p == nr2.localNode.Address() {
			numN2inProposers++
		}
	}

	passCriteria := int(maximumBlockHeight) * int(maximumRoundNumber) / numberOfNodes * 2

	if numN0inProposers >= passCriteria {
		t.Error("failed to have even number of proposers. numN0inProposers", numN0inProposers)
	}
	if numN1inProposers >= passCriteria {
		t.Error("failed to have even number of proposers. numN1inProposers", numN1inProposers)
	}
	if numN2inProposers >= passCriteria {
		t.Error("failed to have even number of proposers. numN2inProposers", numN2inProposers)
	}

	for _, nr := range nodeRunners {
		nr.Stop()
	}
}
