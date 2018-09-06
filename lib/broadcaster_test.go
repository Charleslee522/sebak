package sebak

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectionManagerBroadcaster(t *testing.T) {
	nodeRunners := createTestNodeRunner(3)

	nr := nodeRunners[0]

	recv := make(chan struct{})

	b := NewTestBroadcaster(recv)
	nr.SetBroadcaster(b)
	nr.SetProposerCalculator(SelfProposerCalculator{})

	nr.Consensus().SetLatestConsensusedBlock(genesisBlock)

	conf := NewISAACConfiguration()

	nr.SetConf(conf)

	nr.StartStateManager()

	<-recv
	require.Equal(t, 1, len(b.Messages))
}
