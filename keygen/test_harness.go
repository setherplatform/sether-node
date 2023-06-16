package keygen

import (
	"crypto/rand"
	"fmt"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/setherplatform/sether-node/common"
	"github.com/setherplatform/sether-node/common/sharing"
	kcommon "github.com/setherplatform/sether-node/keygen/common"
	acssc "github.com/setherplatform/sether-node/keygen/common/acss"
	"github.com/setherplatform/sether-node/keygen/message_handlers/aba"
	"github.com/setherplatform/sether-node/keygen/message_handlers/acss"
	"github.com/setherplatform/sether-node/keygen/message_handlers/keyderivation"
	"github.com/setherplatform/sether-node/keygen/message_handlers/keyset"
	log "github.com/sirupsen/logrus"
	"github.com/torusresearch/bijson"
	"math/big"
	"strings"
)

var n = 7
var k = 3
var c = curves.K256()
var randomScalar = c.Scalar.Random(rand.Reader)

func setupNodes(count int, faultyCount int) ([]*Node, *MockTransport) {
	nodes := []*Node{}
	nodeList := make(map[int]common.KeyPair)
	for i := 1; i <= count+faultyCount; i++ {
		keypair := acssc.GenerateKeyPair(curves.K256())
		nodeList[i] = keypair
	}
	transport := NewMockTransport(nodes)

	log.Info("Creating nodes...")
	i := 1

	for j := 0; j < count; j++ {
		log.Info("Creating node", "id", i)
		node := NewNode(i, n, k, nodeList[i], transport, false)
		nodes = append(nodes, node)
		i++
	}
	for j := 0; j < faultyCount; j++ {
		log.Info("Creating faulty node", "id", i)
		node := NewNode(i, n, k, nodeList[i], transport, true)
		nodes = append(nodes, node)
		i++
	}

	transport.Init(nodes)
	return nodes, transport
}

type Node struct {
	id           int
	n            int
	k            int
	transport    *MockTransport
	state        *common.NodeState
	keypair      common.KeyPair
	isFaulty     bool
	messageCount int
	shares       map[int64]*big.Int
}

func NewNode(id, n, k int, keypair common.KeyPair, transport *MockTransport, isFaulty bool) *Node {
	node := Node{
		id: id,
		n:  n,
		k:  k,
		state: &common.NodeState{
			KeygenStore:  &common.SharingStoreMap{},
			SessionStore: &common.ADKGSessionStore{},
			ABAStore:     &common.ABAStoreMap{},
		},
		transport: transport,
		keypair:   keypair,
		isFaulty:  isFaulty,
		shares:    make(map[int64]*big.Int),
	}
	return &node
}

func (node *Node) ID() int {
	return node.id
}

func (node *Node) Params() (int, int, int) {
	return node.n, node.k, node.k - 1
}

func (node *Node) CurveParams(c string) (curves.Point, curves.Point) {
	return sharing.CurveParams(c)
}

func (node *Node) State() *common.NodeState {
	return node.state
}

func (node *Node) Cleanup(id common.ADKGID) {
	node.cleanupKeygenStore(id)
	node.cleanupABAStore(id)
	node.cleanupADKGSessionStore(id)
	// debug.FreeOSMemory()
}

func (node *Node) ReceiveMessage(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
	node.messageCount = node.messageCount + 1
	switch {
	case strings.HasPrefix(keygenMessage.Method, "acss"):
		node.ProcessACSSMessages(sender, keygenMessage)
	case strings.HasPrefix(keygenMessage.Method, "keyset"):
		node.ProcessKeysetMessages(sender, keygenMessage)
	case strings.HasPrefix(keygenMessage.Method, "aba"):
		node.ProcessABAMessages(sender, keygenMessage)
	case strings.HasPrefix(keygenMessage.Method, "key_derivation"):
		node.ProcessKeyDerivationMessages(sender, keygenMessage)

	default:
		log.Info(fmt.Sprintf("No handler found. MsgType=%s", keygenMessage.Method))
	}
}

