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

// FieldOption is a functional option for configuring a Field.
type FieldOption func(*Field)

// WithTable sets the table alias for a field.
func WithTable(tableAlias string) FieldOption {
	return func(f *Field) {
		f.TableAlias = tableAlias
	}
}

// WithAlias sets the field alias for a field.
func WithAlias(fieldAlias string) FieldOption {
	return func(f *Field) {
		f.FieldAlias = fieldAlias
	}
}

// F creates a Field reference with optional configuration.
// 
// Examples:
//   F("id")                                    // Simple field without table alias
//   F("id", WithTable("u"))                    // Field with table alias
//   F("created_at", WithTable("u"), WithAlias("reg_date")) // Field with table and field alias
//   F("name", WithAlias("full_name"))          // Field with field alias but no table alias
func F(name string, opts ...FieldOption) Field {
	f := Field{
		Name: name,
	}
	
	for _, opt := range opts {
		opt(&f)
	}
	
	return f
}

// Exp creates a Field with an expression and alias.
// This is a convenience function for creating computed/aggregate fields.
//
// Examples:
//   Exp("order_count", Literal{Value: "COUNT(?)", Args: []any{F("id", "o")}})
//   Exp("total", Literal{Value: "SUM(?)", Args: []any{F("amount", "o")}})
//   Exp("status_label", Case{...})
func Exp(alias string, expression any) Field {
	return Field{
		FieldAlias: alias,
		Exp:        expression,
	}
}
