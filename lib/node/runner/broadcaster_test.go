package runner

import (
	"testing"
	"time"

	"boscoin.io/sebak/lib/consensus"
	"github.com/stretchr/testify/require"
)

func TestConnectionManagerBroadcaster(t *testing.T) {
	nodeRunners := createTestNodeRunner(3)

	nr := nodeRunners[0]

	recv := make(chan struct{})

	b := NewTestBroadcaster(recv)
	nr.Consensus().SetBroadcaster(b)
	nr.Consensus().SetProposerCalculator(SelfProposerCalculator{
		nodeRunner: nr,
	})

	nr.Consensus().SetLatestConsensusedBlock(genesisBlock)

	conf := consensus.NewISAACConfiguration()
	conf.TimeoutALLCONFIRM = 1 * time.Millisecond

	nr.SetConf(conf)

	nr.StartStateManager()

	<-recv
	require.Equal(t, 1, len(b.Messages()))
}
