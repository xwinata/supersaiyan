package supersaiyan

import (
	"encoding/json"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

// Literal represents a raw SQL expression with optional arguments.
// Use this for custom SQL expressions that aren't covered by the builder's API.
type Literal struct {
	Value string `json:"value"          yaml:"value"`
	Args  []any  `json:"args,omitempty" yaml:"args,omitempty"`
}

// expression converts the Literal to a goqu literal expression.
func (l Literal) expression() exp.LiteralExpression {
	argContainer := make([]any, len(l.Args))
	for i, arg := range l.Args {
		argContainer[i] = handleAny(arg)
	}

	return goqu.L(l.Value, argContainer...)
}

// UnmarshalJSON implements custom JSON unmarshaling for Literal.
func (l *Literal) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Value string            `json:"value"`
		Args  []json.RawMessage `json:"args,omitempty"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	l.Value = aux.Value

	// Unmarshal Args
	if len(aux.Args) > 0 {
		l.Args = make([]any, len(aux.Args))
		for i, raw := range aux.Args {
			arg, err := unmarshalValue(raw)
			if err != nil {
				return fmt.Errorf("failed to unmarshal arg at index %d: %w", i, err)
			}
			l.Args[i] = arg
		}
	}

	return nil
}
