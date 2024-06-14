package model

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)

// MethodChannel defines in/out channel messages.
type MethodChannel struct {
	In  *Type
	Out *Type
}

func newMethodChannel(p *syntax.MethodChannel) (*MethodChannel, error) {
	if p == nil {
		return nil, nil
	}

	if p.In == nil && p.Out == nil {
		return nil, fmt.Errorf("channel in or out must be specified")
	}

	var in *Type
	var err error
	if p.In != nil {
		in, err = newType(p.In)
		if err != nil {
			return nil, err
		}
	}

	var out *Type
	if p.Out != nil {
		out, err = newType(p.Out)
		if err != nil {
			return nil, err
		}
	}

	ch := &MethodChannel{
		In:  in,
		Out: out,
	}
	return ch, nil
}

// resolve

func (ch *MethodChannel) resolve(file *File) error {
	if ch.In != nil {
		if err := ch.In.resolve(file); err != nil {
			return err
		}
	}
	if ch.Out != nil {
		if err := ch.Out.resolve(file); err != nil {
			return err
		}
	}
	return nil
}
