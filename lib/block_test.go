package sebak

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockHeaderTransactionRoot(t *testing.T) {
	_, tx := TestMakeTransaction(networkID, 1)
	encoded, _ := tx.Serialize()
	bt := NewBlockTransactionFromTransaction(tx, encoded)

	block := NewBlock(0, []*BlockTransaction{&bt}, nil, 0)
	block2 := NewBlock(0, []*BlockTransaction{&bt}, nil, 0)

	assert.Equal(t, block.Header.TransactionsRoot, block2.Header.TransactionsRoot)

	bt.Hash = "CHANGED"
	blockWithChangedBlockTransaction := NewBlock(0, []*BlockTransaction{&bt}, nil, 0)

	assert.NotEqual(t, block.Header.TransactionsRoot, blockWithChangedBlockTransaction.Header.TransactionsRoot)

}

func TestNewBlock(t *testing.T) {
	_, tx := TestMakeTransaction(networkID, 1)
	encoded, _ := tx.Serialize()
	bt := NewBlockTransactionFromTransaction(tx, encoded)

	block := NewBlock(0, []*BlockTransaction{&bt}, nil, 0)

	assert.Zero(t, block.Header.Height)
	assert.Nil(t, block.Header.PrevConsensusResult) // nil in genesis block
	assert.NotNil(t, block.Header.Timestamp)
	assert.Empty(t, block.Header.PrevBlockHash)
}

// Make Multiple Txs like Tx Pool Stub

// Make Multiple Blocks

// Block Link

// Block Propose Message

// Selelct Proposer

// Block and Proposer into Isaac(Testcase)

// Make Tx Pool

// Share Txs

// Isaac in Memory Network
