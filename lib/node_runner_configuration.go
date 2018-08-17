//
// Struct that bridges together components of a node
//
// NodeRunner bridges together the connection, storage and `LocalNode`.
// In this regard, it can be seen as a single node, and is used as such
// in unit tests.
//
package sebak

import (
	"time"
)

type NodeRunnerConfiguration struct {
	TimeoutINIT      time.Duration
	TimeoutSIGN      time.Duration
	TimeoutACCEPT    time.Duration
	TimeoutGatherTxs time.Duration

	TransactionsLimit int
}

func NewNodeRunnerConfiguration() *NodeRunnerConfiguration {
	p := NodeRunnerConfiguration{}
	p.SetINIT(2).SetSIGN(2).SetACCEPT(2).SetGatherTx(2).SetTxLimit(1000)
	return &p
}

func (n *NodeRunnerConfiguration) SetINIT(f float64) *NodeRunnerConfiguration {
	n.TimeoutINIT = Millisecond(f)
	return n
}

func (n *NodeRunnerConfiguration) SetSIGN(f float64) *NodeRunnerConfiguration {
	n.TimeoutSIGN = Millisecond(f)
	return n
}

func (n *NodeRunnerConfiguration) SetACCEPT(f float64) *NodeRunnerConfiguration {
	n.TimeoutACCEPT = Millisecond(f)
	return n
}

func (n *NodeRunnerConfiguration) SetGatherTx(f float64) *NodeRunnerConfiguration {
	n.TimeoutGatherTxs = Millisecond(f)
	return n
}

func (n *NodeRunnerConfiguration) SetTxLimit(i int) *NodeRunnerConfiguration {
	n.TransactionsLimit = i
	return n
}

func Millisecond(f float64) time.Duration {
	return time.Millisecond * time.Duration(int(f*1000))
}
