package runner

import (
	"sync"
	"time"

	"boscoin.io/sebak/lib/ballot"
	"boscoin.io/sebak/lib/block"
	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/voting"
)

// ISAACStateManager manages the ISAACState.
// The most important function `Start()` is called in StartStateManager() function in node_runner.go by goroutine.
type ISAACStateManager struct {
	sync.RWMutex

	nr              *NodeRunner
	state           ballot.State
	stateTransit    chan ballot.State
	stop            chan struct{}
	blockTimeBuffer time.Duration      // the time to wait to adjust the block creation time.
	transitSignal   func(ballot.State) // the function is called when the ISAACState is changed.
	genesis         time.Time          // the time at which the GenesisBlock was saved. It is used for calculating `blockTimeBuffer`.

	Conf common.Config
}

func NewISAACStateManager(nr *NodeRunner, conf common.Config) *ISAACStateManager {
	p := &ISAACStateManager{
		nr:              nr,
		state:           ballot.StateINIT,
		stateTransit:    make(chan ballot.State),
		stop:            make(chan struct{}),
		blockTimeBuffer: 2 * time.Second,
		transitSignal:   func(ballot.State) {},
		Conf:            conf,
	}

	genesisBlock := block.GetGenesis(nr.storage)
	p.genesis = genesisBlock.Header.Timestamp

	return p
}

func (sm *ISAACStateManager) SetBlockTimeBuffer() {
	sm.nr.Log().Debug("ISAACStateManager.SetBlockTimeBuffer()", "height", sm.nr.consensus.LatestBlock().Height)
	b := sm.nr.Consensus().LatestBlock()
	ballotProposedTime := getBallotProposedTime(b.Confirmed)
	sm.blockTimeBuffer = calculateBlockTimeBuffer(
		sm.Conf.BlockTime,
		calculateAverageBlockTime(sm.genesis, b.Height),
		time.Now().Sub(ballotProposedTime),
		1*time.Second,
	)
	sm.nr.Log().Debug(
		"calculated blockTimeBuffer",
		"blockTimeBuffer", sm.blockTimeBuffer,
		"blockTime", sm.Conf.BlockTime,
		"genesis", sm.genesis,
		"height", b.Height,
		"confirmed", b.Confirmed,
		"now", time.Now(),
	)

	return
}

func getBallotProposedTime(timeStr string) time.Time {
	ballotProposedTime, _ := common.ParseISO8601(timeStr)
	return ballotProposedTime
}

func calculateAverageBlockTime(genesis time.Time, blockHeight uint64) time.Duration {
	genesisBlockHeight := uint64(1)
	height := blockHeight - genesisBlockHeight
	sinceGenesis := time.Now().Sub(genesis)

	if height == 0 {
		return sinceGenesis
	} else {
		return sinceGenesis / time.Duration(height)
	}
}

func calculateBlockTimeBuffer(goal, average, untilNow, delta time.Duration) time.Duration {
	var blockTimeBuffer time.Duration

	epsilon := 50 * time.Millisecond
	if average >= goal {
		if average-goal < epsilon {
			blockTimeBuffer = goal - untilNow
		} else {
			blockTimeBuffer = goal - delta - untilNow
		}
	} else {
		if goal-average < epsilon {
			blockTimeBuffer = goal - untilNow
		} else {
			blockTimeBuffer = goal + delta - untilNow
		}
	}
	if blockTimeBuffer < 0 {
		blockTimeBuffer = 0
	}
	return blockTimeBuffer
}

func (sm *ISAACStateManager) SetTransitSignal(f func(ballot.State)) {
	sm.Lock()
	defer sm.Unlock()
	sm.transitSignal = f
}

func (sm *ISAACStateManager) TransitISAACState(ballotState ballot.State) {
	go func() {
		sm.stateTransit <- ballotState
	}()
}

func (sm *ISAACStateManager) Confirm() {
	h := sm.nr.consensus.LatestBlock().Height
	r := sm.nr.consensus.LatestRound()
	sm.nr.Log().Debug("ISAACStateManager.Confirm()", "height", h, "round", r)
	sm.TransitISAACState(ballot.StateALLCONFIRM)
}

