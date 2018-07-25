package sebak

type Block struct {
	Header       *Header
	Transactions []*BlockTransaction
}

func getTransactionRoot(txs []*BlockTransaction) string {
	return "ROOT"
}

func NewBlock(height uint64, txs []*BlockTransaction, prevResult *ConsensusResult, prevTotalTxs uint64) *Block {
	txRoot := getTransactionRoot(txs)
	p := Block{
		Header:       NewHeader(height, prevResult, prevTotalTxs, uint64(len(txs)), txRoot),
		Transactions: txs,
	}
	return &p
}
