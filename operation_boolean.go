package supersaiyan

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9/exp"
)

// BoolOp represents a boolean comparison operation (=, !=, >, <, LIKE, IN, etc.).
type BoolOp struct {
	Op         exp.BooleanOperation `json:"op"                   yaml:"op"`
	FieldName  string               `json:"fieldName"            yaml:"fieldName"`
	TableAlias string               `json:"tableAlias,omitempty" yaml:"tableAlias,omitempty"`
	Value      any                  `json:"value"                yaml:"value"`
}

// expression converts the BoolOp to a goqu boolean expression.
func (bo BoolOp) expression() exp.Expression {
	field := Field{
		Name:       bo.FieldName,
		TableAlias: bo.TableAlias,
	}

	switch bo.Op {
	case exp.EqOp:
		return field.identifierExpression().Eq(handleAny(bo.Value))
	case exp.NeqOp:
		return field.identifierExpression().Neq(handleAny(bo.Value))
	case exp.IsOp:
		return field.identifierExpression().Is(handleAny(bo.Value))
	case exp.IsNotOp:
		return field.identifierExpression().IsNot(handleAny(bo.Value))
	case exp.GtOp:
		return field.identifierExpression().Gt(handleAny(bo.Value))
	case exp.GteOp:
		return field.identifierExpression().Gte(handleAny(bo.Value))
	case exp.LtOp:
		return field.identifierExpression().Lt(handleAny(bo.Value))
	case exp.LteOp:
		return field.identifierExpression().Lte(handleAny(bo.Value))
	case exp.InOp:
		return field.identifierExpression().In(handleAny(bo.Value))
	case exp.NotInOp:
		return field.identifierExpression().NotIn(handleAny(bo.Value))
	case exp.LikeOp:
		return field.identifierExpression().Like(handleAny(bo.Value))
	case exp.NotLikeOp:
		return field.identifierExpression().NotLike(handleAny(bo.Value))
	case exp.ILikeOp:
		return field.identifierExpression().ILike(handleAny(bo.Value))
	case exp.NotILikeOp:
		return field.identifierExpression().NotILike(handleAny(bo.Value))
	case exp.RegexpLikeOp:
		return field.identifierExpression().RegexpLike(handleAny(bo.Value))
	case exp.RegexpNotLikeOp:
		return field.identifierExpression().RegexpNotLike(handleAny(bo.Value))
	case exp.RegexpILikeOp:
		return field.identifierExpression().RegexpILike(handleAny(bo.Value))
	case exp.RegexpNotILikeOp:
		return field.identifierExpression().RegexpNotILike(handleAny(bo.Value))
	default:
		return nil
	}
}

// ParseBoolOperation converts a string to a goqu BooleanOperation.
// Supported operators: =, !=, <>, >, >=, <, <=, IS, IS NOT, IN, NOT IN, LIKE, NOT LIKE, ILIKE, NOT ILIKE, ~, !~, ~*, !~*
func ParseBoolOperation(s string) exp.BooleanOperation {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "!=", "<>":
		return exp.NeqOp
	case "is":
		return exp.IsOp
	case "is not":
		return exp.IsNotOp
	case ">":
		return exp.GtOp
	case ">=":
		return exp.GteOp
	case "<":
		return exp.LtOp
	case "<=":
		return exp.LteOp
	case "in":
		return exp.InOp
	case "not in":
		return exp.NotInOp
	case "like":
		return exp.LikeOp
	case "not like":
		return exp.NotLikeOp
	case "ilike":
		return exp.ILikeOp
	case "not ilike":
		return exp.NotILikeOp
	case "~":
		return exp.RegexpLikeOp
	case "!~":
		return exp.RegexpNotLikeOp
	case "~*":
		return exp.RegexpILikeOp
	case "!~*":
		return exp.RegexpNotILikeOp
	case "=":
		fallthrough
	default:
		return exp.EqOp
	}
}

// BoolOperatorStrings contains all supported boolean operator strings, ordered by length (longest first).
// This ordering is important for parsing to avoid matching shorter operators first (e.g., "in" before "not in").
var BoolOperatorStrings = []string{
	" not ilike ",
	" not like ",
	" not in ",
	" is not ",
	"!~*",
	"~*",
	"!~",
	"<>",
	">=",
	"<=",
	"!=",
	"is",
	" in ",
	" ilike ",
	" like ",
	"~",
	">",
	"<",
	"=",
}

