// Copyright 2023 Ivan Korobkov. All rights reserved.

package model

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)

type MethodType string

const (
	MethodType_Undefined  MethodType = ""
	MethodType_Request    MethodType = "request"
	MethodType_Oneway     MethodType = "oneway"
	MethodType_Channel    MethodType = "channel"
	MethodType_Subservice MethodType = "subservice"
)

type Method struct {
	Package *Package
	File    *File
	Service *Service

	Name   string
	Type   MethodType
	Oneway bool // Oneway method

	Request    *Type // Message type
	Response   *Type // Message type
	Channel    *MethodChannel
	Subservice *Type // Subservice type

	_Input        *Type   // Temp, converted into Request
	_InputFields  *Fields // Temp, converted into Request
	_Output       *Type   // Temp, converted into Response or Subservice
	_OutputFields *Fields // Temp, converted into Response
}

func parseMethod(pkg *Package, file *File, service *Service, pm *syntax.Method) (*Method, error) {
	m := &Method{
		Package: pkg,
		File:    file,
		Service: service,

		Name:   pm.Name,
		Oneway: pm.Oneway,
	}

	if err := m.parseInput(pm); err != nil {
		return nil, err
	}
	if err := m.parseOutput(pm); err != nil {
		return nil, err
	}
	if err := m.parseChannel(pm); err != nil {
		return nil, err
	}
	return m, nil
}

// parse

func (m *Method) parse(pm *syntax.Method) error {
	if err := m.parseInput(pm); err != nil {
		return err
	}
	return nil
}

func (m *Method) parseInput(pm *syntax.Method) (err error) {
	switch in := pm.Input.(type) {
	case nil:
		return nil

	case *syntax.Type:
		m._Input, err = newType(in)
		return err

	case syntax.Fields:
		if len(in) > 0 {
			m._InputFields, err = newFields(in)
			return err
		}
		return nil
	}

	panic("unsupported method input")
}

func (m *Method) parseOutput(pm *syntax.Method) (err error) {
	switch out := pm.Output.(type) {
	case nil:
		return nil

	case *syntax.Type:
		m._Output, err = newType(out)
		return err

	case syntax.Fields:
		if len(out) > 0 {
			m._OutputFields, err = newFields(out)
			return err
		}
		return nil
	}

	panic("unsupported method output")
}

func (m *Method) parseChannel(pm *syntax.Method) (err error) {
	ch := pm.Channel
	if ch == nil {
		return nil
	}

	m.Channel, err = newMethodChannel(ch)
	return err
}

// resolve

func (m *Method) resolve(file *File) error {
	if in := m._Input; in != nil {
		if err := in.resolve(file); err != nil {
			return err
		}
	}
	if in := m._InputFields; in != nil {
		if err := in.resolve(file); err != nil {
			return err
		}
	}

	if out := m._Output; out != nil {
		if err := out.resolve(file); err != nil {
			return err
		}
	}
	if out := m._OutputFields; out != nil {
		if err := out.resolve(file); err != nil {
			return err
		}
	}

	if ch := m.Channel; ch != nil {
		if err := ch.resolve(file); err != nil {
			return err
		}
	}
	return nil
}

// compile

func (m *Method) compile() error {
	if err := m.compileInput(); err != nil {
		return err
	}
	if err := m.compileOutput(); err != nil {
		return err
	}
	if err := m.compileType(); err != nil {
		return err
	}
	return nil
}

func (m *Method) compileInput() error {
	if in := m._Input; in != nil {
		switch in.Kind {
		case KindMessage:
			m.Request = in

		default:
			return fmt.Errorf(
				"invalid input, single input must be a message, got %q instead",
				in.Kind)
		}
	}

	if in := m._InputFields; in != nil {
		if err := in.compile(); err != nil {
			return err
		}

		// Generate request message
		request, err := generateMethodRequest(m, in)
		if err != nil {
			return fmt.Errorf("failed to generate request: %w", err)
		}

		m.Request = request
		m._InputFields = nil
	}
	return nil
}

func (m *Method) compileOutput() error {
	if out := m._Output; out != nil {
		switch out.Kind {
		case KindMessage:
			m.Response = out

		case KindService:
			m.Subservice = out

		default:
			return fmt.Errorf(
				"invalid output, single output must be a message or a service, got %q instead",
				out.Kind)
		}
	}

	if out := m._OutputFields; out != nil {
		if err := out.compile(); err != nil {
			return err
		}

		// Convert into response message
		response, err := generateMethodResponse(m, out)
		if err != nil {
			return fmt.Errorf("failed to generate response: %w", err)
		}

		m.Response = response
		m._OutputFields = nil
	}
	return nil
}

func (m *Method) compileType() error {
	if m.Oneway {
		switch {
		case m.Response != nil:
			return fmt.Errorf("oneway method cannot return response")
		case m.Subservice != nil:
			return fmt.Errorf("oneway method cannot return subservice")
		case m.Channel != nil:
			return fmt.Errorf("oneway method cannot have channel")
		}

		m.Type = MethodType_Oneway
		return nil
	}

	if m.Channel != nil {
		switch {
		case m.Oneway:
			return fmt.Errorf("method with channel cannot be oneway")
		case m.Subservice != nil:
			return fmt.Errorf("method with channel cannot return subservice")
		}

		m.Type = MethodType_Channel
		return nil
	}

	if m.Subservice != nil {
		switch {
		case m.Oneway:
			return fmt.Errorf("method which returns subservice cannot be oneway")
		case m.Channel != nil:
			return fmt.Errorf("method which returns subservice cannot have channels")
		}

		m.Type = MethodType_Subservice
		return nil
	}

	m.Type = MethodType_Request
	return nil
}

// generate

func generateMethodRequest(m *Method, fields *Fields) (*Type, error) {
	// Make name
	service := m.Service.Def.Name
	method := toUpperCamelCase(m.Name)
	name := fmt.Sprintf("%v%vRequest", service, method)

	// Make message
	def, err := generateMessageDef(m.Package, m.File, name, fields)
	if err != nil {
		return nil, err
	}

	// Return type
	typ := newTypeRef(def)
	return typ, nil
}

func generateMethodResponse(m *Method, fields *Fields) (*Type, error) {
	// Make name
	service := m.Service.Def.Name
	method := toUpperCamelCase(m.Name)
	name := fmt.Sprintf("%v%vResponse", service, method)

	// Make message
	def, err := generateMessageDef(m.Package, m.File, name, fields)
	if err != nil {
		return nil, err
	}

	// Return type
	typ := newTypeRef(def)
	return typ, nil
}
