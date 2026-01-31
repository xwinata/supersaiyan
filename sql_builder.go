package supersaiyan

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

// ErrMissingWhereCondition is returned when Edit or Delete is called without WHERE conditions.
var ErrMissingWhereCondition = errors.New(
	"WHERE condition is required for Edit and Delete operations",
)

// applyLimitOffset adds LIMIT and OFFSET clauses to the query.
func (qb *SQLBuilder) applyLimitOffset(ds *goqu.SelectDataset) *goqu.SelectDataset {
	if qb.limit > 0 {
		ds = ds.Limit(qb.limit)
	}
	if qb.offset > 0 {
		ds = ds.Offset(qb.offset)
	}
	return ds
}

// SQLBuilder constructs SQL queries using a fluent interface.
// It supports SELECT, INSERT, UPDATE, and DELETE operations with joins, filters, and sorting.
// All queries use prepared statements by default for security.
type SQLBuilder struct {
	Dialect string  `json:"dialect"           yaml:"dialect"`
	Fields  []Field `json:"fields,omitempty"  yaml:"fields,omitempty"`
	Table   Table   `json:"table"             yaml:"table"`
	Wheres  []any   `json:"wheres,omitempty"  yaml:"wheres,omitempty"` // Should contain Condition types (BoolOp, RangeOp, WhereGroup)
	Sorts   []Sort  `json:"sorts,omitempty"   yaml:"sorts,omitempty"`
	GroupBy []Field `json:"groupBy,omitempty" yaml:"groupBy,omitempty"`
	limit   uint
	offset  uint
}

// New creates a new SQLBuilder with the specified dialect and table.
// The default limit is 10. Use Limit(0) to remove the limit, or Limit(n) to set a different limit.
func New(dialect string, tableName string, tableAlias string) *SQLBuilder {
	return &SQLBuilder{
		Dialect: dialect,
		Table: Table{
			Name:  tableName,
			Alias: tableAlias,
		},
		limit: 10,
	}
}

// WithFields adds multiple fields to select.
func (qb *SQLBuilder) WithFields(fields ...Field) *SQLBuilder {
	qb.Fields = append(qb.Fields, fields...)
	return qb
}

// Where adds WHERE conditions.
func (qb *SQLBuilder) Where(conditions ...Condition) *SQLBuilder {
	for _, cond := range conditions {
		qb.Wheres = append(qb.Wheres, cond)
	}
	return qb
}

// OrderBy adds sorting.
func (qb *SQLBuilder) OrderBy(sorts ...Sort) *SQLBuilder {
	qb.Sorts = append(qb.Sorts, sorts...)
	return qb
}

// GroupByFields adds GROUP BY fields.
func (qb *SQLBuilder) GroupByFields(fields ...Field) *SQLBuilder {
	qb.GroupBy = append(qb.GroupBy, fields...)
	return qb
}

// Join adds a join relation.
func (qb *SQLBuilder) Join(joinType exp.JoinType, table Table, on ...Condition) *SQLBuilder {
	onAny := make([]any, len(on))
	for i, cond := range on {
		onAny[i] = cond
	}
	qb.Table.Relations = append(qb.Table.Relations, Relation{
		JoinType: joinType,
		Table:    table,
		On:       onAny,
	})
	return qb
}

// InnerJoin adds an INNER JOIN.
func (qb *SQLBuilder) InnerJoin(tableName, tableAlias string, on ...Condition) *SQLBuilder {
	return qb.Join(exp.InnerJoinType, Table{Name: tableName, Alias: tableAlias}, on...)
}

// LeftJoin adds a LEFT JOIN.
func (qb *SQLBuilder) LeftJoin(tableName, tableAlias string, on ...Condition) *SQLBuilder {
	return qb.Join(exp.LeftJoinType, Table{Name: tableName, Alias: tableAlias}, on...)
}

// RightJoin adds a RIGHT JOIN.
func (qb *SQLBuilder) RightJoin(tableName, tableAlias string, on ...Condition) *SQLBuilder {
	return qb.Join(exp.RightJoinType, Table{Name: tableName, Alias: tableAlias}, on...)
}

// mainSelect builds the base SELECT query with joins, fields, filters, sorting, and grouping.
func (qb *SQLBuilder) mainSelect() *goqu.SelectDataset {
	ds := goqu.From(goqu.T(qb.Table.Name).As(qb.Table.Alias)).WithDialect(qb.Dialect)

	// Apply joins
	for _, rel := range qb.Table.Relations {
		ds = rel.join(ds)
	}

	// Apply field selection
	if len(qb.Fields) > 0 {
		selects := make([]any, len(qb.Fields))
		for i, f := range qb.Fields {
			selects[i] = f.expression()
		}
		ds = ds.Select(selects...)
	}

	// Apply WHERE conditions
	if len(qb.Wheres) > 0 {
		expressions := make([]exp.Expression, len(qb.Wheres))
		for i, w := range qb.Wheres {
			expressions[i] = handleAny(w)
		}
		ds = ds.Where(expressions...)
	}

	// Apply sorting
	if len(qb.Sorts) > 0 {
		orders := make([]exp.OrderedExpression, len(qb.Sorts))
		for i, s := range qb.Sorts {
			orders[i] = s.expression()
		}
		ds = ds.Order(orders...)
	}

	// Apply grouping
	if len(qb.GroupBy) > 0 {
		groupFields := make([]any, len(qb.GroupBy))
		for i, g := range qb.GroupBy {
			if g.Name != "" {
				if g.TableAlias != "" {
					groupFields[i] = g.identifierExpression()
				} else {
					groupFields[i] = g.Name
				}
			} else if g.aliased() {
				groupFields[i] = g.FieldAlias
			}
		}
		ds = ds.GroupBy(groupFields...)
	}

	return ds
}

