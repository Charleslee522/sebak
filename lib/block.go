package sebak

import (
	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/storage"
)

type Block struct {
	Header              *Header
	Transactions        []BlockTransaction
	PrevConsensusResult *ConsensusResult
}

func NewBlock(height uint64, txs []BlockTransaction, prevBlockHash string, prevResult *ConsensusResult, prevTotalTxs uint64) *Block {
	p := Block{
		Header:       NewHeader(height, prevBlockHash, prevResult, prevTotalTxs, uint64(len(txs)), getTransactionRoot(txs)),
		Transactions: txs,
	}
	return &p
}

func getTransactionRoot(txs []BlockTransaction) string {
	return sebakcommon.MustMakeObjectHashString(txs) // [TODO] make root
}

func (b *Block) IsWellFormed([]byte) (err error) { // [TODO] see transaction.go:96
	return
}

func (b *Block) GetType() string {
	return "block"
}

func (b *Block) GetHash() string {
	return b.Header.BlockHash
}

func (b *Block) Serialize() (encoded []byte, err error) {
	encoded, err = sebakcommon.EncodeJSONValue(b)
	return
}

func (b *Block) String() string {
	encoded, _ := sebakcommon.EncodeJSONValue(b)
	return string(encoded)
}

func (b *Block) Equal(m sebakcommon.Message) bool {
	return b.GetHash() == m.GetHash()
}

func (b *Block) Source() string {
	panic("The block doesn't have the source")
}

func (b *Block) Save(st *sebakstorage.LevelDBBackend) (err error) {
	// [TODO] see block_transaction.go:103
	return
}
