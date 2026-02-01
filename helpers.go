package supersaiyan

import (
	"reflect"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

// handleAnyOptions contains options for converting arbitrary values to goqu expressions.
type handleAnyOptions struct {
	alias string
}

// handleAnyOption is a function that modifies handleAnyOptions.
type handleAnyOption func(*handleAnyOptions)

// withAlias returns an option that sets the alias for the expression.
func withAlias(alias string) handleAnyOption {
	return func(opts *handleAnyOptions) {
		opts.alias = alias
	}
}

// handleAny recursively converts arbitrary values to goqu expressions.
// It supports SQLBuilder, Field, BoolOp, WhereGroup, RangeOp, Literal, Case, Coalesce,
// goqu.Expression, slices, and primitive values.
func handleAny(a any, opts ...handleAnyOption) exp.Expression {
	// Handle nil values explicitly
	if a == nil {
		return goqu.L("NULL")
	}

	options := handleAnyOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	// Handle pointers by dereferencing
	r := reflect.ValueOf(a)
	if r.Kind() == reflect.Ptr {
		if r.IsNil() {
			return goqu.L("NULL")
		}
		a = r.Elem().Interface()
		r = reflect.ValueOf(a)
	}

	// Handle SQLBuilder (subquery)
	if qb, ok := a.(SQLBuilder); ok {
		return qb.mainSelect()
	}

	// Handle Field
	if f, ok := a.(Field); ok {
		return f.expression()
	}

	// Handle BoolOp
	if bo, ok := a.(BoolOp); ok {
		return bo.expression()
	}

	// Handle WhereGroup
	if wg, ok := a.(WhereGroup); ok {
		return wg.expression()
	}

	// Handle RangeOp
	if ro, ok := a.(RangeOp); ok {
		return ro.expression()
	}

	// Handle Literal
	if l, ok := a.(Literal); ok {
		if options.alias != "" {
			return l.expression().As(options.alias)
		}
		return l.expression()
	}

	// Handle Case
	if c, ok := a.(Case); ok {
		if options.alias != "" {
			return c.expression().As(options.alias)
		}
		return c.expression()
	}

	// Handle Coalesce
	if co, ok := a.(Coalesce); ok {
		if options.alias != "" {
			return co.expression().As(options.alias)
		}
		return co.expression()
	}

	// Handle goqu.Expression directly
	if goquExpr, ok := a.(exp.Expression); ok {
		return goquExpr
	}

	// Handle slices
	if r.Kind() == reflect.Slice {
		l := r.Len()
		if l == 0 {
			// Return empty slice for IN clauses - goqu handles this correctly
			return goqu.L("(?)", goqu.V([]any{}))
		}

		// For slices, just return the slice itself - goqu handles it
		return goqu.V(a)
	}

	// Default: treat as a literal value
	return goqu.V(a)
}
