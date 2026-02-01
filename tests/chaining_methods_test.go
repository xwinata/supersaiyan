package tests

import (
	"strings"
	"testing"

	"supersaiyan"

	"github.com/doug-martin/goqu/v9/exp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew tests the New constructor
func TestNew(t *testing.T) {
	t.Run("creates builder with default limit", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u")

		assert.Equal(t, "mysql", qb.Dialect)
		assert.Equal(t, "users", qb.Table.Name)
		assert.Equal(t, "u", qb.Table.Alias)

		// Default limit is 10
		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "LIMIT")
	})

	t.Run("creates builder with different dialects", func(t *testing.T) {
		dialects := []string{"mysql", "postgres", "sqlite3", "sqlserver"}

		for _, dialect := range dialects {
			qb := supersaiyan.New(dialect, "test_table", "t")
			assert.Equal(t, dialect, qb.Dialect)
		}
	})
}

// TestWithFields tests the WithFields chaining method
func TestWithFields(t *testing.T) {
	t.Run("adds single field", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(supersaiyan.Field{Name: "id", TableAlias: "u"}).
			Limit(0)

		assert.Len(t, qb.Fields, 1)
		assert.Equal(t, "id", qb.Fields[0].Name)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "SELECT")
		assert.Contains(t, sql, "id")
	})

	t.Run("adds multiple fields at once", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(
				supersaiyan.Field{Name: "id", TableAlias: "u"},
				supersaiyan.Field{Name: "username", TableAlias: "u"},
				supersaiyan.Field{Name: "email", TableAlias: "u"},
			).
			Limit(0)

		assert.Len(t, qb.Fields, 3)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "id")
		assert.Contains(t, sql, "username")
		assert.Contains(t, sql, "email")
	})

	t.Run("chains multiple WithFields calls", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(supersaiyan.Field{Name: "id", TableAlias: "u"}).
			WithFields(supersaiyan.Field{Name: "username", TableAlias: "u"}).
			WithFields(supersaiyan.Field{Name: "email", TableAlias: "u"}).
			Limit(0)

		assert.Len(t, qb.Fields, 3)
	})

	t.Run("adds field with alias", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(supersaiyan.Field{
				Name:       "created_at",
				TableAlias: "u",
				FieldAlias: "registration_date",
			}).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "created_at")
		assert.Contains(t, sql, "registration_date")
	})

	t.Run("adds field with expression", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "orders", "o").
			WithFields(supersaiyan.Field{
				FieldAlias: "total",
				Exp: supersaiyan.Literal{
					Value: "SUM(?)",
					Args:  []any{supersaiyan.Field{Name: "amount", TableAlias: "o"}},
				},
			}).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "SUM")
		assert.Contains(t, sql, "total")
	})
}

// TestWhere tests the Where chaining method
func TestWhere(t *testing.T) {
	t.Run("adds single condition", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.Eq("status", "u", "active")).
			Limit(0)

		assert.Len(t, qb.Wheres, 1)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "WHERE")
		assert.Contains(t, sql, "status")
		assert.Equal(t, "active", args[0])
	})

	t.Run("adds multiple conditions at once", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(
				supersaiyan.Eq("status", "u", "active"),
				supersaiyan.Gt("age", "u", 18),
			).
			Limit(0)

		assert.Len(t, qb.Wheres, 2)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "WHERE")
		assert.Len(t, args, 2)
	})

	t.Run("chains multiple Where calls", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.Eq("status", "u", "active")).
			Where(supersaiyan.Gt("age", "u", 18)).
			Where(supersaiyan.Like("email", "u", "%@example.com")).
			Limit(0)

		assert.Len(t, qb.Wheres, 3)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "WHERE")
		assert.Len(t, args, 3)
	})

	t.Run("adds OR condition group", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.Or(
				supersaiyan.Eq("role", "u", "admin"),
				supersaiyan.Eq("role", "u", "moderator"),
			)).
			Limit(0)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "OR")
		assert.Len(t, args, 2)
	})

	t.Run("adds BETWEEN condition", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "products", "p").
			Where(supersaiyan.Between("price", "p", 100, 1000)).
			Limit(0)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "BETWEEN")
		assert.Len(t, args, 2)
	})

	t.Run("adds IN condition", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.In("status", "u", []string{"active", "pending", "verified"})).
			Limit(0)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "IN")
		assert.NotEmpty(t, args)
	})

	t.Run("adds IS NULL condition", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.IsNull("deleted_at", "u")).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "IS NULL")
	})
}