// MarshalJSON implements custom JSON marshaling for exp.BooleanOperation.
func (bo BoolOp) MarshalJSON() ([]byte, error) {
	type Alias BoolOp
	return json.Marshal(&struct {
		Op string `json:"op"`
		Alias
	}{
		Op:    boolOpToString(bo.Op),
		Alias: (Alias)(bo),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for BoolOp.
func (bo *BoolOp) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Op         string          `json:"op"`
		FieldName  string          `json:"fieldName"`
		TableAlias string          `json:"tableAlias,omitempty"`
		Value      json.RawMessage `json:"value"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	bo.Op = stringToBoolOp(aux.Op)
	bo.FieldName = aux.FieldName
	bo.TableAlias = aux.TableAlias

	// Try to unmarshal Value as an expression first
	if len(aux.Value) > 0 {
		value, err := unmarshalValue(aux.Value)
		if err != nil {
			return fmt.Errorf("failed to unmarshal value: %w", err)
		}
		bo.Value = value
	}

	return nil
}

// Eq creates an equality comparison (=).
func Eq(fieldName, tableAlias string, value any) BoolOp {
	return BoolOp{
		Op:         exp.EqOp,
		FieldName:  fieldName,
		TableAlias: tableAlias,
		Value:      value,
	}
}

// Neq creates a not-equal comparison (!=).
func Neq(fieldName, tableAlias string, value any) BoolOp {
	return BoolOp{
		Op:         exp.NeqOp,
		FieldName:  fieldName,
		TableAlias: tableAlias,
		Value:      value,
	}
}

// Gt creates a greater-than comparison (>).
func Gt(fieldName, tableAlias string, value any) BoolOp {
	return BoolOp{
		Op:         exp.GtOp,
		FieldName:  fieldName,
		TableAlias: tableAlias,
		Value:      value,
	}
}

// Gte creates a greater-than-or-equal comparison (>=).
func Gte(fieldName, tableAlias string, value any) BoolOp {
	return BoolOp{
		Op:         exp.GteOp,
		FieldName:  fieldName,
		TableAlias: tableAlias,
		Value:      value,
	}
}

// Lt creates a less-than comparison (<).
func Lt(fieldName, tableAlias string, value any) BoolOp {
	return BoolOp{
		Op:         exp.LtOp,
		FieldName:  fieldName,
		TableAlias: tableAlias,
		Value:      value,
	}
}

// Lte creates a less-than-or-equal comparison (<=).
func Lte(fieldName, tableAlias string, value any) BoolOp {
	return BoolOp{
		Op:         exp.LteOp,
		FieldName:  fieldName,
		TableAlias: tableAlias,
		Value:      value,
	}
}

// In creates an IN comparison.
func In(fieldName, tableAlias string, values any) BoolOp {
	return BoolOp{
		Op:         exp.InOp,
		FieldName:  fieldName,
		TableAlias: tableAlias,
		Value:      values,
	}
}

// NotIn creates a NOT IN comparison.
func NotIn(fieldName, tableAlias string, values any) BoolOp {
	return BoolOp{
		Op:         exp.NotInOp,
		FieldName:  fieldName,
		TableAlias: tableAlias,
		Value:      values,
	}
}

// Like creates a LIKE comparison.
func Like(fieldName, tableAlias string, pattern string) BoolOp {
	return BoolOp{
		Op:         exp.LikeOp,
		FieldName:  fieldName,
		TableAlias: tableAlias,
		Value:      pattern,
	}
}

// ILike creates a case-insensitive LIKE comparison.
func ILike(fieldName, tableAlias string, pattern string) BoolOp {
	return BoolOp{
		Op:         exp.ILikeOp,
		FieldName:  fieldName,
		TableAlias: tableAlias,
		Value:      pattern,
	}
}

// IsNull creates an IS NULL comparison.
func IsNull(fieldName, tableAlias string) BoolOp {
	return BoolOp{
		Op:         exp.IsOp,
		FieldName:  fieldName,
		TableAlias: tableAlias,
		Value:      nil,
	}
}

// IsNotNull creates an IS NOT NULL comparison.
func IsNotNull(fieldName, tableAlias string) BoolOp {
	return BoolOp{
		Op:         exp.IsNotOp,
		FieldName:  fieldName,
		TableAlias: tableAlias,
		Value:      nil,
	}
}
