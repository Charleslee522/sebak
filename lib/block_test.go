package sebak

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockHeaderTransactionRoot(t *testing.T) {
	_, tx := TestMakeTransaction(networkID, 1)
	encoded, _ := tx.Serialize()
	bt := NewBlockTransactionFromTransaction(tx, encoded)

	block := NewBlock(0, []BlockTransaction{bt}, "", nil, 0)
	block2 := NewBlock(0, []BlockTransaction{bt}, "", nil, 0)

	assert.Equal(t, block.Header.TransactionsRoot, block2.Header.TransactionsRoot)

	bt.Hash = "CHANGED"
	blockWithChangedBlockTransaction := NewBlock(0, []BlockTransaction{bt}, "", nil, 0)

	assert.NotEqual(t, block.Header.TransactionsRoot, blockWithChangedBlockTransaction.Header.TransactionsRoot)

}

func TestNewBlock(t *testing.T) {
	bts := TestMakeBlockTransactions(networkID, 1)

	block := NewBlock(0, bts, "", nil, 0)

	assert.Zero(t, block.Header.Height)
	assert.Nil(t, block.Header.PrevConsensusResult) // nil in genesis block
	assert.NotNil(t, block.Header.Timestamp)
	assert.Empty(t, block.Header.PrevBlockHash)
}

func TestMultipleBlocks(t *testing.T) {
	btsBlock0 := TestMakeBlockTransactions(networkID, 10)
	block0 := NewBlock(0, btsBlock0, "", nil, 0)

	assert.Zero(t, block0.Header.Height)
	assert.Nil(t, block0.Header.PrevConsensusResult) // nil in genesis block
	assert.NotNil(t, block0.Header.Timestamp)
	assert.Empty(t, block0.Header.PrevBlockHash)

	btsBlock1 := TestMakeBlockTransactions(networkID, 10)
	block1 := NewBlock(1, btsBlock1, block0.Header.BlockHash, nil, block0.Header.TotalTxs)

	assert.Equal(t, block1.Header.Height, uint64(0x1))
	assert.NotNil(t, block1.Header.Timestamp)
	assert.Equal(t, block0.Header.BlockHash, block1.Header.PrevBlockHash)

}

// Block Propose Message

// Selelct Proposer

// Block and Proposer into Isaac(Testcase)

// Make Tx Pool

// Share Txs

// Isaac in Memory Network
