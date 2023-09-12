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

	Input   *Type // Message
	Output  *Type // Message or service
	Channel *MethodChannel

	// Temp fields are transformed into input/output messages
	Temp struct {
		InputType   *Type
		InputFields *Fields

		OutputType   *Type
		OutputFields *Fields
	}
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
	if in := m.Temp.InputType; in != nil {
		if err := in.resolve(file); err != nil {
			return fmt.Errorf("%v: %w", m.Name, err)
		}
	}
	if in := m.Temp.InputFields; in != nil {
		if err := in.resolve(file); err != nil {
			return fmt.Errorf("%v: %w", m.Name, err)
		}
	}

	if out := m.Temp.OutputType; out != nil {
		if err := out.resolve(file); err != nil {
			return fmt.Errorf("%v: %w", m.Name, err)
		}
	}
	if out := m.Temp.OutputFields; out != nil {
		if err := out.resolve(file); err != nil {
			return fmt.Errorf("%v: %w", m.Name, err)
		}
	}
	return nil
}

func (m *Method) resolved() error {
	// Input
	if in := m.Temp.InputType; in != nil {
		if in.Kind != KindMessage {
			return fmt.Errorf("%v: single input must be a message, got %q instead", m.Name, in.Kind)
		}

		m.Temp.InputType = nil
		m.Input = in
	}
	if in := m.Temp.InputFields; in != nil {
		if err := in.resolved(); err != nil {
			return err
		}

		name := methodRequestName(m)
		msg, err := generateMessage(m.Package, m.File, name, in)
		if err != nil {
			return fmt.Errorf("%v: failed to generate request message: %w", m.Name, err)
		}

		m.Temp.InputFields = nil
		m.Input = newTypeRef(msg.Def)
	}

	// Output
	if out := m.Temp.OutputType; out != nil {
		switch out.Kind {
		case KindMessage:
			m.Temp.OutputType = nil
			m.Output = out

		case KindService:
			m.Sub = true
			m.Temp.OutputType = nil
			m.Output = out

		default:
			return fmt.Errorf("%v: single output must be a message or a service, got %q instead",
				m.Name, out.Kind)
		}
	}
	if out := m.Temp.OutputFields; out != nil {
		if err := out.resolved(); err != nil {
			return err
		}

		name := methodResponseName(m)
		msg, err := generateMessage(m.Package, m.File, name, out)
		if err != nil {
			return fmt.Errorf("%v: failed to generate response message: %w", m.Name, err)
		}

		m.Temp.OutputFields = nil
		m.Output = newTypeRef(msg.Def)
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
		m.Temp.InputType, err = newType(p)
		return err

	case ast.Fields:
		m.Temp.InputFields, err = newFields(p)
		return err

	case nil:
		return nil
	}

	panic("unsupported method input")
}

func makeMethodOutput(m *Method, p ast.MethodOutput) (err error) {
	switch p := p.(type) {
	case *ast.Type:
		m.Temp.OutputType, err = newType(p)
		return err

	case ast.Fields:
		m.Temp.OutputFields, err = newFields(p)
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
