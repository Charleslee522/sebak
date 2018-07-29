package sebak

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/btcsuite/btcutil/base58"
	"github.com/stellar/go/keypair"
	"github.com/stretchr/testify/assert"

	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/error"
	"boscoin.io/sebak/lib/node"
)

func NewRandomNode() *sebaknode.LocalNode {
	kp, _ := keypair.Random()
	a, _ := sebaknode.NewLocalNode(kp.Address(), &sebakcommon.Endpoint{}, "")
	a.SetKeypair(kp)
	return a
}

type DummyMessage struct {
	T    string
	Hash string
	Data string
}

func NewDummyMessage(data string) DummyMessage {
	d := DummyMessage{T: "dummy-message", Data: data}
	d.UpdateHash()

	return d
}

func (m DummyMessage) IsWellFormed([]byte) error {
	return nil
}

func (m DummyMessage) GetType() string {
	return m.T
}

func (m DummyMessage) Equal(n sebakcommon.Message) bool {
	return m.Hash == n.GetHash()
}

func (m DummyMessage) GetHash() string {
	return m.Hash
}

func (m DummyMessage) Source() string {
	return m.Hash
}

func (m *DummyMessage) UpdateHash() {
	m.Hash = base58.Encode(sebakcommon.MustMakeObjectHash(m.Data))
}

func (m DummyMessage) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

func (m DummyMessage) String() string {
	s, _ := json.MarshalIndent(m, "  ", " ")
	return string(s)
}

func NewBlockBallot() (Ballot, *Block) {
	txs := TestMakeTransactions(networkID, 10)
	block := NewBlock(0, txs, "", ConsensusResult{}, 0)

	kp, _ := keypair.Random()
	ballot := makeBallot(kp, block, sebakcommon.BallotStateINIT)
	return ballot, block
}

func makeISAAC(minimumValidators int) *ISAAC {
	policy, _ := NewDefaultVotingThresholdPolicy(100, 30, 30)
	policy.SetValidators(minimumValidators)

	is, _ := NewISAAC(networkID, NewRandomNode(), policy)

	return is
}

func makeBallot(kp *keypair.Full, m sebakcommon.Message, state sebakcommon.BallotState) Ballot {
	ballot, _ := NewBallotFromMessage(kp.Address(), m)
	ballot.SetState(state)
	ballot.Vote(VotingYES)
	ballot.Sign(kp, networkID)

	return ballot
}

func TestNewISAAC(t *testing.T) {
	policy, _ := NewDefaultVotingThresholdPolicy(100, 30, 30)
	policy.SetValidators(1)

	is, err := NewISAAC(networkID, NewRandomNode(), policy)
	if err != nil {
		t.Errorf("`NewISAAC` must not be failed: %v", err)
		return
	}

	// check BallotBox is empty
	if !is.MsgPool.IsEmpty() {
		t.Error("`MessagePool` must be empty")
	}
}

func TestISAACNewIncomingMessage(t *testing.T) {
	is := makeISAAC(1)

	m := NewDummyMessage(sebakcommon.GenerateUUID())

	{
		if err := is.PutMessage(m); err != nil {
			t.Error(err)
			return
		}

		if !is.MsgPool.HasMessage(m) {
			t.Error("failed to add message into MessagePool")
			return
		}
	}

	// receive same message
	{
		if err := is.PutMessage(m); err != sebakerror.ErrorNewButKnownMessage {
			t.Error("incoming known message must occurr `ErrorNewButKnownMessage`")
			return
		}
	}

	// send another message
	{
		another := NewDummyMessage(sebakcommon.GenerateUUID())

		err := is.PutMessage(another)
		if err != nil {
			t.Errorf("failed to add another message: %v", err)
			return
		}

		if !is.MsgPool.HasMessage(m) {
			t.Error("failed to add message into MessagePool")
			return
		}
	}
}

func TestISAACReceiveBallotStateINIT(t *testing.T) {
	is := makeISAAC(1)
	ballot, block := NewBlockBallot()

	// new ballot from the other node
	if _, err := is.ReceiveBallot(ballot); err != nil {
		t.Error(err)
		return
	}

	if is.ThisRoundBlock == nil {
		t.Error("Block from the received ballot is nil")
		return
	}

	assert.Equal(t, block.BlockHash, is.ThisRoundBlock.BlockHash)
}

