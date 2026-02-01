package tests

import (
	"testing"

	"supersaiyan"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSelect tests the Select method
func TestSelect(t *testing.T) {
	t.Run("generates simple select", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(supersaiyan.F("id", "u")).
			Limit(0)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.NotEmpty(t, sql)
		assert.NotNil(t, args)
		assert.Contains(t, sql, "SELECT")
		assert.Contains(t, sql, "FROM")
	})

	t.Run("generates select with where", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(supersaiyan.F("id", "u")).
			Where(supersaiyan.Eq("status", "u", "active")).
			Limit(0)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "WHERE")
		assert.Len(t, args, 1)
		assert.Equal(t, "active", args[0])
	})

	t.Run("generates select with joins", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(
				supersaiyan.F("id", "u"),
				supersaiyan.F("order_id", "o"),
			).
			InnerJoin("orders", "o", supersaiyan.Eq("user_id", "o", supersaiyan.F("id", "u"))).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "INNER JOIN")
	})

	t.Run("generates select with order by", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(supersaiyan.F("id", "u")).
			OrderBy(supersaiyan.Desc("created_at", "u")).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "ORDER BY")
	})

	t.Run("generates select with limit and offset", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(supersaiyan.F("id", "u")).
			Limit(25).
			Offset(50)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		// With prepared statements, values are in args, not in SQL
		assert.Contains(t, sql, "LIMIT")
		assert.Contains(t, sql, "OFFSET")
		assert.NotEmpty(t, args)
	})

	t.Run("generates select with group by", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "orders", "o").
			WithFields(
				supersaiyan.F("user_id", "o"),
				supersaiyan.Field{
					FieldAlias: "count",
					Exp: supersaiyan.Literal{
						Value: "COUNT(?)",
						Args:  []any{supersaiyan.F("id", "o")},
					},
				},
			).
			GroupByFields(supersaiyan.F("user_id", "o")).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "GROUP BY")
	})

	t.Run("uses prepared statements", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			WithFields(supersaiyan.F("id", "u")).
			Where(supersaiyan.Eq("status", "u", "active")).
			Limit(0)

		sql, args, err := qb.Select()
		require.NoError(t, err)
		// Prepared statements use ? placeholders
		assert.Contains(t, sql, "?")
		assert.NotEmpty(t, args)
	})

	t.Run("generates select without fields (SELECT *)", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "SELECT")
		assert.Contains(t, sql, "FROM")
	})

	t.Run("generates select with different dialects", func(t *testing.T) {
		dialects := []string{"mysql", "postgres", "sqlite3"}

		for _, dialect := range dialects {
			qb := supersaiyan.New(dialect, "users", "u").
				WithFields(supersaiyan.F("id", "u")).
				Limit(0)

			sql, _, err := qb.Select()
			require.NoError(t, err)
			assert.NotEmpty(t, sql)
		}
	})
}

// TestCount tests the Count method
func TestCount(t *testing.T) {
	t.Run("generates count query", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Limit(0)

		sql, _, err := qb.Count()
		require.NoError(t, err)
		assert.Contains(t, sql, "COUNT")
		assert.Contains(t, sql, "*")
	})

	t.Run("generates count with where", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.Eq("status", "u", "active")).
			Limit(0)

		sql, args, err := qb.Count()
		require.NoError(t, err)
		assert.Contains(t, sql, "COUNT")
		assert.Contains(t, sql, "WHERE")
		assert.Len(t, args, 1)
	})

	t.Run("generates count with joins", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			InnerJoin("orders", "o", supersaiyan.Eq("user_id", "o", supersaiyan.F("id", "u"))).
			Where(supersaiyan.Eq("status", "o", "completed")).
			Limit(0)

		sql, _, err := qb.Count()
		require.NoError(t, err)
		assert.Contains(t, sql, "COUNT")
		assert.Contains(t, sql, "INNER JOIN")
	})

	t.Run("count respects limit and offset", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Limit(10).
			Offset(20)

		sql, args, err := qb.Count()
		require.NoError(t, err)
		// With prepared statements, values are in args
		assert.Contains(t, sql, "LIMIT")
		assert.Contains(t, sql, "OFFSET")
		assert.NotEmpty(t, args)
	})
}