// TestOrderBy tests the OrderBy chaining method
func TestOrderBy(t *testing.T) {
	t.Run("adds single sort ascending", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			OrderBy(supersaiyan.Asc("username", "u")).
			Limit(0)

		assert.Len(t, qb.Sorts, 1)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "ORDER BY")
		assert.Contains(t, sql, "username")
		assert.Contains(t, sql, "ASC")
	})

	t.Run("adds single sort descending", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			OrderBy(supersaiyan.Desc("created_at", "u")).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "ORDER BY")
		assert.Contains(t, sql, "created_at")
		assert.Contains(t, sql, "DESC")
	})

	t.Run("chains multiple OrderBy calls", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			OrderBy(supersaiyan.Desc("created_at", "u")).
			OrderBy(supersaiyan.Asc("username", "u")).
			Limit(0)

		assert.Len(t, qb.Sorts, 2)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "ORDER BY")
	})
}

// TestGroupByFields tests the GroupByFields chaining method
func TestGroupByFields(t *testing.T) {
	t.Run("adds single group by field", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "orders", "o").
			WithFields(
				supersaiyan.Field{Name: "user_id", TableAlias: "o"},
				supersaiyan.Field{
					FieldAlias: "total",
					Exp: supersaiyan.Literal{
						Value: "SUM(?)",
						Args:  []any{supersaiyan.Field{Name: "amount", TableAlias: "o"}},
					},
				},
			).
			GroupByFields(supersaiyan.Field{Name: "user_id", TableAlias: "o"}).
			Limit(0)

		assert.Len(t, qb.GroupBy, 1)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "GROUP BY")
		assert.Contains(t, sql, "user_id")
	})

	t.Run("chains multiple GroupByFields calls", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "orders", "o").
			WithFields(
				supersaiyan.Field{Name: "user_id", TableAlias: "o"},
				supersaiyan.Field{Name: "status", TableAlias: "o"},
			).
			GroupByFields(supersaiyan.Field{Name: "user_id", TableAlias: "o"}).
			GroupByFields(supersaiyan.Field{Name: "status", TableAlias: "o"}).
			Limit(0)

		assert.Len(t, qb.GroupBy, 2)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "GROUP BY")
	})
}

// TestJoin tests the Join chaining methods
func TestJoin(t *testing.T) {
	t.Run("adds inner join", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			InnerJoin("orders", "o", supersaiyan.Eq("user_id", "o", supersaiyan.Field{Name: "id", TableAlias: "u"})).
			Limit(0)

		assert.Len(t, qb.Table.Relations, 1)
		assert.Equal(t, exp.InnerJoinType, qb.Table.Relations[0].JoinType)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "INNER JOIN")
		assert.Contains(t, sql, "orders")
	})

	t.Run("adds left join", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			LeftJoin("profiles", "p", supersaiyan.Eq("user_id", "p", supersaiyan.Field{Name: "id", TableAlias: "u"})).
			Limit(0)

		assert.Len(t, qb.Table.Relations, 1)
		assert.Equal(t, exp.LeftJoinType, qb.Table.Relations[0].JoinType)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "LEFT JOIN")
		assert.Contains(t, sql, "profiles")
	})

	t.Run("adds right join", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			RightJoin("departments", "d", supersaiyan.Eq("id", "d", supersaiyan.Field{Name: "department_id", TableAlias: "u"})).
			Limit(0)

		assert.Len(t, qb.Table.Relations, 1)
		assert.Equal(t, exp.RightJoinType, qb.Table.Relations[0].JoinType)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "RIGHT JOIN")
		assert.Contains(t, sql, "departments")
	})

	t.Run("chains multiple joins", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			InnerJoin("orders", "o", supersaiyan.Eq("user_id", "o", supersaiyan.Field{Name: "id", TableAlias: "u"})).
			LeftJoin("profiles", "p", supersaiyan.Eq("user_id", "p", supersaiyan.Field{Name: "id", TableAlias: "u"})).
			Limit(0)

		assert.Len(t, qb.Table.Relations, 2)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "INNER JOIN")
		assert.Contains(t, sql, "LEFT JOIN")
	})
}

