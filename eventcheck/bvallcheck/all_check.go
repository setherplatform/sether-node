package bvallcheck

import (
	"github.com/setherplatform/sether-node/inter"
)

type Checker struct {
	HeavyCheck HeavyCheck
	LightCheck LightCheck
}

type LightCheck func(bvs inter.LlrSignedBlockVotes) error

type HeavyCheck interface {
	Enqueue(bvs inter.LlrSignedBlockVotes, checked func(error)) error
}

type Callback struct {
	HeavyCheck HeavyCheck
	LightCheck LightCheck
}

// Enqueue tries to fill gaps the fetcher's future import queue.
func (c *Checker) Enqueue(bvs inter.LlrSignedBlockVotes, checked func(error)) {
	// Run light checks right away
	err := c.LightCheck(bvs)
	if err != nil {
		checked(err)
		return
	}

	// Run heavy check in parallel
	_ = c.HeavyCheck.Enqueue(bvs, checked)
}
