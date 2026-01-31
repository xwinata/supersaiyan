package supersaiyan

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

// Table represents a database table with its alias and relations (joins).
type Table struct {
	Name      string     `json:"name"                yaml:"name"                validate:"required"`
	Alias     string     `json:"alias"               yaml:"alias"               validate:"required"`
	Relations []Relation `json:"relations,omitempty" yaml:"relations,omitempty"`
}

// Relation represents a JOIN relationship between tables.
// The On field should contain Condition types (BoolOp, RangeOp, or WhereGroup).
type Relation struct {
	JoinType exp.JoinType `json:"joinType"     yaml:"joinType"`
	On       []any        `json:"on,omitempty" yaml:"on,omitempty"` // Should contain Condition types
	Table    Table        `json:"table"        yaml:"table"`
}

// join applies this relation as a JOIN clause to the given dataset.
// It recursively applies nested relations (joins on joined tables).
func (r Relation) join(ds *goqu.SelectDataset) *goqu.SelectDataset {
	onConds := make([]exp.Expression, 0, len(r.On))
	for _, on := range r.On {
		// Use type assertion with Condition interface for better type safety
		if cond, ok := on.(Condition); ok {
			onConds = append(onConds, cond.toExpression())
			continue
		}

		// Fallback to old behavior for backward compatibility
		var expr exp.Expression
		switch v := on.(type) {
		case BoolOp:
			expr = v.expression()
		case RangeOp:
			expr = v.expression()
		case WhereGroup:
			expr = v.expression()
		}

		if expr != nil {
			onConds = append(onConds, expr)
		}
	}

	// Apply the appropriate join type
	switch r.JoinType {
	case exp.InnerJoinType:
		ds = ds.InnerJoin(goqu.T(r.Table.Name).As(r.Table.Alias), goqu.On(onConds...))
	case exp.LeftJoinType:
		ds = ds.LeftJoin(goqu.T(r.Table.Name).As(r.Table.Alias), goqu.On(onConds...))
	case exp.RightJoinType:
		ds = ds.RightJoin(goqu.T(r.Table.Name).As(r.Table.Alias), goqu.On(onConds...))
	default:
		ds = ds.Join(goqu.T(r.Table.Name).As(r.Table.Alias), goqu.On(onConds...))
	}

	// Recursively apply nested joins
	for _, child := range r.Table.Relations {
		ds = child.join(ds)
	}

	return ds
}

// ParseJoinType converts a string to a goqu JoinType.
// Supported values: "left", "right", "inner" (default).
func ParseJoinType(s string) exp.JoinType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "left":
		return exp.LeftJoinType
	case "right":
		return exp.RightJoinType
	case "inner":
		return exp.InnerJoinType
	default:
		return exp.InnerJoinType
	}
}

// MarshalJSON implements custom JSON marshaling for Relation.
func (r Relation) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		JoinType string `json:"joinType"`
		On       []any  `json:"on,omitempty"`
		Table    Table  `json:"table"`
	}{
		JoinType: joinTypeToString(r.JoinType),
		On:       r.On,
		Table:    r.Table,
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for Relation.
func (r *Relation) UnmarshalJSON(data []byte) error {
	aux := &struct {
		JoinType string            `json:"joinType"`
		On       []json.RawMessage `json:"on,omitempty"`
		Table    Table             `json:"table"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	r.JoinType = stringToJoinType(aux.JoinType)
	r.Table = aux.Table

	// Unmarshal On conditions with type detection
	if len(aux.On) > 0 {
		r.On = make([]any, len(aux.On))
		for i, raw := range aux.On {
			condition, err := unmarshalCondition(raw)
			if err != nil {
				return fmt.Errorf("failed to unmarshal on condition at index %d: %w", i, err)
			}
			r.On[i] = condition
		}
	}

	return nil
}

// MarshalYAML implements custom YAML marshaling for Relation.
func (r Relation) MarshalYAML() (interface{}, error) {
	return &struct {
		JoinType string `yaml:"joinType"`
		On       []any  `yaml:"on,omitempty"`
		Table    Table  `yaml:"table"`
	}{
		JoinType: joinTypeToString(r.JoinType),
		On:       r.On,
		Table:    r.Table,
	}, nil
}

// UnmarshalYAML implements custom YAML unmarshaling for Relation.
func (r *Relation) UnmarshalYAML(unmarshal func(interface{}) error) error {
	aux := &struct {
		JoinType string                   `yaml:"joinType"`
		On       []map[string]interface{} `yaml:"on,omitempty"`
		Table    Table                    `yaml:"table"`
	}{}

	if err := unmarshal(&aux); err != nil {
		return err
	}

	r.JoinType = stringToJoinType(aux.JoinType)
	r.Table = aux.Table

	// Unmarshal On conditions with type detection
	if len(aux.On) > 0 {
		r.On = make([]any, len(aux.On))
		for i, onMap := range aux.On {
			// Convert map to JSON and then unmarshal using our JSON logic
			jsonData, err := json.Marshal(onMap)
			if err != nil {
				return fmt.Errorf("failed to marshal on to JSON: %w", err)
			}

			condition, err := unmarshalCondition(jsonData)
			if err != nil {
				return fmt.Errorf("failed to unmarshal on condition at index %d: %w", i, err)
			}
			r.On[i] = condition
		}
	}

	return nil
}

func joinTypeToString(jt exp.JoinType) string {
	switch jt {
	case exp.InnerJoinType:
		return "INNER"
	case exp.LeftJoinType:
		return "LEFT"
	case exp.RightJoinType:
		return "RIGHT"
	case exp.FullOuterJoinType:
		return "FULL OUTER"
	case exp.CrossJoinType:
		return "CROSS"
	default:
		return "INNER"
	}
}

func stringToJoinType(s string) exp.JoinType {
	switch s {
	case "INNER":
		return exp.InnerJoinType
	case "LEFT":
		return exp.LeftJoinType
	case "RIGHT":
		return exp.RightJoinType
	case "FULL OUTER":
		return exp.FullOuterJoinType
	case "CROSS":
		return exp.CrossJoinType
	default:
		return exp.InnerJoinType
	}
}
