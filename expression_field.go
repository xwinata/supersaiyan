package supersaiyan

import (
	"encoding/json"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

// Field represents a database column with optional table alias and field alias.
// It can also contain complex expressions like CASE, COALESCE, or literals.
type Field struct {
	Name       string `json:"name,omitempty"       yaml:"name,omitempty"`
	TableAlias string `json:"tableAlias,omitempty" yaml:"tableAlias,omitempty"`
	FieldAlias string `json:"fieldAlias,omitempty" yaml:"fieldAlias,omitempty"`
	Exp        any    `json:"exp,omitempty"        yaml:"exp,omitempty"`
}

// expression converts the Field to a goqu expression.
// It handles aliased fields, complex expressions, and simple column references.
func (f Field) expression() exp.Expression {
	if f.Exp != nil {
		var opt handleAnyOption
		if f.aliased() {
			opt = withAlias(f.FieldAlias)
		} else if f.Name != "" {
			opt = withAlias(f.Name)
		}

		return handleAny(f.Exp, opt)
	}

	if f.aliased() {
		return f.aliasedExpression()
	}

	return f.identifierExpression()
}

// aliased returns true if the field has an alias.
func (f Field) aliased() bool {
	return f.FieldAlias != ""
}

// identifierExpression returns a simple column identifier with optional table alias.
func (f Field) identifierExpression() exp.IdentifierExpression {
	return goqu.C(f.Name).Table(f.TableAlias)
}

// aliasedExpression returns the field expression with an alias.
func (f Field) aliasedExpression() exp.Expression {
	if f.Exp != nil {
		return handleAny(f.Exp, withAlias(f.FieldAlias))
	}
	return f.identifierExpression().As(f.FieldAlias)
}

// UnmarshalJSON implements custom JSON unmarshaling for Field.
func (f *Field) UnmarshalJSON(data []byte) error {
	type Alias Field
	aux := &struct {
		Exp json.RawMessage `json:"exp,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(f),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Unmarshal Exp with type detection
	if len(aux.Exp) > 0 {
		exp, err := unmarshalExpression(aux.Exp)
		if err != nil {
			return fmt.Errorf("failed to unmarshal field expression: %w", err)
		}
		f.Exp = exp
	}

	return nil
}

// UnmarshalYAML implements custom YAML unmarshaling for Field.
func (f *Field) UnmarshalYAML(unmarshal func(interface{}) error) error {
	aux := &struct {
		Name       string                 `yaml:"name,omitempty"`
		TableAlias string                 `yaml:"tableAlias,omitempty"`
		FieldAlias string                 `yaml:"fieldAlias,omitempty"`
		Exp        map[string]interface{} `yaml:"exp,omitempty"`
	}{}

	if err := unmarshal(&aux); err != nil {
		return err
	}

	f.Name = aux.Name
	f.TableAlias = aux.TableAlias
	f.FieldAlias = aux.FieldAlias

	// Unmarshal Exp with type detection
	if len(aux.Exp) > 0 {
		// Convert map to JSON and then unmarshal using our JSON logic
		jsonData, err := json.Marshal(aux.Exp)
		if err != nil {
			return fmt.Errorf("failed to marshal exp to JSON: %w", err)
		}

		exp, err := unmarshalExpression(jsonData)
		if err != nil {
			return fmt.Errorf("failed to unmarshal field expression: %w", err)
		}
		f.Exp = exp
	}

	return nil
}

// F creates a Field reference.
func F(name, tableAlias string) Field {
	return Field{
		Name:       name,
		TableAlias: tableAlias,
	}
}