// TestLimit tests the Limit chaining method
func TestLimit(t *testing.T) {
	t.Run("sets limit", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Limit(25)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "LIMIT")
		// With prepared statements, limit value is in args (as int64)
		found := false
		for _, arg := range args {
			if v, ok := arg.(int64); ok && v == 25 {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected limit value 25 in args")
	})

	t.Run("removes limit with zero", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.NotContains(t, sql, "LIMIT")
	})

	t.Run("overrides previous limit", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Limit(10).
			Limit(25)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "LIMIT")
		// Check for 25, not 10
		found25 := false
		found10 := false
		for _, arg := range args {
			if v, ok := arg.(int64); ok {
				if v == 25 {
					found25 = true
				}
				if v == 10 {
					found10 = true
				}
			}
		}
		assert.True(t, found25, "Expected limit value 25 in args")
		assert.False(t, found10, "Should not have limit value 10 in args")
	})
}

// TestOffset tests the Offset chaining method
func TestOffset(t *testing.T) {
	t.Run("sets offset", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Offset(50)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "OFFSET")
		// With prepared statements, offset value is in args (as int64)
		found := false
		for _, arg := range args {
			if v, ok := arg.(int64); ok && v == 50 {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected offset value 50 in args")
	})

	t.Run("chains with limit", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Limit(25).
			Offset(50)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "LIMIT")
		assert.Contains(t, sql, "OFFSET")
		// Check for both values (as int64)
		found25 := false
		found50 := false
		for _, arg := range args {
			if v, ok := arg.(int64); ok {
				if v == 25 {
					found25 = true
				}
				if v == 50 {
					found50 = true
				}
			}
		}
		assert.True(t, found25, "Expected limit value 25 in args")
		assert.True(t, found50, "Expected offset value 50 in args")
	})
}

