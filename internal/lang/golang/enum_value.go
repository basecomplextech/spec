package golang

import (
	"fmt"
	"strings"

	"github.com/basecomplextech/spec/internal/lang/model"
)

type EnumValue struct {
	Name   string
	Value  string
	Number int
}

func newEnumValue(m *model.EnumValue) (*EnumValue, error) {
	name := name_upperCamelCase(m.Name)
	name = fmt.Sprintf("%v_%v", m.Enum.Def.Name, name)
	value := strings.ToLower(m.Name)

	v := &EnumValue{
		Name:   name,
		Value:  value,
		Number: m.Number,
	}
	return v, nil
}
