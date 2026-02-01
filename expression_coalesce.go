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

// Coal creates a COALESCE expression from fields with optional default value.
// Returns the first non-NULL value from the provided fields, or the default if all are NULL.
//
// Examples:
//   Coal(nil, F("nickname", WithTable("u")), F("username", WithTable("u")))
//   Coal("Anonymous", F("nickname", WithTable("u")), F("username", WithTable("u")))
//   Coal("N/A", F("phone", WithTable("u")), F("email", WithTable("u")))
func Coal(defaultValue any, fields ...Field) Coalesce {
	return Coalesce{
		Fields:       fields,
		DefaultValue: defaultValue,
	}
}
