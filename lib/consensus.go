package sebak

import (
	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/node"
)

type Consensus interface {
	GetNode() *sebaknode.LocalNode
	HasMessage(sebakcommon.Message) bool
	HasMessageByHash(string) bool

	PutMessage(sebakcommon.Message) error
	GetBallot() (Ballot, error)

	ReceiveMessage(sebakcommon.Message) (Ballot, error)
	ReceiveBallot(Ballot) (VotingStateStaging, error)
	ReceiveBlock(Block) (VotingStateStaging, error)

	AddBallot(Ballot) error
	CloseConsensus(Ballot) error
}
