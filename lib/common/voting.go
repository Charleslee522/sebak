package sebakcommon

type VotingHole string

const (
	VotingNOTYET  VotingHole = "NOT-YET"
	VotingYES     VotingHole = "YES"
	VotingNO      VotingHole = "NO"
	VotingEXPIRED VotingHole = "EXPIRED"
)

type VotingThresholdPolicy interface {
	Threshold(BallotState, VotingHole) int
	Validators() int
	SetValidators(int) error
	Connected() int
	SetConnected(int) error

	String() string
}
