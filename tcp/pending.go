package tcp

import "github.com/basecomplextech/baselibrary/bin"

// pending is a set of pending stream IDs, used to avoid duplicates.
// implemented as a list and a map, because iteration over map[bin.Bin128] cases allocations.
type pending struct {
	list []bin.Bin128
	keys map[bin.Bin128]struct{}
}

func newPending() *pending {
	return &pending{
		keys: make(map[bin.Bin128]struct{}),
	}
}

func (p *pending) add(id bin.Bin128) bool {
	if _, ok := p.keys[id]; ok {
		return false
	}

	p.keys[id] = struct{}{}
	p.list = append(p.list, id)
	return true
}

func (p *pending) clear() {
	p.list = p.list[:0]

	for k := range p.keys {
		delete(p.keys, k)
	}
}

func (p *pending) len() int {
	return len(p.list)
}
