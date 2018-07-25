package sebak

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBlock(t *testing.T) {
	_, tx := TestMakeTransaction(networkID, 1)
	encoded, _ := tx.Serialize()
	bt := NewBlockTransactionFromTransaction(tx, encoded)

	block := NewBlock(0, []*BlockTransaction{&bt}, nil)

	assert.Zero(t, block.Header.Height)
	assert.Nil(t, block.Header.PrevConsensusResult) // nil in genesis block
	assert.NotNil(t, block.Header.Timestamp)
	assert.Empty(t, block.Header.PrevBlockHash)
}
