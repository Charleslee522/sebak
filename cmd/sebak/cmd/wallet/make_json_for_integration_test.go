package wallet

import (
	"fmt"
	"testing"

	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/common/keypair"
	"boscoin.io/sebak/lib/transaction"
	"boscoin.io/sebak/lib/transaction/operation"
	"github.com/stretchr/testify/require"
)

const (
	networkID = "sebak-test-network"
	GSeed     = "SBECGI3FSCYHNQIMANNCWQSVA6S5C6L4BXFKAPMBAMI5V47NWXNE37MN"
	CSeed     = "SDZICYGKDNBTNO2YPCJPRIY2NP22A5AJBQ677NCHEOSHRT3EE4MSKVLZ"
	Seed1     = "SCYLEJMJA25DIHOLIJODNKHXW7QHQTA2A6SJVGFTBVBQM7JR5ZUU4POL"
	Seed2     = "SBO6DC5KIZNU7VM3WEG56NCYM3XRTSOB6DWVHVW232SBJ2B5NJDMRZDH"
	Seed3     = "SCOAEMGLQC7JICFXKQXSEKUEDCI7XNXLSICHH4VQTTA4HWJFWDS2ES6G"
	Seed4     = "SBHF5UQRODXFLTUTHF3IFRHBXTWTTJXRT2QIFSUISANRAJZ4YB7WKHBD"
	Seed5     = "SC4U37E2QPDD6B256CIAMH7KSU53WCMRFG24WPOFMFFV5DJZVFNVZUUD"
	Seed6     = "SDANXUH2T2BQC26B5HJVO32JGEZEWTUCLGA76O7IWXXGIHQR2DX3DA2T"
	Seed7     = "SB6EVL4A3GKH3OPQO3FLNXN3FXHGBA5XPSQBW5KYG7LZ34A7INPUY3MY"
	Seed8     = "SBGCF7GVF55RE2OWZUHGPQH4WT5RGDZYBSDJPEPLKVZT2GJP3W5XKQAL"
	Seed9     = "SBJ3PTHBNKTU4G3L5OVJA7MKCLYTWUJFFE3XKEKMOTJW6ESAZZJBBRME"
	Seed10    = "SDGPPOFBXCZUCXYLIAO4CKVLD4LHVR3KVASZBQDYKQFKJA4X452HP6XO"
)