// Count generates a COUNT query and returns the SQL string, arguments, and any error.
// Uses prepared statements by default for security.
func (qb *SQLBuilder) Count() (string, []any, error) {
	ds := qb.mainSelect()

	// Apply chained options
	ds = qb.applyLimitOffset(ds)
	ds = ds.Prepared(true)

	return ds.Select(goqu.COUNT(goqu.Star())).ToSQL()
}

// Select generates a SELECT query and returns the SQL string, arguments, and any error.
// Uses prepared statements by default for security.
func (qb *SQLBuilder) Select() (string, []any, error) {
	ds := qb.mainSelect()

	// Apply chained options
	ds = qb.applyLimitOffset(ds)
	ds = ds.Prepared(true)

	return ds.ToSQL()
}

// Limit adds a LIMIT clause and returns the query for chaining.
func (qb *SQLBuilder) Limit(limit uint) *SQLBuilder {
	qb.limit = limit
	return qb
}

// Offset adds an OFFSET clause and returns the query for chaining.
func (qb *SQLBuilder) Offset(offset uint) *SQLBuilder {
	qb.offset = offset
	return qb
}

// Add generates an INSERT query and returns the SQL string, arguments, and any error.
// Uses prepared statements by default for security.
func (qb *SQLBuilder) Add(entry map[string]any) (string, []any, error) {
	ds := goqu.Insert(goqu.T(qb.Table.Name)).
		WithDialect(qb.Dialect).
		Rows(goqu.Record(entry)).
		Prepared(true)

	return ds.ToSQL()
}

// Edit generates an UPDATE query and returns the SQL string, arguments, and any error.
// Requires WHERE conditions to be set via Where() method to prevent accidental updates.
// Uses prepared statements by default for security.
func (qb *SQLBuilder) Edit(entry map[string]any) (string, []any, error) {
	if len(qb.Wheres) == 0 {
		return "", nil, ErrMissingWhereCondition
	}

	ds := goqu.Update(goqu.T(qb.Table.Name)).WithDialect(qb.Dialect)

	// Apply WHERE conditions from builder
	expressions := make([]exp.Expression, len(qb.Wheres))
	for i, w := range qb.Wheres {
		expressions[i] = handleAny(w)
	}
	ds = ds.Where(expressions...)

	ds = ds.Set(goqu.Record(entry)).Prepared(true)

	return ds.ToSQL()
}

// Delete generates a DELETE query and returns the SQL string, arguments, and any error.
// Requires WHERE conditions to be set via Where() method to prevent accidental deletes.
// Uses prepared statements by default for security.
func (qb *SQLBuilder) Delete() (string, []any, error) {
	if len(qb.Wheres) == 0 {
		return "", nil, ErrMissingWhereCondition
	}

	ds := goqu.Delete(goqu.T(qb.Table.Name)).WithDialect(qb.Dialect)

	// Apply WHERE conditions from builder
	expressions := make([]exp.Expression, len(qb.Wheres))
	for i, w := range qb.Wheres {
		expressions[i] = handleAny(w)
	}
	ds = ds.Where(expressions...)

	ds = ds.Prepared(true)

	return ds.ToSQL()
}

// UnmarshalJSON implements custom JSON unmarshaling for SQLBuilder.
func (qb *SQLBuilder) UnmarshalJSON(data []byte) error {
	type Alias SQLBuilder
	aux := &struct {
		Wheres []json.RawMessage `json:"wheres,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(qb),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Unmarshal Wheres with type detection
	if len(aux.Wheres) > 0 {
		qb.Wheres = make([]any, len(aux.Wheres))
		for i, raw := range aux.Wheres {
			condition, err := unmarshalCondition(raw)
			if err != nil {
				return fmt.Errorf("failed to unmarshal where condition at index %d: %w", i, err)
			}
			qb.Wheres[i] = condition
		}
	}

	return nil
}

// UnmarshalYAML implements custom YAML unmarshaling for SQLBuilder.
func (qb *SQLBuilder) UnmarshalYAML(unmarshal func(interface{}) error) error {
	aux := &struct {
		Dialect string                   `yaml:"dialect"`
		Fields  []Field                  `yaml:"fields,omitempty"`
		Table   Table                    `yaml:"table"`
		Wheres  []map[string]interface{} `yaml:"wheres,omitempty"`
		Sorts   []Sort                   `yaml:"sorts,omitempty"`
		GroupBy []Field                  `yaml:"groupBy,omitempty"`
	}{}

	if err := unmarshal(&aux); err != nil {
		return err
	}

	qb.Dialect = aux.Dialect
	qb.Fields = aux.Fields
	qb.Table = aux.Table
	qb.Sorts = aux.Sorts
	qb.GroupBy = aux.GroupBy

	// Unmarshal Wheres with type detection
	if len(aux.Wheres) > 0 {
		qb.Wheres = make([]any, len(aux.Wheres))
		for i, whereMap := range aux.Wheres {
			// Convert map to JSON and then unmarshal using our JSON logic
			jsonData, err := json.Marshal(whereMap)
			if err != nil {
				return fmt.Errorf("failed to marshal where to JSON: %w", err)
			}

			condition, err := unmarshalCondition(jsonData)
			if err != nil {
				return fmt.Errorf("failed to unmarshal where condition at index %d: %w", i, err)
			}
			qb.Wheres[i] = condition
		}
	}

	return nil
}
