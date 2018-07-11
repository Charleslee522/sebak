package sebakconsensus

import (
	"boscoin.io/sebak/lib/node"
)

func GetProposer(isaacBa *IsaacBA) *sebaknode.Validator {
	validators := isaacBa.Node.GetValidators()

	return
}
