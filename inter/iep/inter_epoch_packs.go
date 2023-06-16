package iep

import (
	"github.com/setherplatform/sether-node/inter"
	"github.com/setherplatform/sether-node/inter/ier"
)

type LlrEpochPack struct {
	Votes  []inter.LlrSignedEpochVote
	Record ier.LlrIdxFullEpochRecord
}
