package model

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/ast"
)

type Method struct {
	Name string

	Sub  bool // Returns subservice
	Chan bool // Returns channel

	Input   MethodInput
	Output  MethodOutput
	Channel *MethodChannel
}

func newMethod(pm *ast.Method) (*Method, error) {
	m := &Method{
		Name: pm.Name,
	}

	// Input
	if pm.Input != nil {
		input, err := newMethodInput(pm.Input)
		if err != nil {
			return nil, fmt.Errorf("invalid input: %w", err)
		}
		m.Input = input
	}

	// Output
	if pm.Output != nil {
		output, err := newMethodOutput(pm.Output)
		if err != nil {
			return nil, fmt.Errorf("invalid output: %w", err)
		}
		m.Output = output
	}

	// Channel
	if pm.Channel != nil {
		channel, err := newMethodChannel(pm.Channel)
		if err != nil {
			return nil, fmt.Errorf("invalid channel: %w", err)
		}
		m.Channel = channel
	}
	return m, nil
}

func (m *Method) resolve(file *File) error {
	if m.Input != nil {
		if err := m.Input.resolve(file); err != nil {
			return fmt.Errorf("%v: %w", m.Name, err)
		}
	}
	if m.Output != nil {
		if err := m.Output.resolve(file); err != nil {
			return fmt.Errorf("%v: %w", m.Name, err)
		}
	}
	if m.Channel != nil {
		if err := m.Channel.resolve(file); err != nil {
			return fmt.Errorf("%v: %w", m.Name, err)
		}
	}
	return nil
}

func (m *Method) resolved() error {
	// Input
	if m.Input != nil {
		switch in := m.Input.(type) {
		case *Type:
			if in.Kind != KindMessage {
				return fmt.Errorf("%v: single input must be a message, got %q instead",
					m.Name, in.Kind)
			}

		case *Fields:
			if err := in.resolved(); err != nil {
				return err
			}
		}
	}

	// Output
	if m.Output != nil {
		switch out := m.Output.(type) {
		case *Type:
			switch out.Kind {
			case KindMessage:
			case KindService:
				m.Sub = true
			default:
				return fmt.Errorf("%v: single output must be a message or a service, got %q instead",
					m.Name, out.Kind)
			}

		case *Fields:
			if err := out.resolved(); err != nil {
				return err
			}
		}
	}

	// Channel
	if m.Channel != nil {
		m.Chan = true

		if m.Sub {
			return fmt.Errorf("invalid method %q: subservice methods cannot have channels", m.Name)
		}
	}
	return nil
}

// Input

// MethodInput is a union type for method input.
//
//	MethodInput:
//	| *Type
//	| *Fields
//	| nil
type MethodInput interface {
	resolve(file *File) error
}

func newMethodInput(p ast.MethodInput) (MethodInput, error) {
	switch p := p.(type) {
	case *ast.Type:
		return newType(p)
	case ast.Fields:
		return newFields(p)
	case nil:
		return nil, nil
	}

	panic("unsupported method input")
}

// Output

// MethodOutput is a union type for method output.
//
//	MethodOutput:
//	| *Type
//	| *Fields
//	| nil
type MethodOutput interface {
	resolve(file *File) error
}

func newMethodOutput(p ast.MethodOutput) (MethodOutput, error) {
	switch p := p.(type) {
	case *ast.Type:
		return newType(p)
	case ast.Fields:
		return newFields(p)
	case nil:
		return nil, nil
	}

	panic("unsupported method output")
}

// Channel

// MethodChannel defines in/out channel messages.
type MethodChannel struct {
	In  *Type
	Out *Type
}

func newMethodChannel(p *ast.MethodChannel) (*MethodChannel, error) {
	if p == nil {
		return nil, nil
	}

	if p.In == nil && p.Out == nil {
		return nil, fmt.Errorf("at least in or out must be specified")
	}

	var in *Type
	var err error
	if p.In != nil {
		in, err = newType(p.In)
		if err != nil {
			return nil, fmt.Errorf("invalid in: %w", err)
		}
	}

	var out *Type
	if p.Out != nil {
		out, err = newType(p.Out)
		if err != nil {
			return nil, fmt.Errorf("invalid out: %w", err)
		}
	}

	ch := &MethodChannel{
		In:  in,
		Out: out,
	}
	return ch, nil
}

func (ch *MethodChannel) resolve(file *File) error {
	if ch.In != nil {
		if err := ch.In.resolve(file); err != nil {
			return fmt.Errorf("in: %w", err)
		}
	}
	if ch.Out != nil {
		if err := ch.Out.resolve(file); err != nil {
			return fmt.Errorf("out: %w", err)
		}
	}
	return nil
}
