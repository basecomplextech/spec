package model

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/ast"
)

type Method struct {
	Package *Package
	File    *File
	Service *Service

	Name string
	Sub  bool // Returns subservice
	Chan bool // Returns channel

	Input       *Type // Message
	InputFields *Fields

	Output       *Type // Message or service
	OutputFields *Fields

	Channel *MethodChannel
}

func newMethod(pkg *Package, file *File, service *Service, pm *ast.Method) (*Method, error) {
	m := &Method{
		Package: pkg,
		File:    file,
		Service: service,

		Name: pm.Name,
	}

	// Input
	if pm.Input != nil {
		if err := makeMethodInput(m, pm.Input); err != nil {
			return nil, fmt.Errorf("%v: %w", m.Name, err)
		}
	}

	// Output
	if pm.Output != nil {
		if err := makeMethodOutput(m, pm.Output); err != nil {
			return nil, fmt.Errorf("%v: %w", m.Name, err)
		}
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
	if in := m.Input; in != nil {
		if err := in.resolve(file); err != nil {
			return fmt.Errorf("%v: %w", m.Name, err)
		}
	}
	if in := m.InputFields; in != nil {
		if err := in.resolve(file); err != nil {
			return fmt.Errorf("%v: %w", m.Name, err)
		}
	}

	if out := m.Output; out != nil {
		if err := out.resolve(file); err != nil {
			return fmt.Errorf("%v: %w", m.Name, err)
		}
	}
	if out := m.OutputFields; out != nil {
		if err := out.resolve(file); err != nil {
			return fmt.Errorf("%v: %w", m.Name, err)
		}
	}
	return nil
}

func (m *Method) resolved() error {
	// Input
	if in := m.Input; in != nil {
		if in.Kind != KindMessage {
			return fmt.Errorf("%v: single input must be a message, got %q instead", m.Name, in.Kind)
		}
	}
	if in := m.InputFields; in != nil {
		if err := in.resolved(); err != nil {
			return err
		}

		// Convert into request message
		// Client input accepts any value types as arguments.
		if !in.value() {
			name := methodRequestName(m)
			msg, err := generateMessage(m.Package, m.File, name, in)
			if err != nil {
				return fmt.Errorf("%v: failed to generate request message: %w", m.Name, err)
			}

			m.Input = newTypeRef(msg.Def)
			m.InputFields = nil
		}
	}

	// Output
	if out := m.Output; out != nil {
		switch out.Kind {
		case KindMessage:
		case KindService:
			m.Sub = true
		default:
			return fmt.Errorf("%v: single output must be a message or a service, got %q instead",
				m.Name, out.Kind)
		}
	}
	if out := m.OutputFields; out != nil {
		if err := out.resolved(); err != nil {
			return err
		}

		// Convert into response message
		// Client output accepts only non-allocated primitive types as return values.
		if !out.primitive() {
			name := methodResponseName(m)
			msg, err := generateMessage(m.Package, m.File, name, out)
			if err != nil {
				return fmt.Errorf("%v: failed to generate response message: %w", m.Name, err)
			}

			m.OutputFields = nil
			m.Output = newTypeRef(msg.Def)
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

// Input/output

func makeMethodInput(m *Method, p ast.MethodInput) (err error) {
	switch p := p.(type) {
	case *ast.Type:
		m.Input, err = newType(p)
		return err

	case ast.Fields:
		m.InputFields, err = newFields(p)
		return err

	case nil:
		return nil
	}

	panic("unsupported method input")
}

func makeMethodOutput(m *Method, p ast.MethodOutput) (err error) {
	switch p := p.(type) {
	case *ast.Type:
		m.Output, err = newType(p)
		return err

	case ast.Fields:
		m.OutputFields, err = newFields(p)
		return err

	case nil:
		return nil
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

// Request/Response

func methodRequestName(m *Method) string {
	service := m.Service.Def.Name
	method := toUpperCamelCase(m.Name)

	return fmt.Sprintf("%v%vRequest", service, method)
}

func methodResponseName(m *Method) string {
	service := m.Service.Def.Name
	method := toUpperCamelCase(m.Name)

	return fmt.Sprintf("%v%vResponse", service, method)
}
