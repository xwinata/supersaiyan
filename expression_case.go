package supersaiyan

import (
	"encoding/json"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

// Case represents a SQL CASE expression with multiple WHEN/THEN conditions and an optional ELSE.
type Case struct {
	Conditions []WhenThen `json:"conditions"     yaml:"conditions"`
	Else       any        `json:"else,omitempty" yaml:"else,omitempty"`
}

// WhenThen represents a single WHEN condition and its THEN result.
type WhenThen struct {
	When any `json:"when" yaml:"when"`
	Then any `json:"then" yaml:"then"`
}

// expression converts the Case to a goqu case expression.
func (c Case) expression() exp.CaseExpression {
	caseExpr := goqu.Case()

	for _, cond := range c.Conditions {
		caseExpr = caseExpr.When(handleAny(cond.When), handleAny(cond.Then))
	}

	if c.Else != nil {
		caseExpr = caseExpr.Else(handleAny(c.Else))
	}

	return caseExpr
}

// UnmarshalJSON implements custom JSON unmarshaling for WhenThen.
func (wt *WhenThen) UnmarshalJSON(data []byte) error {
	aux := &struct {
		When json.RawMessage `json:"when"`
		Then json.RawMessage `json:"then"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Unmarshal When - try as condition first, then as expression
	if len(aux.When) > 0 {
		when, err := unmarshalCondition(aux.When)
		if err != nil {
			// Try as expression
			when, err = unmarshalExpression(aux.When)
			if err != nil {
				// Try as simple value
				var simpleValue any
				if err := json.Unmarshal(aux.When, &simpleValue); err != nil {
					return fmt.Errorf("failed to unmarshal when: %w", err)
				}
				wt.When = simpleValue
			} else {
				wt.When = when
			}
		} else {
			wt.When = when
		}
	}

	// Unmarshal Then
	if len(aux.Then) > 0 {
		then, err := unmarshalValue(aux.Then)
		if err != nil {
			return fmt.Errorf("failed to unmarshal then value: %w", err)
		}
		wt.Then = then
	}

	return nil
}

// UnmarshalJSON implements custom JSON unmarshaling for Case.
func (c *Case) UnmarshalJSON(data []byte) error {
	type Alias Case
	aux := &struct {
		Else json.RawMessage `json:"else,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Unmarshal Else if present
	if len(aux.Else) > 0 {
		elseVal, err := unmarshalExpression(aux.Else)
		if err != nil {
			// If it fails, try as a simple value
			var simpleValue any
			if err := json.Unmarshal(aux.Else, &simpleValue); err != nil {
				return fmt.Errorf("failed to unmarshal else value: %w", err)
			}
			c.Else = simpleValue
		} else {
			c.Else = elseVal
		}
	}

	return nil
}

// WT creates a WhenThen condition for use in Case expressions.
//
// Examples:
//
//	WT(Eq("status", "u", "active"), "Active")
//	WT(Gt("age", "u", 18), "Adult")
func WT(condition any, thenValue any) WhenThen {
	return WhenThen{
		When: condition,
		Then: thenValue,
	}
}

// C creates a CASE expression with WHEN/THEN conditions and optional ELSE value.
//
// Examples:
//
//	C(nil, WT(Eq("status", "u", "active"), "Active"), WT(Eq("status", "u", "inactive"), "Inactive"))
//	C("Unknown", WT(Eq("status", "u", "active"), "Active"), WT(Eq("status", "u", "inactive"), "Inactive"))
//	C("Adult", WT(Lt("age", "u", 18), "Minor"))
func C(elseValue any, conditions ...WhenThen) Case {
	return Case{
		Conditions: conditions,
		Else:       elseValue,
	}
}
