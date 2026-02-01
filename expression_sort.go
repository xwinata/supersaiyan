package supersaiyan

import (
	"encoding/json"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

// Sort represents an ORDER BY clause with column name, table alias, and direction.
type Sort struct {
	Name       string            `json:"name"                 yaml:"name"`
	TableAlias string            `json:"tableAlias,omitempty" yaml:"tableAlias,omitempty"`
	Order      exp.SortDirection `json:"order"                yaml:"order"`
}

// expression converts the Sort to a goqu ordered expression.
func (s Sort) expression() exp.OrderedExpression {
	switch s.Order {
	case exp.DescSortDir:
		return goqu.C(s.Name).Table(s.TableAlias).Desc()
	default:
		return goqu.C(s.Name).Table(s.TableAlias).Asc()
	}
}

// ParseSortDirection converts a string to a goqu SortDirection.
// Supported values: "DESC", "ASC" (default).
func ParseSortDirection(s string) exp.SortDirection {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DESC":
		return exp.DescSortDir
	default:
		return exp.AscDir
	}
}

// MarshalJSON implements custom JSON marshaling for Sort.
func (s Sort) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name       string `json:"name"`
		TableAlias string `json:"tableAlias,omitempty"`
		Order      string `json:"order"`
	}{
		Name:       s.Name,
		TableAlias: s.TableAlias,
		Order:      sortDirectionToString(s.Order),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for Sort.
func (s *Sort) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Name       string `json:"name"`
		TableAlias string `json:"tableAlias,omitempty"`
		Order      string `json:"order"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	s.Name = aux.Name
	s.TableAlias = aux.TableAlias
	s.Order = stringToSortDirection(aux.Order)

	return nil
}

// MarshalYAML implements custom YAML marshaling for Sort.
func (s Sort) MarshalYAML() (interface{}, error) {
	return &struct {
		Name       string `yaml:"name"`
		TableAlias string `yaml:"tableAlias,omitempty"`
		Order      string `yaml:"order"`
	}{
		Name:       s.Name,
		TableAlias: s.TableAlias,
		Order:      sortDirectionToString(s.Order),
	}, nil
}

// UnmarshalYAML implements custom YAML unmarshaling for Sort.
func (s *Sort) UnmarshalYAML(unmarshal func(interface{}) error) error {
	aux := &struct {
		Name       string `yaml:"name"`
		TableAlias string `yaml:"tableAlias,omitempty"`
		Order      string `yaml:"order"`
	}{}

	if err := unmarshal(&aux); err != nil {
		return err
	}

	s.Name = aux.Name
	s.TableAlias = aux.TableAlias
	s.Order = stringToSortDirection(aux.Order)

	return nil
}

func sortDirectionToString(sd exp.SortDirection) string {
	switch sd {
	case exp.AscDir:
		return "ASC"
	case exp.DescSortDir:
		return "DESC"
	default:
		return "ASC"
	}
}

func stringToSortDirection(s string) exp.SortDirection {
	switch s {
	case "DESC":
		return exp.DescSortDir
	case "ASC":
		return exp.AscDir
	default:
		return exp.AscDir
	}
}

// Asc creates an ascending sort direction.
func Asc(name, tableAlias string) Sort {
	return Sort{
		Name:       name,
		TableAlias: tableAlias,
		Order:      exp.AscDir,
	}
}

// Desc creates a descending sort direction.
func Desc(name, tableAlias string) Sort {
	return Sort{
		Name:       name,
		TableAlias: tableAlias,
		Order:      exp.DescSortDir,
	}
}
