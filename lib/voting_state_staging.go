package sebak

import (
	"encoding/json"

	"boscoin.io/sebak/lib/common"
)

// VotingStateStaging will keep the snapshot at changing state.
type VotingStateStaging struct {
	State         sebakcommon.BallotState
	PreviousState sebakcommon.BallotState

	ID          string     // ID is unique and sequential
	MessageHash string     // MessageHash is `Message.Hash`
	VotingHole  VotingHole // voting is closed and it's last `VotingHole`
	Reason      error      // if `VotingNO` is concluded, the reason

	Ballots map[ /* NodeKey */ string]VotingResultBallot
}

func (vs VotingStateStaging) String() string {
	encoded, _ := json.MarshalIndent(vs, "", "  ")
	return string(encoded)
}

func (vs VotingStateStaging) IsChanged() bool {
	return vs.State > vs.PreviousState
}

func (vs VotingStateStaging) IsEmpty() bool {
	return len(vs.Ballots) < 1
}

func (vs VotingStateStaging) IsClosed() bool {
	if vs.IsEmpty() {
		return false
	}
	if vs.VotingHole == VotingNO {
		return true
	}
	if vs.State == sebakcommon.BallotStateALLCONFIRM {
		return true
	}

	return false
}

func (vs VotingStateStaging) IsStorable() bool {
	if !vs.IsClosed() {
		return false
	}
	if vs.State != sebakcommon.BallotStateALLCONFIRM {
		return false
	}
	if vs.VotingHole == VotingNO {
		return false
	}

	return true
}
