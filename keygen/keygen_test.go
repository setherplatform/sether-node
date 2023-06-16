package keygen

import (
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/setherplatform/sether-node/common"
	abacommon "github.com/setherplatform/sether-node/keygen/common/aba"
	"github.com/setherplatform/sether-node/keygen/message_handlers/acss"
	log "github.com/sirupsen/logrus"
	"math/big"
	"runtime"
	"testing"
	"time"
)

func TestKeygen(t *testing.T) {
	timeout := time.After(30 * time.Second)
	done := make(chan bool)
	log.SetLevel(log.DebugLevel)

	nodes, transport := setupNodes(5, 0)
	id := common.GenerateADKGID(*big.NewInt(int64(1)))
	for _, n := range nodes {
		go func(node *Node) {
			round := common.RoundDetails{
				ADKGID: id,
				Dealer: node.ID(),
				Kind:   "acss",
			}
			msg, err := acss.NewShareMessage(
				round.ID(),
				common.SECP256K1,
			)
			if err != nil {
				log.Error("EndBlock:Acss.NewShareMessage")
			}
			node.ReceiveMessage(node.Details(), *msg)
		}(n)
	}

	go func() {
		var outputCount = 0
		var outputs []string
		for {
			output := <-transport.output
			t.Logf("Output: %s", output)
			outputs = append(outputs, output)
			outputCount++
			var shares map[int]*big.Int = make(map[int]*big.Int)
			var identities []int
			if outputCount == n {
				for _, node := range nodes {
					if _, ok := node.shares[1]; ok {
						shares[node.id] = node.shares[1]
						identities = append(identities, node.id)
					}
				}

				coeff, _ := abacommon.LagrangeCoeffs(identities, curves.K256())

				z := curves.K256().NewScalar().Zero()
				for i := range coeff {
					si, _ := curves.K256().NewScalar().SetBigInt(shares[i])
					z = z.Add(si.Mul(coeff[i]))
				}

				publicKey := curves.K256().ScalarBaseMult(z).ToAffineUncompressed()
				t.Logf("derivedPublicKey: %x", publicKey[1:])
				t.Logf("actualPublicKey: %s", output)
				t.Logf("outputtedpublickeys: %s", outputs)
				done <- true
			}
		}
	}()

	select {
	case <-timeout:
		t.Fatal("Test didn't finish in time")
	case <-done:
	}
}

func TestKeygenWithNodesDown(t *testing.T) {
	timeout := time.After(8 * time.Second)
	done := make(chan bool)

	t.Deadline()
	log.SetLevel(log.DebugLevel)
	runtime.GOMAXPROCS(10)
	nodes, transport := setupNodes(5, 2)
	id := common.GenerateADKGID(*big.NewInt(int64(1)))
	for _, n := range nodes {
		go func(node *Node) {
			round := common.RoundDetails{
				ADKGID: id,
				Dealer: node.ID(),
				Kind:   "acss",
			}
			msg, err := acss.NewShareMessage(
				round.ID(),
				common.SECP256K1,
			)
			if err != nil {
				log.WithError(err).Error("EndBlock:Acss.NewShareMessage")
			}
			node.ReceiveMessage(node.Details(), *msg)
		}(n)
	}

	// var output string
	go func() {
		var outputCount = 0
		var res string
		for {
			output := <-transport.output
			t.Logf("Output: %s", output)
			if res == "" {
				res = output[2:]
			}

			if output[2:] == res {
				outputCount++
			}
			if outputCount >= k {
				done <- true
			}
		}
	}()

	select {
	case <-timeout:
		t.Fatal("Test didn't finish in time")
	case <-done:
	}

}
