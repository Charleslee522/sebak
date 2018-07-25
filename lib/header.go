package sebak

import "time"

type Header struct {
	Version             uint32
	PrevBlockHash       string // [TODO] change to Hash or Uint256 type
	TransactionsRoot    string // Merkle root of Txs
	Timestamp           time.Time
	Height              uint64
	TotalTxs            uint64
	PrevConsensusResult *ConsensusResult

	prevTotalTxs uint64
	blockHash    string // or types.Uint256
	// PrevConsensusResultHash types.Hash (TODO)
	// ConsensusPayloadHash    types.Hash (TODO)
	// ConsensusPayload        Payload  // or []byte
	// StateRoot types.Hash    // MPT of state
	// + smart contract fields (TODO)
}

func NewHeader(height uint64, prevResult *ConsensusResult, prevTotalTxs uint64, currentTxs uint64, txRoot string) *Header {
	p := Header{
		Timestamp:           time.Now(),
		Height:              height,
		PrevConsensusResult: prevResult,
		TotalTxs:            prevTotalTxs + currentTxs,
		TransactionsRoot:    txRoot,
	}
	p.fill()

	return &p
}

func (h *Header) fill() {
	if h.Version == 0 {
		// [TODO] fill Version
	}

	if h.PrevBlockHash == "" {
		if h.Height != 0 &&
			h.PrevConsensusResult != nil {
			h.PrevBlockHash = h.PrevConsensusResult.BlockHash
		}
	}

	if h.TransactionsRoot == "" {
		// [TODO] fill TransactionsRoot
	}

}

type ConsensusResult struct {
	BlockHash string // [TODO] change to Hash or Uint256 type
	Ballots   []*Ballot
}
