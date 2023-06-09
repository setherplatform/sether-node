package aba

import (
	"encoding/json"
	"github.com/setherplatform/sether-node/common"

	log "github.com/sirupsen/logrus"
)

var Est1MessageType string = "aba_est1"

type Est1Message struct {
	RoundID common.RoundID
	Kind    string
	Curve   common.CurveName
	V       int
	R       int
}

func NewEst1Message(id common.RoundID, v, r int, curve common.CurveName) (*common.DKGMessage, error) {
	m := Est1Message{
		id,
		Est1MessageType,
		curve,
		v,
		r,
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	msg := common.CreateMessage(m.RoundID, m.Kind, bytes)
	return &msg, nil
}

func (m Est1Message) Process(sender common.KeygenNodeDetails, self common.DkgParticipant) {
	v, r := m.V, m.R

	store, complete := self.State().ABAStore.GetOrSetIfNotComplete(m.RoundID, common.DefaultABAStore())
	if complete {
		log.Infof("Keygen already complete: %s", m.RoundID)
		return
	}

	store.Lock()
	defer store.Unlock()

	if store.Round() != r {
		return
	}
	// Check if already present
	if Contains(store.Values("est", r, v), sender.Index) {
		log.Debugf("Got redundant EST message from %d", sender.Index)
		return
	}
	//Otherwise, add sender
	store.SetValues("est", r, v, sender.Index)

	_, _, f := self.Params()
	estLength := len(store.Values("est", r, v))
	log.Debugf("EstCount: %d, required: %d, round: %v", estLength, f+1, m.RoundID)
	if estLength > f+1 && !store.Sent("est", r, v) {
		store.SetSent("est", r, v)
		msg, err := NewEst1Message(m.RoundID, v, r, m.Curve)
		if err != nil {
			return
		}
		go self.Broadcast(*msg)
	}

	if estLength == (2*f)+1 {
		store.SetBin("bin", r, v)
		w := store.Bin("bin", r)[0]
		msg, err := NewAux1Message(m.RoundID, w, r, m.Curve)
		if err != nil {
			return
		}
		go self.Broadcast(*msg)
	}
}

func Contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
