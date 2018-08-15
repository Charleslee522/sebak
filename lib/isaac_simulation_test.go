package sebak

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"boscoin.io/sebak/lib/network"
)

type ProposerManager interface {
}

func TestIsaacSimulationProposer(t *testing.T) {
	assert.True(t, true)

	nodeRunners := CreateTestNodeRunner(5)

	_, tx := TestMakeTransaction(networkID, 1)
	var b []byte
	var err error

	if b, err = tx.Serialize(); err != nil {
		return
	}

	message := sebaknetwork.Message{Type: "message", Data: b}

	nodeRunner := nodeRunners[0]

	nodeRunner.HandleMessageFromClient(message)
	assert.Nil(t, err)
	assert.True(t, nodeRunner.Consensus().TransactionPool.Has(tx.GetHash()))

	assert.Equal(t, "", nodeRunner.CalculateProposer(0, 1))

}