func (node *Node) ProcessACSSMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
	switch keygenMessage.Method {
	case acss.ShareMessageType:
		log.Infof("Got %s", acss.ShareMessageType)
		var msg acss.ShareMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)

	case acss.ProposeMessageType:
		log.Info(fmt.Sprintf("Got %s", acss.ProposeMessageType))
		var msg acss.ProposeMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	case acss.EchoMessageType:
		log.Info(fmt.Sprintf("Got %s", acss.EchoMessageType))
		var msg acss.EchoMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	case acss.ReadyMessageType:
		log.Info(fmt.Sprintf("Got %s", acss.ReadyMessageType))
		var msg acss.ReadyMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	case acss.OutputMessageType:
		log.Info(fmt.Sprintf("Got %s", acss.OutputMessageType))
		var msg acss.OutputMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	}
}

func (node *Node) ProcessKeysetMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
	switch keygenMessage.Method {
	case keyset.InitMessageType:
		log.Info(fmt.Sprintf("Got %s", keyset.InitMessageType))
		var msg keyset.InitMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	case keyset.ProposeMessageType:
		log.Info(fmt.Sprintf("Got %s", keyset.ProposeMessageType))
		var msg keyset.ProposeMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	case keyset.EchoMessageType:
		log.Info(fmt.Sprintf("Got %s", keyset.EchoMessageType))
		var msg keyset.EchoMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	case keyset.ReadyMessageType:
		log.Info(fmt.Sprintf("Got %s", keyset.ReadyMessageType))
		var msg keyset.ReadyMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	case keyset.OutputMessageType:
		log.Info(fmt.Sprintf("Got %s", keyset.OutputMessageType))
		var msg keyset.OutputMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	}
}

func (node *Node) ProcessABAMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
	switch keygenMessage.Method {
	case aba.InitMessageType:
		log.Info(fmt.Sprintf("Got %s", aba.InitMessageType))
		var msg aba.InitMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Errorf("Could not unmarshal: MsgType=%s", keygenMessage.Method)
			return
		}
		msg.Process(sender, node)
	case aba.Est1MessageType:
		log.Info(fmt.Sprintf("Got %s", aba.Est1MessageType))
		var msg aba.Est1Message
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	case aba.Aux1MessageType:
		log.Info(fmt.Sprintf("Got %s", aba.Aux1MessageType))
		var msg aba.Aux1Message
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	case aba.AuxsetMessageType:
		log.Info(fmt.Sprintf("Got %s", aba.AuxsetMessageType))
		var msg aba.AuxsetMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	case aba.Est2MessageType:
		log.Info(fmt.Sprintf("Got %s", aba.Est2MessageType))
		var msg aba.Est2Message
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	case aba.Aux2MessageType:
		log.Info(fmt.Sprintf("Got %s", aba.Aux2MessageType))
		var msg aba.Aux2Message
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	case aba.CoinInitMessageType:
		log.Info(fmt.Sprintf("Got %s", aba.CoinInitMessageType))
		var msg aba.CoinInitMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	case aba.CoinMessageType:
		log.Info(fmt.Sprintf("Got %s", aba.CoinMessageType))
		var msg aba.CoinMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	}
}

func (node *Node) ProcessKeyDerivationMessages(sender common.KeygenNodeDetails, keygenMessage common.DKGMessage) {
	switch keygenMessage.Method {
	case keyderivation.InitMessageType:
		log.Info(fmt.Sprintf("Got %s", keyderivation.InitMessageType))
		var msg keyderivation.InitMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	case keyderivation.ShareMessageType:
		log.Info(fmt.Sprintf("Got %s", keyderivation.ShareMessageType))
		var msg keyderivation.ShareMessage
		err := bijson.Unmarshal(keygenMessage.Data, &msg)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Could not unmarshal: MsgType=%s", keygenMessage.Method))
			return
		}
		msg.Process(sender, node)
	}
}

func (node *Node) cleanupKeygenStore(id common.ADKGID) {
	for _, n := range node.Nodes() {
		node.state.KeygenStore.Complete((&common.RoundDetails{
			ADKGID: id,
			Dealer: n.Index,
			Kind:   "acss",
		}).ID())
		node.state.KeygenStore.Complete((&common.RoundDetails{
			ADKGID: id,
			Dealer: n.Index,
			Kind:   "keyset",
		}).ID())
	}
}

