package acss

import (
	"encoding/json"
	"github.com/setherplatform/sether-node/common"
	"github.com/setherplatform/sether-node/keygen/common/acss"
	"github.com/setherplatform/sether-node/keygen/messages"
	"math/big"

	log "github.com/sirupsen/logrus"
	"github.com/vivint/infectious"
)

var ProposeMessageType string = "acss_propose"

type ProposeMessage struct {
	RoundID common.RoundID
	Kind    string
	Curve   common.CurveName
	Data    []byte
}

func NewAcssProposeMessage(id common.RoundID, d []byte, curve common.CurveName) (*common.DKGMessage, error) {
	m := ProposeMessage{
		id,
		ProposeMessageType,
		curve,
		d,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m ProposeMessage) Process(sender common.KeygenNodeDetails, self common.DkgParticipant) {
	log.Debugf("Received Propose: Round=%sReceived Propose", m.RoundID)

	curve := common.CurveFromName(m.Curve)
	leader, err := m.RoundID.Leader()
	if err != nil {
		log.Errorf("Cound not get leader from roundID, err=%s", err)
		return
	}
	senderID := *new(big.Int).SetInt64(int64(sender.Index))

	// If leader of the round is not sender skip
	if leader.Cmp(&senderID) != 0 {
		log.Errorf("leader is not sender in acss_propose: sender=%d, leader=%d",
			sender.Index, leader.Int64())
		return
	}

	// Generated shared symmetric key
	n, k, _ := self.Params()
	priv := self.PrivateKey()

	dealerKey, err := curve.Point.Set(&sender.PubKey.X, &sender.PubKey.Y)
	// dealerKey, err := curve.Point.FromAffineCompressed(m.DealerPublicKey)
	if err != nil {
		log.Errorf("could not deserialize dealer public key: %s", err)
		return
	}

	key := acss.SharedKey(priv, dealerKey)

	// Verify self share against commitments
	data := &messages.MessageData{}
	err = data.Deserialize(m.Data)
	if err != nil {
		log.Errorf("could not deserialize message data: %s", err)
		return
	}
	_, _, verified := acss.Predicate(key[:], data.ShareMap[uint32(self.ID())][:],
		data.Commitments[:], k, curve)

	// If verified, send echo to each node
	if verified {

		// Create RS encoding
		fec, err := infectious.NewFEC(k, n)
		if err != nil {
			log.Errorf("error during creation of fec, err=%s", err)
			return
		}

		hash := common.HashByte(m.Data)

		shares, err := acss.Encode(fec, m.Data)
		if err != nil {
			log.Errorf("error during fec encoding, err=%s", err)
			return
		}

		for _, n := range self.Nodes() {
			go func(node common.KeygenNodeDetails) {
				msg, err := NewAcssEchoMessage(m.RoundID, shares[node.Index-1], hash, m.Curve)
				if err != nil {
					log.WithError(err).Info()
				}
				err = self.Send(node, *msg)
				if err != nil {
					log.WithError(err).Info()
				}
			}(n)
		}
	} else {
		log.Errorf("acss predicate failed on %d for propose message by %d", self.ID(), sender.Index)
	}
}
