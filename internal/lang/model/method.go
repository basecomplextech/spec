package model

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)

type MethodType string

const (
	MethodType_Request    MethodType = "request"
	MethodType_Oneway     MethodType = "oneway"
	MethodType_Channel    MethodType = "channel"
	MethodType_Subservice MethodType = "subservice"
)

type Method struct {
	Package *Package
	File    *File
	Service *Service

	Name string
	Type MethodType

	Req    bool // Request/response method, response may be nil
	Sub    bool // Subservice method
	Chan   bool // Channel method
	Oneway bool // Oneway method

	Request    *Message
	Response   *Message
	Channel    *MethodChannel
	Subservice *Service

	_InputFields  *Fields // Temp, converted into Request
	_OutputFields *Fields // Temp, converted into Response
}

func newMethod(pkg *Package, file *File, service *Service, pm *syntax.Method) (*Method, error) {
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
	if in := m._InputFields; in != nil {
		if err := in.resolve(file); err != nil {
			return fmt.Errorf("%v: %w", m.Name, err)
		}
	}

	if out := m.Output; out != nil {
		if err := out.resolve(file); err != nil {
			return fmt.Errorf("%v: %w", m.Name, err)
		}
	}
	if out := m._OutputFields; out != nil {
		if err := out.resolve(file); err != nil {
			return fmt.Errorf("%v: %w", m.Name, err)
		}
	}

	if ch := m.Channel; ch != nil {
		if err := ch.resolve(file); err != nil {
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
	if in := m._InputFields; in != nil {
		if err := in.resolved(); err != nil {
			return err
		}

		// Convert into request message
		name := methodRequestName(m)
		msg, err := generateMessage(m.Package, m.File, name, in)
		if err != nil {
			return fmt.Errorf("%v: failed to generate request message: %w", m.Name, err)
		}

		m.Input = newTypeRef(msg.Def)
		m._InputFields = nil
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
	if out := m._OutputFields; out != nil {
		if err := out.resolved(); err != nil {
			return err
		}

		// Convert into response message
		name := methodResponseName(m)
		msg, err := generateMessage(m.Package, m.File, name, out)
		if err != nil {
			return fmt.Errorf("%v: failed to generate response message: %w", m.Name, err)
		}

		m.Output = newTypeRef(msg.Def)
		m._OutputFields = nil
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

func makeMethodInput(m *Method, p syntax.MethodInput) (err error) {
	switch p := p.(type) {
	case *syntax.Type:
		m.Input, err = newType(p)
		return err

	case syntax.Fields:
		if len(p) > 0 {
			m._InputFields, err = newFields(p)
			return err
		}
		return nil

	case nil:
		return nil
	}

	panic("unsupported method input")
}

func makeMethodOutput(m *Method, p syntax.MethodOutput) (err error) {
	switch p := p.(type) {
	case *syntax.Type:
		m.Output, err = newType(p)
		return err

	case syntax.Fields:
		if len(p) > 0 {
			m._OutputFields, err = newFields(p)
			return err
		}
		return nil

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

func newMethodChannel(p *syntax.MethodChannel) (*MethodChannel, error) {
	if p == nil {
		return nil, nil
	}

	if p.In == nil && p.Out == nil {
		return nil, fmt.Errorf("at lesyntax in or out must be specified")
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