// TestAdd tests the Add method (INSERT)
func TestAdd(t *testing.T) {
	t.Run("generates insert with single field", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u")

		entry := map[string]any{
			"username": "john_doe",
		}

		sql, args, err := qb.Add(entry)
		require.NoError(t, err)
		assert.Contains(t, sql, "INSERT")
		assert.Contains(t, sql, "users")
		assert.Len(t, args, 1)
		assert.Equal(t, "john_doe", args[0])
	})

	t.Run("generates insert with multiple fields", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u")

		entry := map[string]any{
			"username": "john_doe",
			"email":    "john@example.com",
			"age":      30,
			"status":   "active",
		}

		sql, args, err := qb.Add(entry)
		require.NoError(t, err)
		assert.Contains(t, sql, "INSERT")
		assert.Len(t, args, 4)
	})

	t.Run("uses prepared statements", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u")

		entry := map[string]any{
			"username": "john_doe",
			"email":    "john@example.com",
		}

		sql, args, err := qb.Add(entry)
		require.NoError(t, err)
		assert.Contains(t, sql, "?")
		assert.NotEmpty(t, args)
	})

	t.Run("works with different dialects", func(t *testing.T) {
		dialects := []string{"mysql", "postgres", "sqlite3"}

		for _, dialect := range dialects {
			qb := supersaiyan.New(dialect, "users", "u")

			entry := map[string]any{
				"username": "john_doe",
			}

			sql, _, err := qb.Add(entry)
			require.NoError(t, err)
			assert.Contains(t, sql, "INSERT")
		}
	})

	t.Run("handles nil values", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u")

		entry := map[string]any{
			"username":   "john_doe",
			"deleted_at": nil,
		}

		sql, args, err := qb.Add(entry)
		require.NoError(t, err)
		assert.NotEmpty(t, sql)
		assert.Len(t, args, 2)
	})

	t.Run("handles empty map", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u")

		entry := map[string]any{}

		sql, args, err := qb.Add(entry)
		require.NoError(t, err)
		assert.NotEmpty(t, sql)
		assert.Empty(t, args)
	})
}

// TestEdit tests the Edit method (UPDATE)
func TestEdit(t *testing.T) {
	t.Run("generates update with where condition", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.Eq("id", "u", 1))

		entry := map[string]any{
			"username": "jane_doe",
		}

		sql, args, err := qb.Edit(entry)
		require.NoError(t, err)
		assert.Contains(t, sql, "UPDATE")
		assert.Contains(t, sql, "users")
		assert.Contains(t, sql, "WHERE")
		assert.Len(t, args, 2) // 1 for SET, 1 for WHERE
	})

	t.Run("generates update with multiple fields", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.Eq("id", "u", 1))

		entry := map[string]any{
			"username": "jane_doe",
			"email":    "jane@example.com",
			"age":      25,
		}

		sql, args, err := qb.Edit(entry)
		require.NoError(t, err)
		assert.Contains(t, sql, "UPDATE")
		assert.Contains(t, sql, "SET")
		assert.Len(t, args, 4) // 3 for SET, 1 for WHERE
	})

	t.Run("requires where condition", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u")

		entry := map[string]any{
			"username": "jane_doe",
		}

		sql, args, err := qb.Edit(entry)
		require.Error(t, err)
		assert.Equal(t, supersaiyan.ErrMissingWhereCondition, err)
		assert.Empty(t, sql)
		assert.Nil(t, args)
	})

	t.Run("works with multiple where conditions", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(
				supersaiyan.Eq("status", "u", "active"),
				supersaiyan.Gt("age", "u", 18),
			)

		entry := map[string]any{
			"verified": true,
		}

		sql, args, err := qb.Edit(entry)
		require.NoError(t, err)
		assert.Contains(t, sql, "UPDATE")
		assert.Contains(t, sql, "WHERE")
		assert.Len(t, args, 3) // 1 for SET, 2 for WHERE
	})

	t.Run("uses prepared statements", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.Eq("id", "u", 1))

		entry := map[string]any{
			"username": "jane_doe",
		}

		sql, args, err := qb.Edit(entry)
		require.NoError(t, err)
		assert.Contains(t, sql, "?")
		assert.NotEmpty(t, args)
	})

	t.Run("works with different dialects", func(t *testing.T) {
		dialects := []string{"mysql", "postgres", "sqlite3"}

		for _, dialect := range dialects {
			qb := supersaiyan.New(dialect, "users", "u").
				Where(supersaiyan.Eq("id", "u", 1))

			entry := map[string]any{
				"username": "jane_doe",
			}

			sql, _, err := qb.Edit(entry)
			require.NoError(t, err)
			assert.Contains(t, sql, "UPDATE")
		}
	})

	t.Run("handles nil values", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.Eq("id", "u", 1))

		entry := map[string]any{
			"deleted_at": nil,
		}

		sql, args, err := qb.Edit(entry)
		require.NoError(t, err)
		assert.NotEmpty(t, sql)
		assert.Len(t, args, 2)
	})

	t.Run("works with OR conditions", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.Or(
				supersaiyan.Eq("status", "u", "inactive"),
				supersaiyan.Eq("status", "u", "pending"),
			))

		entry := map[string]any{
			"status": "active",
		}

		sql, args, err := qb.Edit(entry)
		require.NoError(t, err)
		assert.Contains(t, sql, "UPDATE")
		assert.Contains(t, sql, "OR")
		assert.Len(t, args, 3)
	})
}