func TestISAACIsVoted(t *testing.T) {
	is := makeISAAC(1)
	m := NewDummyMessage(sebakcommon.GenerateUUID())

	is.ReceiveMessage(m)

	kp, _ := keypair.Random()

	ballot := makeBallot(kp, m, sebakcommon.BallotStateINIT)

	if is.Boxes.IsVoted(ballot) {
		t.Error("`IsVoted` must be `false` ")
		return
	}

	is.ReceiveBallot(ballot)
	if !is.Boxes.IsVoted(ballot) {
		t.Error("failed to vote")
		return
	}
}

func TestISAACReceiveBallotStateINITAndMoveNextState(t *testing.T) {
	is := makeISAAC(5)

	ballot, _ := NewBlockBallot()

	if _, err := is.ReceiveBallot(ballot); err != nil {
		t.Error(err)
		return
	}

	// [TODO] Check `is` is Sign state

}

func TestISAACReceiveBallotStateSIGNAndMoveNextState(t *testing.T) {
	is := makeISAAC(5)

	var numberOfBallots = 5

	m := NewDummyMessage(sebakcommon.GenerateUUID())

	// make ballots
	var err error
	var ballots []Ballot
	var vs VotingStateStaging

	for i := 0; i < int(numberOfBallots); i++ {
		kp, _ := keypair.Random()

		ballot := makeBallot(kp, m, sebakcommon.BallotStateSIGN)
		ballots = append(ballots, ballot)

		if vs, err = is.ReceiveBallot(ballot); err != nil {
			t.Error(err)
			return
		}

		if !is.Boxes.IsVoted(ballot) {
			t.Error("failed to vote")
			return
		}
	}

	if vs.IsClosed() {
		t.Error("just state changed, not `VotingResult` closed")
		return
	}
	vr, err := is.Boxes.VotingResult(ballots[0])
	if err != nil {
		t.Error(err)
	}
	vs = vr.LatestStaging()
	if vs.IsEmpty() {
		t.Error("failed to get valid `VotingStateStaging`")
		return
	}
	if !vs.IsChanged() {
		t.Error("failed to change state")
		return
	}
}

func TestISAACReceiveBallotStateSIGNAndVotingBox(t *testing.T) {
	is := makeISAAC(5)

	var numberOfBallots = 5

	m := NewDummyMessage(sebakcommon.GenerateUUID())

	// make ballots
	var err error
	var vs VotingStateStaging
	var ballots []Ballot
	for i := 0; i < int(numberOfBallots); i++ {
		kp, _ := keypair.Random()

		ballot := makeBallot(kp, m, sebakcommon.BallotStateSIGN)
		ballots = append(ballots, ballot)

		if vs, err = is.ReceiveBallot(ballot); err != nil {
			t.Error(err)
			return
		}
		if !is.Boxes.IsVoted(ballot) {
			t.Error("failed to vote")
			return
		}
	}

	vr, err := is.Boxes.VotingResult(ballots[0])
	if err != nil {
		t.Error(err)
	}
	vs = vr.LatestStaging()
	if !vs.IsChanged() {
		t.Error("failed to get result")
		return
	}

	// if is.Boxes.WaitingBox.HasMessageByHash(ballots[0].MessageHash()) {
	// 	t.Error("after `INIT`, the ballot must move to `VotingBox`")
	// }
}

func voteISAACReceiveBallot(is *ISAAC, ballots []Ballot, kps []*keypair.Full, state sebakcommon.BallotState) (vs VotingStateStaging, err error) {
	for i, ballot := range ballots {
		ballot.SetState(state)
		ballot.Sign(kps[i], networkID)

		if vs, err = is.ReceiveBallot(ballot); err != nil {
			break
		}
		if !is.Boxes.IsVoted(ballot) {
			return
		}
	}
	if err != nil {
		return
	}

	vr, err := is.Boxes.VotingResult(ballots[0])
	if err != nil {
		return
	}
	vs = vr.LatestStaging()

	return
}