// TestComplexChaining tests complex chaining scenarios
func TestComplexChaining(t *testing.T) {
	t.Run("full query with all chaining methods", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(
				supersaiyan.Field{Name: "id", TableAlias: "u"},
				supersaiyan.Field{Name: "username", TableAlias: "u"},
				supersaiyan.Field{Name: "email", TableAlias: "u"},
			).
			InnerJoin("orders", "o", supersaiyan.Eq("user_id", "o", supersaiyan.Field{Name: "id", TableAlias: "u"})).
			LeftJoin("profiles", "p", supersaiyan.Eq("user_id", "p", supersaiyan.Field{Name: "id", TableAlias: "u"})).
			Where(
				supersaiyan.Eq("status", "u", "active"),
				supersaiyan.Gt("age", "u", 18),
			).
			Where(supersaiyan.Or(
				supersaiyan.Like("email", "u", "%@gmail.com"),
				supersaiyan.Like("email", "u", "%@yahoo.com"),
			)).
			OrderBy(
				supersaiyan.Desc("created_at", "u"),
				supersaiyan.Asc("username", "u"),
			).
			Limit(25).
			Offset(50)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.NotEmpty(t, sql)
		assert.NotEmpty(t, args)

		// Verify all parts are present
		assert.Contains(t, sql, "SELECT")
		assert.Contains(t, sql, "FROM")
		assert.Contains(t, sql, "users")
		assert.Contains(t, sql, "INNER JOIN")
		assert.Contains(t, sql, "LEFT JOIN")
		assert.Contains(t, sql, "WHERE")
		assert.Contains(t, sql, "ORDER BY")
		assert.Contains(t, sql, "LIMIT")
		assert.Contains(t, sql, "OFFSET")
	})

	t.Run("query with aggregation and grouping", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "orders", "o").
			WithFields(
				supersaiyan.Field{Name: "user_id", TableAlias: "o"},
				supersaiyan.Field{
					FieldAlias: "order_count",
					Exp: supersaiyan.Literal{
						Value: "COUNT(?)",
						Args:  []any{supersaiyan.Field{Name: "id", TableAlias: "o"}},
					},
				},
				supersaiyan.Field{
					FieldAlias: "total_amount",
					Exp: supersaiyan.Literal{
						Value: "SUM(?)",
						Args:  []any{supersaiyan.Field{Name: "amount", TableAlias: "o"}},
					},
				},
			).
			Where(supersaiyan.Eq("status", "o", "completed")).
			GroupByFields(supersaiyan.Field{Name: "user_id", TableAlias: "o"}).
			OrderBy(supersaiyan.Desc("total_amount", "")).
			Limit(10)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.NotEmpty(t, sql)
		assert.Contains(t, sql, "COUNT")
		assert.Contains(t, sql, "SUM")
		assert.Contains(t, sql, "GROUP BY")
		// Check that completed is in args
		found := false
		for _, arg := range args {
			if arg == "completed" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected 'completed' in args")
	})

	t.Run("query with case expression", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(
				supersaiyan.Field{Name: "id", TableAlias: "u"},
				supersaiyan.Field{Name: "username", TableAlias: "u"},
				supersaiyan.Field{
					FieldAlias: "status_label",
					Exp: supersaiyan.Case{
						Conditions: []supersaiyan.WhenThen{
							{
								When: supersaiyan.Eq("status", "u", "active"),
								Then: "Active User",
							},
							{
								When: supersaiyan.Eq("status", "u", "inactive"),
								Then: "Inactive User",
							},
						},
						Else: "Unknown",
					},
				},
			).
			Where(supersaiyan.IsNotNull("email", "u")).
			OrderBy(supersaiyan.Asc("username", "u")).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "CASE")
		assert.Contains(t, sql, "WHEN")
		assert.Contains(t, sql, "THEN")
		assert.Contains(t, sql, "ELSE")
	})
}

// TestMethodChaining tests that methods return the builder for chaining
func TestMethodChaining(t *testing.T) {
	t.Run("all methods return builder", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u")

		// Test that each method returns *SQLBuilder
		qb2 := qb.WithFields(supersaiyan.Field{Name: "id", TableAlias: "u"})
		assert.Equal(t, qb, qb2)

		qb3 := qb.Where(supersaiyan.Eq("status", "u", "active"))
		assert.Equal(t, qb, qb3)

		qb4 := qb.OrderBy(supersaiyan.Asc("id", "u"))
		assert.Equal(t, qb, qb4)

		qb5 := qb.GroupByFields(supersaiyan.Field{Name: "id", TableAlias: "u"})
		assert.Equal(t, qb, qb5)

		qb6 := qb.InnerJoin("orders", "o", supersaiyan.Eq("user_id", "o", supersaiyan.Field{Name: "id", TableAlias: "u"}))
		assert.Equal(t, qb, qb6)

		qb7 := qb.Limit(10)
		assert.Equal(t, qb, qb7)

		qb8 := qb.Offset(20)
		assert.Equal(t, qb, qb8)
	})

	t.Run("can chain in any order", func(t *testing.T) {
		// Order 1
		qb1 := supersaiyan.New("mysql", "users", "u").
			Limit(10).
			Where(supersaiyan.Eq("status", "u", "active")).
			WithFields(supersaiyan.Field{Name: "id", TableAlias: "u"}).
			OrderBy(supersaiyan.Asc("id", "u"))

		sql1, _, err1 := qb1.Select()
		require.NoError(t, err1)

		// Order 2 (different order, same result)
		qb2 := supersaiyan.New("mysql", "users", "u").
			WithFields(supersaiyan.Field{Name: "id", TableAlias: "u"}).
			OrderBy(supersaiyan.Asc("id", "u")).
			Where(supersaiyan.Eq("status", "u", "active")).
			Limit(10)

		sql2, _, err2 := qb2.Select()
		require.NoError(t, err2)

		// Both should produce the same SQL
		assert.Equal(t, sql1, sql2)
	})
}