func (node *Node) cleanupABAStore(id common.ADKGID) {
	for _, n := range node.Nodes() {
		node.state.ABAStore.Complete((&common.RoundDetails{
			ADKGID: id,
			Dealer: n.Index,
			Kind:   "keyset",
		}).ID())
	}
}

func (node *Node) cleanupADKGSessionStore(id common.ADKGID) {
	node.state.SessionStore.Complete(id)
}

func (node *Node) StoreCompletedShare(index big.Int, si big.Int) {
	node.shares[index.Int64()] = &si
}

func (node *Node) StoreCommitment(index big.Int, metadata common.ADKGMetadata) {
	// n.shares[index.Int64()] = &si
}

func (node *Node) Broadcast(m common.DKGMessage) {
	if node.isFaulty {
		log.Info(fmt.Sprintf("Got Broadcast %s at faulty node %d", m.Method, node.id))
		return
	}
	node.transport.Broadcast(node.Details(), m)
}

func (node *Node) Send(receiver common.KeygenNodeDetails, msg common.DKGMessage) error {
	if node.isFaulty {
		log.Info(fmt.Sprintf("Got Send %s at faulty node %d", msg.Method, node.id))
		return nil
	}
	node.transport.Send(node.Details(), receiver, msg)
	return nil
}

func (node *Node) Nodes() map[common.NodeDetailsID]common.KeygenNodeDetails {
	return node.transport.nodeDetails
}

func (node *Node) Details() common.KeygenNodeDetails {
	return common.KeygenNodeDetails{
		Index:  node.id,
		PubKey: kcommon.CurvePointToPoint(node.keypair.PublicKey),
	}
}

func (n *Node) ReceiveBFTMessage(msg common.DKGMessage) {
	if msg.Method == keyderivation.PubKeygenType {
		var m keyderivation.PubKeygenMessage
		if err := bijson.Unmarshal(msg.Data, &m); err != nil {
			log.Info("ReceiveBFTMessage()")
			return
		}
		adkgid, _ := common.ADKGIDFromRoundID(m.RoundID)
		log.Info(fmt.Sprintf("ADKGID=%s", adkgid))
		res := m.PublicKey.X.Text(16) + m.PublicKey.Y.Text(16)
		go func() { n.transport.output <- res }()
	}
}

func (node *Node) PrivateKey() curves.Scalar {
	return node.keypair.PrivateKey
}

func (node *Node) PublicKey(index int) curves.Point {
	for _, n := range node.transport.nodes {
		if n.ID() == index {
			return n.keypair.PublicKey
		}
	}
	c := curves.K256()
	return c.Point.Identity()
}

type MockTransport struct {
	nodes       []*Node
	nodeDetails map[common.NodeDetailsID]common.KeygenNodeDetails
	output      chan string
}

func NewMockTransport(nodes []*Node) *MockTransport {
	return &MockTransport{output: make(chan string, 100)}
}

func (t *MockTransport) Init(nodes []*Node) {
	t.nodes = nodes
	nodeDetails := make(map[common.NodeDetailsID]common.KeygenNodeDetails)

	for _, node := range nodes {
		d := node.Details()
		nodeDetails[(&d).ToNodeDetailsID()] = node.Details()
	}
	t.nodeDetails = nodeDetails
}

// Sends message to everyone on transport
func (t *MockTransport) Broadcast(sender common.KeygenNodeDetails, m common.DKGMessage) {
	for _, p := range t.nodes {
		go func(node common.DkgParticipant) {
			node.ReceiveMessage(sender, m)
		}(p)
	}
}

// Sends message to the participant
func (t *MockTransport) Send(sender, receiver common.KeygenNodeDetails, msg common.DKGMessage) {
	// time.Sleep(500 * time.Millisecond)
	for _, n := range t.nodes {
		log.Infof("msg=%s, sender=%d, receiver=%d, round=%s", msg.Method, n.ID(), receiver.Index, msg.RoundID)
		if n.ID() == receiver.Index {
			go n.ReceiveMessage(sender, msg)
			break
		}
	}
}