func TestTransaction(t *testing.T) {
	sender, err := keypair.Parse(GSeed)
	require.NoError(t, err)

	commonKeypair, err := keypair.Parse(CSeed)
	require.NoError(t, err)

	str := "\n"
	str += fmt.Sprintln("export GENESIS_ACCOUNT=" + sender.Address())
	str += fmt.Sprintln("export GENESIS_SEED=" + GSeed)

	str += "\n"
	str += fmt.Sprintln("export COMMON_ACCOUNT=" + commonKeypair.Address())
	str += fmt.Sprintln("export COMMON_SEED=" + CSeed)

	seeds := []string{}
	seeds = append(seeds, Seed1)
	seeds = append(seeds, Seed2)
	seeds = append(seeds, Seed3)
	seeds = append(seeds, Seed4)
	seeds = append(seeds, Seed5)
	seeds = append(seeds, Seed6)
	seeds = append(seeds, Seed7)
	seeds = append(seeds, Seed8)
	seeds = append(seeds, Seed9)
	seeds = append(seeds, Seed10)
	keys := []keypair.KP{}
	i := 0
	str += "\n"
	for _, seed := range seeds {
		i++
		kp, err := keypair.Parse(seed)
		require.NoError(t, err)
		str += fmt.Sprintf("export PUB%d=%s\n", i, kp.Address())
		str += fmt.Sprintf("export SEED%d=%s\n", i, seed)
		keys = append(keys, kp)
	}

	t.Log(str)

	ops := makeCreateAccountOperations(keys, 100000000000)
	t.Log(makeTransactionFromOps(sender, 0, ops))
	t.Log(makePaymentTransaction(keys[0], 0, 10000000000, keys[1].Address(), keys[2].Address()))
	t.Log(makePaymentTransaction(keys[1], 0, 10000000000, keys[0].Address(), keys[2].Address()))
	t.Log(makePaymentTransaction(keys[2], 0, 10000000000, keys[3].Address(), keys[4].Address(), keys[5].Address()))
	t.Log(makePaymentTransaction(keys[3], 0, 10000000000, keys[2].Address(), keys[0].Address()))
	t.Log(makePaymentTransaction(keys[4], 0, 10000000000, keys[5].Address(), keys[6].Address(), keys[7].Address()))
	t.Log(makePaymentTransaction(keys[5], 0, 10000000000, keys[4].Address(), keys[1].Address()))
	t.Log(makePaymentTransaction(keys[6], 0, 10000000000, keys[7].Address(), keys[9].Address()))
	t.Log(makePaymentTransaction(keys[7], 0, 10000000000, keys[6].Address(), keys[4].Address()))
	t.Log(makePaymentTransaction(keys[8], 0, 10000000000, keys[9].Address()))
	t.Log(makePaymentTransaction(keys[9], 0, 10000000000, keys[8].Address(), keys[3].Address()))

	t.Log(makePaymentTransaction(keys[0], 1, 10000000000, keys[1].Address(), keys[2].Address()))
	t.Log(makePaymentTransaction(keys[1], 1, 10000000000, keys[0].Address(), keys[2].Address()))
	t.Log(makePaymentTransaction(keys[2], 1, 10000000000, keys[3].Address(), keys[4].Address(), keys[5].Address()))
	t.Log(makePaymentTransaction(keys[3], 1, 10000000000, keys[2].Address(), keys[0].Address()))
	t.Log(makePaymentTransaction(keys[4], 1, 10000000000, keys[5].Address(), keys[6].Address(), keys[7].Address()))
	t.Log(makePaymentTransaction(keys[5], 1, 10000000000, keys[4].Address(), keys[1].Address()))
	t.Log(makePaymentTransaction(keys[6], 1, 10000000000, keys[7].Address(), keys[9].Address()))
	t.Log(makePaymentTransaction(keys[7], 1, 10000000000, keys[6].Address(), keys[4].Address()))
	t.Log(makePaymentTransaction(keys[8], 1, 10000000000, keys[9].Address()))
	t.Log(makePaymentTransaction(keys[9], 1, 10000000000, keys[8].Address(), keys[3].Address()))

	t.Log(makePaymentTransaction(keys[9], 1, 10000000000, keys[8].Address(), keys[3].Address()))

}

func makeCreateAccountOperations(keys []keypair.KP, amount common.Amount) []operation.Operation {
	operations := []operation.Operation{}

	for _, key := range keys {
		opb := operation.NewCreateAccount(key.Address(), amount, "")

		op := operation.Operation{
			H: operation.Header{
				Type: operation.TypeCreateAccount,
			},
			B: opb,
		}
		operations = append(operations, op)
	}
	return operations
}

func makePaymentTransaction(keypair keypair.KP, seqid uint64, amount common.Amount, targets ...string) transaction.Transaction {
	ops := []operation.Operation{}
	for _, target := range targets {
		opb := operation.NewPayment(target, amount)

		op := operation.Operation{
			H: operation.Header{
				Type: operation.TypePayment,
			},
			B: opb,
		}
		ops = append(ops, op)
	}

	return makeTransactionFromOps(keypair, seqid, ops)
}

func makeTransactionFromOps(keypair keypair.KP, seqid uint64, ops []operation.Operation) transaction.Transaction {
	tx, _ := transaction.NewTransaction(keypair.Address(), seqid, ops...)
	tx.Sign(keypair, []byte(networkID))

	return tx
}
