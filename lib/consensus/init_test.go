package sebakconsensus

import (
	logging "github.com/inconshreveable/log15"

	"boscoin.io/sebak/lib"
	"boscoin.io/sebak/lib/common/test"
)

var networkID []byte = []byte("sebak-consensus-test-network")

func init() {
	sebak.SetLogging(logging.LvlDebug, test.LogHandler())
}