func TestISAACReceiveBallotStateTransition(t *testing.T) {
	var numberOfBallots = 5
	var minimumValidators = 3 // must be passed

	is := makeISAAC(minimumValidators)

	m := NewDummyMessage(sebakcommon.GenerateUUID())

	var ballots []Ballot
	var kps []*keypair.Full

	for i := 0; i < int(numberOfBallots); i++ {
		kp, _ := keypair.Random()
		kps = append(kps, kp)

		ballots = append(ballots, makeBallot(kp, m, sebakcommon.BallotStateINIT))
	}

	// INIT -> SIGN
	{
		vs, err := voteISAACReceiveBallot(is, ballots, kps, sebakcommon.BallotStateINIT)
		if err != nil {
			t.Error(err)
			return
		}

		// if is.Boxes.WaitingBox.HasMessageByHash(ballots[0].MessageHash()) {
		// 	t.Error("after `INIT`, the ballot must move to `VotingBox`")
		// }

		if vs.IsEmpty() {
			err = errors.New("failed to get result")
			return
		}
		if vs.State != sebakcommon.BallotStateSIGN {
			err = errors.New("`VotingResult.State` must be `BallotStateSIGN`")
			return
		}

		// if !is.Boxes.VotingBox.HasMessageByHash(ballots[0].MessageHash()) {
		// 	err = errors.New("after `INIT`, the ballot must move to `VotingBox`")
		// 	return
		// }
	}

	// SIGN -> ACCEPT
	{
		vs, err := voteISAACReceiveBallot(is, ballots, kps, sebakcommon.BallotStateSIGN)
		if err != nil {
			t.Error(err)
			return
		}

		if vs.IsEmpty() {
			err = errors.New("failed to get result")
			return
		}
		if vs.State != sebakcommon.BallotStateACCEPT {
			err = errors.New("`VotingResult.State` must be `BallotStateACCEPT`")
			return
		}

		// if !is.Boxes.VotingBox.HasMessageByHash(ballots[0].MessageHash()) {
		// 	err = errors.New("after `INIT`, the ballot must move to `VotingBox`")
		// 	return
		// }
	}

	// ACCEPT -> ALL-CONFIRM
	{
		vs, err := voteISAACReceiveBallot(is, ballots, kps, sebakcommon.BallotStateACCEPT)
		if err != nil {
			t.Error(err)
			return
		}
		if vs.IsEmpty() {
			err = errors.New("failed to get result")
			return
		}
		if vs.State != sebakcommon.BallotStateALLCONFIRM {
			err = errors.New("`VotingResult.State` must be `BallotStateALLCONFIRM`")
			return
		}

		// if !is.Boxes.VotingBox.HasMessageByHash(ballots[0].MessageHash()) {
		// 	err = errors.New("after `INIT`, the ballot must move to `VotingBox`")
		// 	return
		// }
	}
}

func TestISAACReceiveSameBallotStates(t *testing.T) {
	var numberOfBallots = 5
	var minimumValidators = 3

	is := makeISAAC(minimumValidators)

	ballot, _ := NewBlockBallot()

	_, err := is.ReceiveBallot(ballot)
	assert.Nil(t, err)

	m := NewDummyMessage(sebakcommon.GenerateUUID())

	var ballots []Ballot
	var kps []*keypair.Full

	for i := 0; i < int(numberOfBallots); i++ {
		kp, _ := keypair.Random()
		kps = append(kps, kp)

		ballots = append(ballots, makeBallot(kp, m, sebakcommon.BallotStateSIGN))
	}

	{
		vs, err := voteISAACReceiveBallot(is, ballots, kps, sebakcommon.BallotStateSIGN)
		if err != nil {
			t.Error(err)
			return
		}

		// if is.Boxes.WaitingBox.HasMessageByHash(ballots[0].MessageHash()) {
		// 	t.Error("after `INIT`, the ballot must move to `VotingBox`")
		// }
		if vs.IsEmpty() {
			t.Error("failed to get result")
			return
		}
		if vs.State != sebakcommon.BallotStateACCEPT {
			err = errors.New("`VotingResult.State` must be `BallotStateACCEPT`")
		}

		vr, err := is.Boxes.VotingResult(ballots[0])
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, int(numberOfBallots), vr.VotedCount(sebakcommon.BallotStateSIGN))

		if vr.VotedCount(sebakcommon.BallotStateACCEPT) != 0 || vr.VotedCount(sebakcommon.BallotStateALLCONFIRM) != 0 {
			t.Error("unexpected ballots found")
			return
		}
	}

	vrFirst, err := is.Boxes.VotingResult(ballots[0])
	if err != nil {
		t.Error(err)
	}
	{
		_, err := voteISAACReceiveBallot(is, ballots, kps, sebakcommon.BallotStateSIGN)
		if err != nil {
			t.Error(err)
			return
		}
	}
	vrSecond, err := is.Boxes.VotingResult(ballots[0])
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, int(numberOfBallots), vrSecond.VotedCount(sebakcommon.BallotStateSIGN))

	if vrSecond.VotedCount(sebakcommon.BallotStateACCEPT) != 0 || vrSecond.VotedCount(sebakcommon.BallotStateALLCONFIRM) != 0 {
		t.Error("unexpected ballots found")
		return
	}

	for k, v := range vrFirst.Ballots {
		for k0, v0 := range v {
			if v0.Hash != vrSecond.Ballots[k][k0].Hash {
				t.Error("not matched")
				break
			}
		}
	}
}