// TestDelete tests the Delete method
func TestDelete(t *testing.T) {
	t.Run("generates delete with where condition", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.Eq("id", "u", 1))

		sql, args, err := qb.Delete()
		require.NoError(t, err)
		assert.Contains(t, sql, "DELETE")
		assert.Contains(t, sql, "FROM")
		assert.Contains(t, sql, "users")
		assert.Contains(t, sql, "WHERE")
		assert.Len(t, args, 1)
	})

	t.Run("requires where condition", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u")

		sql, args, err := qb.Delete()
		require.Error(t, err)
		assert.Equal(t, supersaiyan.ErrMissingWhereCondition, err)
		assert.Empty(t, sql)
		assert.Nil(t, args)
	})

	t.Run("works with multiple where conditions", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(
				supersaiyan.Eq("status", "u", "inactive"),
				supersaiyan.Lt("last_login", "u", "2020-01-01"),
			)

		sql, args, err := qb.Delete()
		require.NoError(t, err)
		assert.Contains(t, sql, "DELETE")
		assert.Contains(t, sql, "WHERE")
		assert.Len(t, args, 2)
	})

	t.Run("uses prepared statements", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.Eq("id", "u", 1))

		sql, args, err := qb.Delete()
		require.NoError(t, err)
		assert.Contains(t, sql, "?")
		assert.NotEmpty(t, args)
	})

	t.Run("works with different dialects", func(t *testing.T) {
		dialects := []string{"mysql", "postgres", "sqlite3"}

		for _, dialect := range dialects {
			qb := supersaiyan.New(dialect, "users", "u").
				Where(supersaiyan.Eq("id", "u", 1))

			sql, _, err := qb.Delete()
			require.NoError(t, err)
			assert.Contains(t, sql, "DELETE")
		}
	})

	t.Run("works with OR conditions", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.Or(
				supersaiyan.Eq("status", "u", "spam"),
				supersaiyan.Eq("status", "u", "banned"),
			))

		sql, args, err := qb.Delete()
		require.NoError(t, err)
		assert.Contains(t, sql, "DELETE")
		assert.Contains(t, sql, "OR")
		assert.Len(t, args, 2)
	})

	t.Run("works with BETWEEN condition", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "logs", "l").
			Where(supersaiyan.Between("created_at", "l", "2020-01-01", "2020-12-31"))

		sql, args, err := qb.Delete()
		require.NoError(t, err)
		assert.Contains(t, sql, "DELETE")
		assert.Contains(t, sql, "BETWEEN")
		assert.Len(t, args, 2)
	})

	t.Run("works with IN condition", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.In("id", "u", []int{1, 2, 3, 4, 5}))

		sql, args, err := qb.Delete()
		require.NoError(t, err)
		assert.Contains(t, sql, "DELETE")
		assert.Contains(t, sql, "IN")
		// IN condition wraps the array as a single arg
		assert.NotEmpty(t, args)
	})
}

// TestSQLGenerationEdgeCases tests edge cases in SQL generation
func TestSQLGenerationEdgeCases(t *testing.T) {
	t.Run("select with no fields selects all", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.Eq("status", "u", "active")).
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.Contains(t, sql, "SELECT")
		assert.Contains(t, sql, "FROM")
	})

	t.Run("select with empty where array", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.NotContains(t, sql, "WHERE")
	})

	t.Run("select with empty sorts array", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.NotContains(t, sql, "ORDER BY")
	})

	t.Run("select with empty group by array", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Limit(0)

		sql, _, err := qb.Select()
		require.NoError(t, err)
		assert.NotContains(t, sql, "GROUP BY")
	})

	t.Run("count without limit shows default limit", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u")

		sql, _, err := qb.Count()
		require.NoError(t, err)
		assert.Contains(t, sql, "LIMIT")
	})

	t.Run("add with empty map", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u")

		sql, args, err := qb.Add(map[string]any{})
		require.NoError(t, err)
		assert.NotEmpty(t, sql)
		assert.Empty(t, args)
	})

	t.Run("edit with empty map but valid where", func(t *testing.T) {
		qb := supersaiyan.New("mysql", "users", "u").
			Where(supersaiyan.Eq("id", "u", 1))

		sql, args, err := qb.Edit(map[string]any{})
		// Empty map should cause an error (no values to update)
		assert.Error(t, err)
		_ = sql
		_ = args
	})
}