// TestBuilderImmutability tests that the builder doesn't mutate unexpectedly
func TestBuilderImmutability(t *testing.T) {
	t.Run("chaining doesn't create new instances", func(t *testing.T) {
		qb1 := supersaiyan.New("mysql", "users", "u")
		qb2 := qb1.WithFields(supersaiyan.Field{Name: "id", TableAlias: "u"})

		// Should be the same instance
		assert.Equal(t, qb1, qb2)

		// Changes should be reflected in both
		assert.Len(t, qb1.Fields, 1)
		assert.Len(t, qb2.Fields, 1)
	})

	t.Run("multiple chains affect same builder", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u")

		qb.WithFields(supersaiyan.Field{Name: "id", TableAlias: "u"})
		qb.Where(supersaiyan.Eq("status", "u", "active"))
		qb.OrderBy(supersaiyan.Asc("id", "u"))

		assert.Len(t, qb.Fields, 1)
		assert.Len(t, qb.Wheres, 1)
		assert.Len(t, qb.Sorts, 1)
	})
}

// TestEdgeCases tests edge cases in chaining
func TestEdgeCases(t *testing.T) {
	t.Run("empty table name", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "", "")
		sql, _, err := qb.Select()
		// Empty table name should cause an error
		assert.Error(t, err)
		_ = sql
	})

	t.Run("no fields specified", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").Limit(0)
		sql, _, err := qb.Select()
		require.NoError(t, err)
		// Should select all fields
		assert.Contains(t, strings.ToUpper(sql), "SELECT")
	})

	t.Run("no where conditions", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").Limit(0)
		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.NotContains(t, sql, "WHERE")
	})

	t.Run("no order by", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").Limit(0)
		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.NotContains(t, sql, "ORDER BY")
	})

	t.Run("no group by", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").Limit(0)
		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.NotContains(t, sql, "GROUP BY")
	})
}