// In `Start()` method a node proposes ballot.
// Or it sets or resets timeout. If it is expired, it broadcasts B(`EXP`).
// And it manages the node round.
func (sm *ISAACStateManager) Start() {
	sm.nr.localNode.SetConsensus()
	sm.nr.Log().Debug("ISAACStateManager.Start()", "ISAACState", sm.state)
	go func() {
		timer := time.NewTimer(time.Duration(1 * time.Hour))
		for {
			select {
			case <-timer.C:
				sm.nr.Log().Debug("timeout", "ISAACState", sm.state)
				if sm.state == ballot.StateACCEPT {
					sm.SetBlockTimeBuffer()
					sm.nr.consensus.SetLatestRound(sm.nr.consensus.LatestRound())
					sm.TransitISAACState(ballot.StateINIT)
					break
				}
				go sm.broadcastExpiredBallot(sm.state)
				sm.setState(sm.state.Next())
				sm.resetTimer(timer, sm.state)
				sm.transitSignal(sm.state)

			case state := <-sm.stateTransit:
				switch state {
				case ballot.StateINIT:
					sm.proposeOrWait(timer, state)
				case ballot.StateSIGN:
					sm.setState(state)
					sm.transitSignal(state)
					timer.Reset(sm.Conf.TimeoutSIGN)
				case ballot.StateACCEPT:
					sm.setState(state)
					sm.transitSignal(state)
					timer.Reset(sm.Conf.TimeoutACCEPT)
				case ballot.StateALLCONFIRM:
					sm.setState(state)
					sm.transitSignal(state)
					sm.SetBlockTimeBuffer()
					sm.TransitISAACState(ballot.StateINIT)
				}

			case <-sm.stop:
				return
			}
		}
	}()
}

func (sm *ISAACStateManager) broadcastExpiredBallot(state ballot.State) {
	sm.nr.Log().Debug("broadcastExpiredBallot", "ISAACState", state)
	b := sm.nr.consensus.LatestBlock()
	r := sm.nr.consensus.LatestRound()
	basis := voting.Basis{
		Round:     r + 1,
		Height:    b.Height,
		BlockHash: b.Hash,
		TotalTxs:  b.TotalTxs,
		TotalOps:  b.TotalOps,
	}

	proposerAddr := sm.nr.consensus.SelectProposer(b.Height, r)

	newExpiredBallot := ballot.NewBallot(sm.nr.localNode.Address(), proposerAddr, basis, []string{})
	newExpiredBallot.SetVote(state.Next(), voting.EXP)

	opc, _ := ballot.NewCollectTxFeeFromBallot(*newExpiredBallot, sm.nr.CommonAccountAddress)
	opi, _ := ballot.NewInflationFromBallot(*newExpiredBallot, sm.nr.CommonAccountAddress, sm.nr.InitialBalance)
	ptx, _ := ballot.NewProposerTransactionFromBallot(*newExpiredBallot, opc, opi)

	newExpiredBallot.SetProposerTransaction(ptx)
	newExpiredBallot.SignByProposer(sm.nr.localNode.Keypair(), sm.nr.networkID)
	newExpiredBallot.Sign(sm.nr.localNode.Keypair(), sm.nr.networkID)

	sm.nr.Log().Debug("broadcast", "ballot", *newExpiredBallot)
	sm.nr.ConnectionManager().Broadcast(*newExpiredBallot)
}

func (sm *ISAACStateManager) resetTimer(timer *time.Timer, state ballot.State) {
	switch state {
	case ballot.StateINIT:
		timer.Reset(sm.Conf.TimeoutINIT)
	case ballot.StateSIGN:
		timer.Reset(sm.Conf.TimeoutSIGN)
	case ballot.StateACCEPT:
		timer.Reset(sm.Conf.TimeoutACCEPT)
	}
}

// In proposeOrWait,
// if nr.localNode is proposer, it proposes new ballot,
// but if not, it waits for receiving ballot from the other proposer.
func (sm *ISAACStateManager) proposeOrWait(timer *time.Timer, state ballot.State) {
	timer.Reset(time.Duration(1 * time.Hour))
	h := sm.nr.consensus.LatestBlock().Height
	r := sm.nr.consensus.LatestRound()
	proposer := sm.nr.Consensus().SelectProposer(h, r)
	log.Debug("selected proposer", "proposer", proposer)

	if proposer == sm.nr.localNode.Address() {
		time.Sleep(sm.blockTimeBuffer)
		if _, err := sm.nr.proposeNewBallot(); err == nil {
			log.Debug("propose new ballot", "proposer", proposer, "round", r, "ballotState", ballot.StateSIGN)
		} else {
			log.Error("failed to proposeNewBallot", "height", h, "error", err)
		}
		timer.Reset(sm.Conf.TimeoutINIT)
	} else {
		timer.Reset(sm.blockTimeBuffer + sm.Conf.TimeoutINIT)
	}
	sm.setState(state)
	sm.transitSignal(state)
}

func (sm *ISAACStateManager) setState(state ballot.State) {
	sm.Lock()
	defer sm.Unlock()

	sm.nr.Log().Debug("ISAACStateManager.setState()", "state", state)

	sm.state = state

	return
}

func (sm *ISAACStateManager) State() ballot.State {
	sm.RLock()
	defer sm.RUnlock()

	sm.nr.Log().Debug("ISAACStateManager.State()", "state", sm.state)

	return sm.state
}

func (sm *ISAACStateManager) Stop() {
	go func() {
		sm.stop <- struct{}{}
	}()
}
