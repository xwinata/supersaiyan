package supersaiyan

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

// Coalesce represents a SQL COALESCE function that returns the first non-NULL value.
type Coalesce struct {
	Fields       []Field `json:"fields"                 yaml:"fields"`
	DefaultValue any     `json:"defaultValue,omitempty" yaml:"defaultValue,omitempty"`
}

// expression converts the Coalesce to a goqu SQL function expression.
func (co Coalesce) expression() exp.SQLFunctionExpression {
	fields := make([]any, 0, len(co.Fields)+1)

	for _, f := range co.Fields {
		fields = append(fields, f.expression())
	}

	if co.DefaultValue != nil {
		fields = append(fields, handleAny(co.DefaultValue))
	}

	return goqu.COALESCE(fields...)
}
