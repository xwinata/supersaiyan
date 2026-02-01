package supersaiyan

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

// RangeOp represents a BETWEEN or NOT BETWEEN operation.
type RangeOp struct {
	Op         exp.RangeOperation `json:"op"                   yaml:"op"`
	FieldName  string             `json:"fieldName"            yaml:"fieldName"`
	TableAlias string             `json:"tableAlias,omitempty" yaml:"tableAlias,omitempty"`
	Start      any                `json:"start"                yaml:"start"`
	End        any                `json:"end"                  yaml:"end"`
}

// expression converts the RangeOp to a goqu range expression.
func (ro RangeOp) expression() exp.Expression {
	field := Field{
		Name:       ro.FieldName,
		TableAlias: ro.TableAlias,
	}

	rangeVal := goqu.Range(handleAny(ro.Start), handleAny(ro.End))

	switch ro.Op {
	case exp.NotBetweenOp:
		return field.identifierExpression().NotBetween(rangeVal)
	default:
		return field.identifierExpression().Between(rangeVal)
	}
}

// ParseRangeOperation converts a string to a goqu RangeOperation.
// Supported values: "between" (default), "not between", "not".
func ParseRangeOperation(s string) exp.RangeOperation {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "not between", "not":
		return exp.NotBetweenOp
	default:
		return exp.BetweenOp
	}
}

// MarshalJSON implements custom JSON marshaling for exp.RangeOperation.
func (ro RangeOp) MarshalJSON() ([]byte, error) {
	type Alias RangeOp
	return json.Marshal(&struct {
		Op string `json:"op"`
		Alias
	}{
		Op:    rangeOpToString(ro.Op),
		Alias: (Alias)(ro),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for RangeOp.
func (ro *RangeOp) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Op         string          `json:"op"`
		FieldName  string          `json:"fieldName"`
		TableAlias string          `json:"tableAlias,omitempty"`
		Start      json.RawMessage `json:"start"`
		End        json.RawMessage `json:"end"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	ro.Op = stringToRangeOp(aux.Op)
	ro.FieldName = aux.FieldName
	ro.TableAlias = aux.TableAlias

	// Unmarshal Start
	if len(aux.Start) > 0 {
		start, err := unmarshalValue(aux.Start)
		if err != nil {
			return fmt.Errorf("failed to unmarshal start: %w", err)
		}
		ro.Start = start
	}

	// Unmarshal End
	if len(aux.End) > 0 {
		end, err := unmarshalValue(aux.End)
		if err != nil {
			return fmt.Errorf("failed to unmarshal end: %w", err)
		}
		ro.End = end
	}

	return nil
}

// Between creates a BETWEEN comparison.
func Between(fieldName, tableAlias string, start, end any) RangeOp {
	return RangeOp{
		Op:         exp.BetweenOp,
		FieldName:  fieldName,
		TableAlias: tableAlias,
		Start:      start,
		End:        end,
	}
}

// NotBetween creates a NOT BETWEEN comparison.
func NotBetween(fieldName, tableAlias string, start, end any) RangeOp {
	return RangeOp{
		Op:         exp.NotBetweenOp,
		FieldName:  fieldName,
		TableAlias: tableAlias,
		Start:      start,
		End:        end,
	}
}
