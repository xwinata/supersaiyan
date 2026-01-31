package supersaiyan

import (
	"encoding/json"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

// WhereGroup represents a group of WHERE conditions combined with AND or OR.
type WhereGroup struct {
	Op         exp.ExpressionListType `json:"op"         yaml:"op"`
	Conditions []any                  `json:"conditions" yaml:"conditions"`
}

// expression converts the WhereGroup to a goqu expression.
// It recursively handles nested groups and combines conditions with the specified operator.
func (wg WhereGroup) expression() exp.Expression {
	exps := make([]exp.Expression, 0, len(wg.Conditions))

	for _, cond := range wg.Conditions {
		var expr exp.Expression

		switch v := cond.(type) {
		case BoolOp:
			expr = v.expression()
		case RangeOp:
			expr = v.expression()
		case WhereGroup:
			expr = v.expression()
		}

		if expr != nil {
			exps = append(exps, expr)
		}
	}

	// Return early if no valid expressions
	if len(exps) == 0 {
		return nil
	}

	switch wg.Op {
	case exp.OrType:
		return goqu.Or(exps...)
	default:
		return goqu.And(exps...)
	}
}

// MarshalJSON implements custom JSON marshaling for WhereGroup.
func (wg WhereGroup) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Op         string `json:"op"`
		Conditions []any  `json:"conditions"`
	}{
		Op:         expressionListTypeToString(wg.Op),
		Conditions: wg.Conditions,
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for WhereGroup.
func (wg *WhereGroup) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Op         string            `json:"op"`
		Conditions []json.RawMessage `json:"conditions"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	wg.Op = stringToExpressionListType(aux.Op)

	// Unmarshal Conditions with type detection
	if len(aux.Conditions) > 0 {
		wg.Conditions = make([]any, len(aux.Conditions))
		for i, raw := range aux.Conditions {
			condition, err := unmarshalCondition(raw)
			if err != nil {
				return fmt.Errorf("failed to unmarshal condition at index %d: %w", i, err)
			}
			wg.Conditions[i] = condition
		}
	}

	return nil
}

func expressionListTypeToString(elt exp.ExpressionListType) string {
	switch elt {
	case exp.AndType:
		return "AND"
	case exp.OrType:
		return "OR"
	default:
		return "AND"
	}
}

func stringToExpressionListType(s string) exp.ExpressionListType {
	switch s {
	case "OR":
		return exp.OrType
	case "AND":
		return exp.AndType
	default:
		return exp.AndType
	}
}

// And creates an AND group of conditions.
func And(conditions ...any) WhereGroup {
	return WhereGroup{
		Op:         exp.AndType,
		Conditions: conditions,
	}
}

// Or creates an OR group of conditions.
func Or(conditions ...any) WhereGroup {
	return WhereGroup{
		Op:         exp.OrType,
		Conditions: conditions,
	}
}