// TestFieldHelpers tests the F() and Exp() helper functions
func TestFieldHelpers(t *testing.T) {
	t.Run("F creates simple field without alias", func(t *testing.T) {
		field := supersaiyan.F("username", supersaiyan.WithTable("u"))

		assert.Equal(t, "username", field.Name)
		assert.Equal(t, "u", field.TableAlias)
		assert.Empty(t, field.FieldAlias)
		assert.Nil(t, field.Exp)
	})

	t.Run("F creates field with alias", func(t *testing.T) {
		field := supersaiyan.F("created_at", supersaiyan.WithTable("u"), supersaiyan.WithAlias("registration_date"))

		assert.Equal(t, "created_at", field.Name)
		assert.Equal(t, "u", field.TableAlias)
		assert.Equal(t, "registration_date", field.FieldAlias)
		assert.Nil(t, field.Exp)
	})

	t.Run("F ignores empty alias", func(t *testing.T) {
		field := supersaiyan.F("username", supersaiyan.WithTable("u"), supersaiyan.WithAlias(""))

		assert.Equal(t, "username", field.Name)
		assert.Equal(t, "u", field.TableAlias)
		assert.Empty(t, field.FieldAlias)
		assert.Nil(t, field.Exp)
	})

	t.Run("F creates field without table alias", func(t *testing.T) {
		field := supersaiyan.F("username")

		assert.Equal(t, "username", field.Name)
		assert.Empty(t, field.TableAlias)
		assert.Empty(t, field.FieldAlias)
		assert.Nil(t, field.Exp)
	})

	t.Run("F creates field with only field alias (no table alias)", func(t *testing.T) {
		field := supersaiyan.F("username", supersaiyan.WithAlias("user_name"))

		assert.Equal(t, "username", field.Name)
		assert.Empty(t, field.TableAlias)
		assert.Equal(t, "user_name", field.FieldAlias)
		assert.Nil(t, field.Exp)
	})

	t.Run("Exp creates expression field with COUNT", func(t *testing.T) {
		field := supersaiyan.Exp("order_count", supersaiyan.Literal{
			Value: "COUNT(?)",
			Args:  []any{supersaiyan.F("id", supersaiyan.WithTable("o"))},
		})

		assert.Empty(t, field.Name)
		assert.Empty(t, field.TableAlias)
		assert.Equal(t, "order_count", field.FieldAlias)
		assert.NotNil(t, field.Exp)
	})

	t.Run("Exp creates expression field with SUM", func(t *testing.T) {
		field := supersaiyan.Exp("total_amount", supersaiyan.Literal{
			Value: "SUM(?)",
			Args:  []any{supersaiyan.F("amount", supersaiyan.WithTable("o"))},
		})

		assert.Equal(t, "total_amount", field.FieldAlias)
		assert.NotNil(t, field.Exp)
	})

	t.Run("Exp creates expression field with CASE", func(t *testing.T) {
		caseExpr := supersaiyan.Case{
			Conditions: []supersaiyan.WhenThen{
				{
					When: supersaiyan.Eq("status", "u", "active"),
					Then: "Active",
				},
			},
			Else: "Unknown",
		}

		field := supersaiyan.Exp("status_label", caseExpr)

		assert.Equal(t, "status_label", field.FieldAlias)
		assert.NotNil(t, field.Exp)
	})

	t.Run("Exp creates expression field with COALESCE", func(t *testing.T) {
		coalesceExpr := supersaiyan.Coalesce{
			Fields: []supersaiyan.Field{
				supersaiyan.F("nickname", supersaiyan.WithTable("u")),
				supersaiyan.F("username", supersaiyan.WithTable("u")),
			},
			DefaultValue: "Anonymous",
		}

		field := supersaiyan.Exp("display_name", coalesceExpr)

		assert.Equal(t, "display_name", field.FieldAlias)
		assert.NotNil(t, field.Exp)
	})
}

