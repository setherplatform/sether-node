package gossip

import (
	"github.com/setherplatform/sether-node/params"
)

func (s *Store) AddUpgradeHeight(h params.UpgradeHeight) {
	orig := s.GetUpgradeHeights()
	// allocate new memory to avoid race condition in cache
	cp := make([]params.UpgradeHeight, 0, len(orig)+1)
	cp = append(append(cp, orig...), h)

	s.rlp.Set(s.table.UpgradeHeights, []byte{}, cp)
	s.cache.UpgradeHeights.Store(cp)
}

func (s *Store) GetUpgradeHeights() []params.UpgradeHeight {
	if v := s.cache.UpgradeHeights.Load(); v != nil {
		return v.([]params.UpgradeHeight)
	}
	hh, ok := s.rlp.Get(s.table.UpgradeHeights, []byte{}, &[]params.UpgradeHeight{}).(*[]params.UpgradeHeight)
	if !ok {
		return []params.UpgradeHeight{}
	}
	s.cache.UpgradeHeights.Store(*hh)
	return *hh
}
