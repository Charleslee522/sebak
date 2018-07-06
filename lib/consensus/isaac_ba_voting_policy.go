package sebakconsensus

import (
	"math"

	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/error"
)

type Percent struct {
	data int
}

func NewPercent(i int) (*Percent, error) {
	if i < 0 || i > 100 {
		return nil, sebakerror.ErrorBlockTransactionDoesNotExists
	}
	p := Percent{data: i}
	return &p, nil
}

func (p *Percent) String() string {
	return string(p.data)
}

func (p *Percent) Int() int {
	return p.data
}

type IsaacBaVotingThresholdPolicy struct {
	threshold Percent

	validators int
}

func (vt *IsaacBaVotingThresholdPolicy) String() string {
	o := sebakcommon.MustJSONMarshal(map[string]interface{}{
		"threshold":  vt.threshold,
		"validators": vt.validators,
	})

	return string(o)
}

func (vt *IsaacBaVotingThresholdPolicy) Validators() int {
	return vt.validators
}

func (vt *IsaacBaVotingThresholdPolicy) SetValidators(n int) error {
	if n < 1 {
		return sebakerror.ErrorVotingThresholdInvalidValidators
	}

	vt.validators = n

	return nil
}

func (vt *IsaacBaVotingThresholdPolicy) Connected() int {
	return -1
}

func (vt *IsaacBaVotingThresholdPolicy) SetConnected(n int) error {
	return nil
}

func (vt *IsaacBaVotingThresholdPolicy) Threshold(state sebakcommon.BallotState) int {
	v := float64(vt.validators) * (float64(vt.threshold.Int()) / float64(100))
	return int(math.Ceil(v))
}

func (vt *IsaacBaVotingThresholdPolicy) Reset(state sebakcommon.BallotState, threshold int) (err error) {
	return nil
}

func NewIsaacBaVotingThresholdPolicy(threshold int) (*IsaacBaVotingThresholdPolicy, error) {
	if th, err := NewPercent(threshold); err != nil {
		return nil, err
	} else {
		vt := &IsaacBaVotingThresholdPolicy{
			threshold:  *th,
			validators: 0,
		}
		return vt, err
	}
}