// TestFieldHelpersInQueries tests F() and Exp() in actual queries
func TestFieldHelpersInQueries(t *testing.T) {
	t.Run("query with aliased field using F()", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(
				supersaiyan.F("id", supersaiyan.WithTable("u")),
				supersaiyan.F("created_at", supersaiyan.WithTable("u"), supersaiyan.WithAlias("reg_date")),
			).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "created_at")
		assert.Contains(t, sql, "reg_date")
	})

	t.Run("query with expression field using Exp()", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "orders", "o").
			WithFields(
				supersaiyan.F("user_id", supersaiyan.WithTable("o")),
				supersaiyan.Exp("order_count", supersaiyan.Literal{
					Value: "COUNT(?)",
					Args:  []any{supersaiyan.F("id", supersaiyan.WithTable("o"))},
				}),
			).
			GroupByFields(supersaiyan.F("user_id", supersaiyan.WithTable("o"))).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "COUNT")
		assert.Contains(t, sql, "order_count")
		assert.Contains(t, sql, "GROUP BY")
	})

	t.Run("query with CASE expression using Exp()", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(
				supersaiyan.F("id", supersaiyan.WithTable("u")),
				supersaiyan.F("username", supersaiyan.WithTable("u")),
				supersaiyan.Exp("status_label", supersaiyan.Case{
					Conditions: []supersaiyan.WhenThen{
						{
							When: supersaiyan.Eq("status", "u", "active"),
							Then: "Active User",
						},
						{
							When: supersaiyan.Eq("status", "u", "inactive"),
							Then: "Inactive User",
						},
					},
					Else: "Unknown",
				}),
			).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "CASE")
		assert.Contains(t, sql, "WHEN")
		assert.Contains(t, sql, "status_label")
	})

	t.Run("query with COALESCE using Exp()", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(
				supersaiyan.F("id", supersaiyan.WithTable("u")),
				supersaiyan.Exp("display_name", supersaiyan.Coalesce{
					Fields: []supersaiyan.Field{
						supersaiyan.F("nickname", supersaiyan.WithTable("u")),
						supersaiyan.F("username", supersaiyan.WithTable("u")),
						supersaiyan.F("email", supersaiyan.WithTable("u")),
					},
					DefaultValue: "Anonymous",
				}),
			).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "COALESCE")
		assert.Contains(t, sql, "display_name")
	})

	t.Run("complex query with multiple aggregations", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "orders", "o").
			WithFields(
				supersaiyan.F("user_id", supersaiyan.WithTable("o")),
				supersaiyan.F("status", supersaiyan.WithTable("o")),
				supersaiyan.Exp("order_count", supersaiyan.Literal{
					Value: "COUNT(?)",
					Args:  []any{supersaiyan.F("id", supersaiyan.WithTable("o"))},
				}),
				supersaiyan.Exp("total_amount", supersaiyan.Literal{
					Value: "SUM(?)",
					Args:  []any{supersaiyan.F("amount", supersaiyan.WithTable("o"))},
				}),
				supersaiyan.Exp("avg_amount", supersaiyan.Literal{
					Value: "AVG(?)",
					Args:  []any{supersaiyan.F("amount", supersaiyan.WithTable("o"))},
				}),
			).
			Where(supersaiyan.Eq("status", "o", "completed")).
			GroupByFields(
				supersaiyan.F("user_id", supersaiyan.WithTable("o")),
				supersaiyan.F("status", supersaiyan.WithTable("o")),
			).
			OrderBy(supersaiyan.Desc("total_amount", "")).
			Limit(10)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "COUNT")
		assert.Contains(t, sql, "SUM")
		assert.Contains(t, sql, "AVG")
		assert.Contains(t, sql, "order_count")
		assert.Contains(t, sql, "total_amount")
		assert.Contains(t, sql, "avg_amount")
		assert.Contains(t, sql, "GROUP BY")
		assert.NotEmpty(t, args)
	})

	t.Run("multiple fields with and without aliases", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(
				supersaiyan.F("id", supersaiyan.WithTable("u")),
				supersaiyan.F("username", supersaiyan.WithTable("u")),
				supersaiyan.F("created_at", supersaiyan.WithTable("u"), supersaiyan.WithAlias("reg_date")),
				supersaiyan.F("updated_at", supersaiyan.WithTable("u"), supersaiyan.WithAlias("mod_date")),
			).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "id")
		assert.Contains(t, sql, "username")
		assert.Contains(t, sql, "reg_date")
		assert.Contains(t, sql, "mod_date")
	})
}

// TestFieldStructBackwardCompatibility tests that Field struct still works
func TestFieldStructBackwardCompatibility(t *testing.T) {
	t.Run("Field struct with name and table alias", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(
				supersaiyan.Field{
					Name:       "id",
					TableAlias: "u",
				},
				supersaiyan.Field{
					Name:       "username",
					TableAlias: "u",
				},
			).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "id")
		assert.Contains(t, sql, "username")
	})

	t.Run("Field struct with expression", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "orders", "o").
			WithFields(
				supersaiyan.Field{
					Name:       "user_id",
					TableAlias: "o",
				},
				supersaiyan.Field{
					FieldAlias: "total",
					Exp: supersaiyan.Literal{
						Value: "SUM(?)",
						Args:  []any{supersaiyan.F("amount", supersaiyan.WithTable("o"))},
					},
				},
			).
			GroupByFields(supersaiyan.F("user_id", supersaiyan.WithTable("o"))).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "SUM")
		assert.Contains(t, sql, "total")
	})

	t.Run("Field struct with alias", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(
				supersaiyan.Field{
					Name:       "created_at",
					TableAlias: "u",
					FieldAlias: "registration_date",
				},
			).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "created_at")
		assert.Contains(t, sql, "registration_date")
	})
}
