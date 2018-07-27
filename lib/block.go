package sebak

import "boscoin.io/sebak/lib/common"

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
