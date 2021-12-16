package protocol

import "fmt"

type dataEntry struct {
	start int
	end   int
}

type dataStack struct {
	stack []dataEntry
}

func (s *dataStack) len() int {
	return len(s.stack)
}

func (s *dataStack) reset() {
	s.stack = s.stack[:0]
}

func (s *dataStack) pop() (dataEntry, error) {
	ln := len(s.stack)
	if ln == 0 {
		return dataEntry{}, fmt.Errorf("pop: data stack is empty")
	}

	e := s.stack[ln-1]
	s.stack = s.stack[:ln-1]
	return e, nil
}

func (s *dataStack) push(start int, end int) {
	e := dataEntry{
		start: start,
		end:   end,
	}
	s.stack = append(s.stack, e)
}
