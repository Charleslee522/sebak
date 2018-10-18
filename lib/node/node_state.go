package node

import (
	"fmt"
)

type State uint

const (
	StateCONSENSUS State = iota
	StateSYNC
)

func (s State) String() string {
	switch s {
	case 0:
		return "CONSENSUS"
	case 1:
		return "SYNC"
	}

	return ""
}

func (s State) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", s.String())), nil
}

func (s *State) UnmarshalJSON(b []byte) (err error) {
	var c int
	switch string(b[1 : len(b)-1]) {
	case "CONSENSUS":
	case "SYNC":
		c = 1
	}

	*s = State(c)

	return
}
