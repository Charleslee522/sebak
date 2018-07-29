package sebak

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockSerializeDeserialize(t *testing.T) {
	_, tx := TestMakeTransaction(networkID, 1)

	block := NewBlock(0, []Transaction{tx}, "", ConsensusResult{}, 0)
	bEncoded, _ := (*block).Serialize()
	blockDeserialized, _ := NewBlockFromJSON(bEncoded)

	assert.Equal(t, block.GetHash(), blockDeserialized.GetHash())
	assert.Equal(t, block.GetType(), blockDeserialized.GetType())
	assert.Equal(t, block.Header.Version, blockDeserialized.Header.Version)
	assert.Equal(t, block.Header.PrevBlockHash, blockDeserialized.Header.PrevBlockHash)
	assert.Equal(t, block.Header.TransactionsRoot, blockDeserialized.Header.TransactionsRoot)
	assert.Equal(t, block.Header.Height, blockDeserialized.Header.Height)
	assert.Equal(t, block.Header.TotalTxs, blockDeserialized.Header.TotalTxs)

	// assert.Equal(t, block.Transactions, blockDeserialized.Transactions)  [TODO] passing this assert
}

func TestBlockHeaderTransactionRoot(t *testing.T) {
	_, tx := TestMakeTransaction(networkID, 1)

	block := NewBlock(0, []Transaction{tx}, "", ConsensusResult{}, 0)
	block2 := NewBlock(0, []Transaction{tx}, "", ConsensusResult{}, 0)

	assert.Equal(t, block.Header.TransactionsRoot, block2.Header.TransactionsRoot)

	tx.H.Hash = "CHANGED"
	blockWithChangedBlockTransaction := NewBlock(0, []Transaction{tx}, "", ConsensusResult{}, 0)

	assert.NotEqual(t, block.Header.TransactionsRoot, blockWithChangedBlockTransaction.Header.TransactionsRoot)

}

func TestNewBlock(t *testing.T) {
	txs := TestMakeTransactions(networkID, 1)

	block := NewBlock(0, txs, "", ConsensusResult{}, 0)

	assert.Zero(t, block.Header.Height)
	assert.NotNil(t, block.Header.Timestamp)
	assert.Empty(t, block.Header.PrevBlockHash)
}

func TestMultipleBlocks(t *testing.T) {
	txsBlock0 := TestMakeTransactions(networkID, 10)
	block0 := NewBlock(0, txsBlock0, "", ConsensusResult{}, 0)

	assert.Zero(t, block0.Header.Height)
	assert.NotNil(t, block0.Header.Timestamp)
	assert.Empty(t, block0.Header.PrevBlockHash)

	txsBlock1 := TestMakeTransactions(networkID, 10)
	block1 := NewBlock(1, txsBlock1, block0.Header.BlockHash, ConsensusResult{}, block0.Header.TotalTxs)

	assert.Equal(t, block1.Header.Height, uint64(0x1))
	assert.NotNil(t, block1.Header.Timestamp)
	assert.Equal(t, block0.Header.BlockHash, block1.Header.PrevBlockHash)
}

// Transaction
/*
1. [O]클라이언트가 서버에 메시지를 보냄
2.1 [O]서버는 받은 메시지를 메시지 풀에 저장
2.2 [O]서버는 메시지 풀로 block을 만듦
2.3 [O]서버는 block으로 ballot을 만듦
3. []`특정 시간(블록 생성 주기)`이 되면,
3.1 []Proposer는 거래를 묶어다가 블록에 넣어서 propose한다.
3.2 []Non-proposer는 이번 proposer가 누군지 확인한 후, 그에게 다음 블록을 받기를 기다린다.
*/

// Consensus entry point(pull. not push)

// Block Propose Message

// Selelct Proposer

// Block and Proposer into Isaac(Testcase)

// Make Tx Pool

// Share Txs

// Isaac in Memory Network
