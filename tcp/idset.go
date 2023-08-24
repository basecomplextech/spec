package tcp

import "github.com/basecomplextech/baselibrary/bin"

// idset is a combination of list/map instead of a simple map,
// because iteration over map[bin.Bin128] cases allocations.
type idset struct {
	list []bin.Bin128
	keys map[bin.Bin128]struct{}
}

func newIDSet() *idset {
	return &idset{
		keys: make(map[bin.Bin128]struct{}),
	}
}

func (s *idset) add(id bin.Bin128) bool {
	if _, ok := s.keys[id]; ok {
		return false
	}

	s.keys[id] = struct{}{}
	s.list = append(s.list, id)
	return true
}

func (s *idset) clear() {
	s.list = s.list[:0]

	for k := range s.keys {
		delete(s.keys, k)
	}
}

func (s *idset) len() int {
	return len(s.list)
}
