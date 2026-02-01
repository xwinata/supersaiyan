package tests

import (
	"encoding/json"
	"os"
	"testing"

	"supersaiyan"

	"github.com/doug-martin/goqu/v9/exp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestUnmarshal_SQLBuilder tests unmarshaling of SQLBuilder with various scenarios
func TestUnmarshal_SQLBuilder(t *testing.T) {
	t.Run("comprehensive query with all features", func(t *testing.T) {
		jsonData := map[string]any{
			"dialect": "mysql",
			"table": map[string]any{
				"name":  "users",
				"alias": "u",
				"relations": []map[string]any{
					{
						"joinType": "INNER",
						"table": map[string]any{
							"name":  "orders",
							"alias": "o",
						},
						"on": []map[string]any{
							{
								"op":         "eq",
								"fieldName":  "user_id",
								"tableAlias": "o",
								"value": map[string]any{
									"name":       "id",
									"tableAlias": "u",
								},
							},
						},
					},
				},
			},
			"fields": []map[string]any{
				{
					"name":       "id",
					"tableAlias": "u",
				},
				{
					"name":       "username",
					"tableAlias": "u",
				},
				{
					"name":       "email",
					"tableAlias": "u",
				},
				{
					"fieldAlias": "order_count",
					"exp": map[string]any{
						"value": "COUNT(?)",
						"args": []map[string]any{
							{
								"name":       "id",
								"tableAlias": "o",
							},
						},
					},
				},
			},
			"wheres": []map[string]any{
				{
					"op":         "eq",
					"fieldName":  "status",
					"tableAlias": "u",
					"value":      "active",
				},
				{
					"op":         "gt",
					"fieldName":  "age",
					"tableAlias": "u",
					"value":      18,
				},
				{
					"op": "OR",
					"conditions": []map[string]any{
						{
							"op":         "like",
							"fieldName":  "email",
							"tableAlias": "u",
							"value":      "%@gmail.com",
						},
						{
							"op":         "like",
							"fieldName":  "email",
							"tableAlias": "u",
							"value":      "%@yahoo.com",
						},
					},
				},
			},
			"sorts": []map[string]any{
				{
					"name":       "created_at",
					"tableAlias": "u",
					"order":      "DESC",
				},
			},
			"groupBy": []map[string]any{
				{
					"name":       "id",
					"tableAlias": "u",
				},
			},
		}

		jsonBytes, err := json.Marshal(jsonData)
		require.NoError(t, err)

		var qb supersaiyan.SQLBuilder
		err = json.Unmarshal(jsonBytes, &qb)
		require.NoError(t, err)

		// Verify basic fields
		assert.Equal(t, "mysql", qb.Dialect)
		assert.Equal(t, "users", qb.Table.Name)
		assert.Equal(t, "u", qb.Table.Alias)

		// Verify fields
		assert.Len(t, qb.Fields, 4)
		assert.Equal(t, "id", qb.Fields[0].Name)
		assert.Equal(t, "u", qb.Fields[0].TableAlias)
		assert.Equal(t, "order_count", qb.Fields[3].FieldAlias)

		// Verify wheres
		assert.Len(t, qb.Wheres, 3)

		// First where - simple BoolOp
		boolOp1, ok := qb.Wheres[0].(supersaiyan.BoolOp)
		require.True(t, ok)
		assert.Equal(t, exp.EqOp, boolOp1.Op)
		assert.Equal(t, "status", boolOp1.FieldName)
		assert.Equal(t, "active", boolOp1.Value)

		// Second where - BoolOp with number
		boolOp2, ok := qb.Wheres[1].(supersaiyan.BoolOp)
		require.True(t, ok)
		assert.Equal(t, exp.GtOp, boolOp2.Op)
		assert.Equal(t, "age", boolOp2.FieldName)

		// Third where - WhereGroup with OR
		whereGroup, ok := qb.Wheres[2].(supersaiyan.WhereGroup)
		require.True(t, ok)
		assert.Equal(t, exp.OrType, whereGroup.Op)
		assert.Len(t, whereGroup.Conditions, 2)

		// Verify relations
		assert.Len(t, qb.Table.Relations, 1)
		assert.Equal(t, exp.InnerJoinType, qb.Table.Relations[0].JoinType)
		assert.Equal(t, "orders", qb.Table.Relations[0].Table.Name)

		// Verify sorts
		assert.Len(t, qb.Sorts, 1)
		assert.Equal(t, "created_at", qb.Sorts[0].Name)
		assert.Equal(t, exp.DescSortDir, qb.Sorts[0].Order)

		// Verify group by
		assert.Len(t, qb.GroupBy, 1)
		assert.Equal(t, "id", qb.GroupBy[0].Name)
	})

	t.Run("simple query", func(t *testing.T) {
		jsonData := map[string]any{
			"dialect": "postgres",
			"table": map[string]any{
				"name":  "products",
				"alias": "p",
			},
			"fields": []map[string]any{
				{
					"name":       "id",
					"tableAlias": "p",
				},
				{
					"name":       "name",
					"tableAlias": "p",
				},
			},
			"wheres": []map[string]any{
				{
					"op":         "in",
					"fieldName":  "category",
					"tableAlias": "p",
					"value":      []string{"electronics", "computers"},
				},
			},
		}

		jsonBytes, err := json.Marshal(jsonData)
		require.NoError(t, err)

		var qb supersaiyan.SQLBuilder
		err = json.Unmarshal(jsonBytes, &qb)
		require.NoError(t, err)

		assert.Equal(t, "postgres", qb.Dialect)
		assert.Len(t, qb.Fields, 2)
		assert.Len(t, qb.Wheres, 1)
	})
}

// TestUnmarshal_BoolOp tests unmarshaling of BoolOp
func TestUnmarshal_BoolOp(t *testing.T) {
	tests := []struct {
		name     string
		jsonData map[string]any
		expected supersaiyan.BoolOp
	}{
		{
			name: "eq operation",
			jsonData: map[string]any{
				"op":         "eq",
				"fieldName":  "status",
				"tableAlias": "u",
				"value":      "active",
			},
			expected: supersaiyan.BoolOp{
				Op:         exp.EqOp,
				FieldName:  "status",
				TableAlias: "u",
				Value:      "active",
			},
		},
		{
			name: "in operation with array",
			jsonData: map[string]any{
				"op":         "in",
				"fieldName":  "id",
				"tableAlias": "u",
				"value":      []any{1.0, 2.0, 3.0}, // JSON numbers are float64
			},
			expected: supersaiyan.BoolOp{
				Op:         exp.InOp,
				FieldName:  "id",
				TableAlias: "u",
				Value:      []any{1.0, 2.0, 3.0},
			},
		},
		{
			name: "like operation",
			jsonData: map[string]any{
				"op":         "like",
				"fieldName":  "email",
				"tableAlias": "u",
				"value":      "%@example.com",
			},
			expected: supersaiyan.BoolOp{
				Op:         exp.LikeOp,
				FieldName:  "email",
				TableAlias: "u",
				Value:      "%@example.com",
			},
		},
		{
			name: "gt operation with number",
			jsonData: map[string]any{
				"op":         "gt",
				"fieldName":  "age",
				"tableAlias": "u",
				"value":      18.0,
			},
			expected: supersaiyan.BoolOp{
				Op:         exp.GtOp,
				FieldName:  "age",
				TableAlias: "u",
				Value:      18.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.jsonData)
			require.NoError(t, err)

			var boolOp supersaiyan.BoolOp
			err = json.Unmarshal(jsonBytes, &boolOp)
			require.NoError(t, err)

			assert.Equal(t, tt.expected.Op, boolOp.Op)
			assert.Equal(t, tt.expected.FieldName, boolOp.FieldName)
			assert.Equal(t, tt.expected.TableAlias, boolOp.TableAlias)
			assert.Equal(t, tt.expected.Value, boolOp.Value)
		})
	}
}

// TestUnmarshal_RangeOp tests unmarshaling of RangeOp
func TestUnmarshal_RangeOp(t *testing.T) {
	t.Run("between operation", func(t *testing.T) {
		jsonData := map[string]any{
			"op":         "between",
			"fieldName":  "price",
			"tableAlias": "p",
			"start":      100.0,
			"end":        1000.0,
		}

		jsonBytes, err := json.Marshal(jsonData)
		require.NoError(t, err)

		var rangeOp supersaiyan.RangeOp
		err = json.Unmarshal(jsonBytes, &rangeOp)
		require.NoError(t, err)

		assert.Equal(t, exp.BetweenOp, rangeOp.Op)
		assert.Equal(t, "price", rangeOp.FieldName)
		assert.Equal(t, "p", rangeOp.TableAlias)
		assert.Equal(t, 100.0, rangeOp.Start)
		assert.Equal(t, 1000.0, rangeOp.End)
	})

	t.Run("not between operation", func(t *testing.T) {
		jsonData := map[string]any{
			"op":         "notBetween",
			"fieldName":  "age",
			"tableAlias": "u",
			"start":      18.0,
			"end":        65.0,
		}

		jsonBytes, err := json.Marshal(jsonData)
		require.NoError(t, err)

		var rangeOp supersaiyan.RangeOp
		err = json.Unmarshal(jsonBytes, &rangeOp)
		require.NoError(t, err)

		assert.Equal(t, exp.NotBetweenOp, rangeOp.Op)
		assert.Equal(t, "age", rangeOp.FieldName)
	})
}

// TestUnmarshal_WhereGroup tests unmarshaling of WhereGroup
func TestUnmarshal_WhereGroup(t *testing.T) {
	t.Run("OR group with multiple conditions", func(t *testing.T) {
		jsonData := map[string]any{
			"op": "OR",
			"conditions": []map[string]any{
				{
					"op":         "eq",
					"fieldName":  "status",
					"tableAlias": "u",
					"value":      "active",
				},
				{
					"op":         "eq",
					"fieldName":  "status",
					"tableAlias": "u",
					"value":      "pending",
				},
			},
		}

		jsonBytes, err := json.Marshal(jsonData)
		require.NoError(t, err)

		var whereGroup supersaiyan.WhereGroup
		err = json.Unmarshal(jsonBytes, &whereGroup)
		require.NoError(t, err)

		assert.Equal(t, exp.OrType, whereGroup.Op)
		assert.Len(t, whereGroup.Conditions, 2)

		boolOp1, ok := whereGroup.Conditions[0].(supersaiyan.BoolOp)
		require.True(t, ok)
		assert.Equal(t, "active", boolOp1.Value)

		boolOp2, ok := whereGroup.Conditions[1].(supersaiyan.BoolOp)
		require.True(t, ok)
		assert.Equal(t, "pending", boolOp2.Value)
	})

	t.Run("nested groups", func(t *testing.T) {
		jsonData := map[string]any{
			"op": "AND",
			"conditions": []map[string]any{
				{
					"op":         "eq",
					"fieldName":  "active",
					"tableAlias": "u",
					"value":      true,
				},
				{
					"op": "OR",
					"conditions": []map[string]any{
						{
							"op":         "eq",
							"fieldName":  "role",
							"tableAlias": "u",
							"value":      "admin",
						},
						{
							"op":         "eq",
							"fieldName":  "role",
							"tableAlias": "u",
							"value":      "moderator",
						},
					},
				},
			},
		}

		jsonBytes, err := json.Marshal(jsonData)
		require.NoError(t, err)

		var whereGroup supersaiyan.WhereGroup
		err = json.Unmarshal(jsonBytes, &whereGroup)
		require.NoError(t, err)

		assert.Equal(t, exp.AndType, whereGroup.Op)
		assert.Len(t, whereGroup.Conditions, 2)

		// First condition is BoolOp
		_, ok := whereGroup.Conditions[0].(supersaiyan.BoolOp)
		require.True(t, ok)

		// Second condition is nested WhereGroup
		nestedGroup, ok := whereGroup.Conditions[1].(supersaiyan.WhereGroup)
		require.True(t, ok)
		assert.Equal(t, exp.OrType, nestedGroup.Op)
		assert.Len(t, nestedGroup.Conditions, 2)
	})
}

// TestUnmarshal_Case tests unmarshaling of Case expressions
func TestUnmarshal_Case(t *testing.T) {
	t.Run("case with when/then/else", func(t *testing.T) {
		jsonData := map[string]any{
			"conditions": []map[string]any{
				{
					"when": map[string]any{
						"op":         "eq",
						"fieldName":  "status",
						"tableAlias": "u",
						"value":      "active",
					},
					"then": "Active User",
				},
				{
					"when": map[string]any{
						"op":         "eq",
						"fieldName":  "status",
						"tableAlias": "u",
						"value":      "inactive",
					},
					"then": "Inactive User",
				},
			},
			"else": "Unknown",
		}

		jsonBytes, err := json.Marshal(jsonData)
		require.NoError(t, err)

		var caseExpr supersaiyan.Case
		err = json.Unmarshal(jsonBytes, &caseExpr)
		require.NoError(t, err)

		assert.Len(t, caseExpr.Conditions, 2)
		assert.Equal(t, "Unknown", caseExpr.Else)

		// Verify first condition
		whenCond1, ok := caseExpr.Conditions[0].When.(supersaiyan.BoolOp)
		require.True(t, ok)
		assert.Equal(t, "status", whenCond1.FieldName)
		assert.Equal(t, "Active User", caseExpr.Conditions[0].Then)
	})
}

// TestUnmarshal_Literal tests unmarshaling of Literal expressions
func TestUnmarshal_Literal(t *testing.T) {
	t.Run("literal with args", func(t *testing.T) {
		jsonData := map[string]any{
			"value": "COUNT(?)",
			"args": []map[string]any{
				{
					"name":       "id",
					"tableAlias": "u",
				},
			},
		}

		jsonBytes, err := json.Marshal(jsonData)
		require.NoError(t, err)

		var literal supersaiyan.Literal
		err = json.Unmarshal(jsonBytes, &literal)
		require.NoError(t, err)

		assert.Equal(t, "COUNT(?)", literal.Value)
		assert.Len(t, literal.Args, 1)
	})

	t.Run("literal without args", func(t *testing.T) {
		jsonData := map[string]any{
			"value": "NOW()",
		}

		jsonBytes, err := json.Marshal(jsonData)
		require.NoError(t, err)

		var literal supersaiyan.Literal
		err = json.Unmarshal(jsonBytes, &literal)
		require.NoError(t, err)

		assert.Equal(t, "NOW()", literal.Value)
		assert.Len(t, literal.Args, 0)
	})
}

// TestUnmarshal_Relation tests unmarshaling of Relation (joins)
func TestUnmarshal_Relation(t *testing.T) {
	t.Run("inner join with condition", func(t *testing.T) {
		jsonData := map[string]any{
			"joinType": "INNER",
			"table": map[string]any{
				"name":  "orders",
				"alias": "o",
			},
			"on": []map[string]any{
				{
					"op":         "eq",
					"fieldName":  "user_id",
					"tableAlias": "o",
					"value": map[string]any{
						"name":       "id",
						"tableAlias": "u",
					},
				},
			},
		}

		jsonBytes, err := json.Marshal(jsonData)
		require.NoError(t, err)

		var relation supersaiyan.Relation
		err = json.Unmarshal(jsonBytes, &relation)
		require.NoError(t, err)

		assert.Equal(t, exp.InnerJoinType, relation.JoinType)
		assert.Equal(t, "orders", relation.Table.Name)
		assert.Equal(t, "o", relation.Table.Alias)
		assert.Len(t, relation.On, 1)

		boolOp, ok := relation.On[0].(supersaiyan.BoolOp)
		require.True(t, ok)
		assert.Equal(t, "user_id", boolOp.FieldName)
	})
}

// TestUnmarshal_Field tests unmarshaling of Field
func TestUnmarshal_Field(t *testing.T) {
	t.Run("simple field", func(t *testing.T) {
		jsonData := map[string]any{
			"name":       "username",
			"tableAlias": "u",
		}

		jsonBytes, err := json.Marshal(jsonData)
		require.NoError(t, err)

		var field supersaiyan.Field
		err = json.Unmarshal(jsonBytes, &field)
		require.NoError(t, err)

		assert.Equal(t, "username", field.Name)
		assert.Equal(t, "u", field.TableAlias)
	})

	t.Run("field with alias", func(t *testing.T) {
		jsonData := map[string]any{
			"name":       "created_at",
			"tableAlias": "u",
			"fieldAlias": "registration_date",
		}

		jsonBytes, err := json.Marshal(jsonData)
		require.NoError(t, err)

		var field supersaiyan.Field
		err = json.Unmarshal(jsonBytes, &field)
		require.NoError(t, err)

		assert.Equal(t, "created_at", field.Name)
		assert.Equal(t, "registration_date", field.FieldAlias)
	})

	t.Run("field with expression", func(t *testing.T) {
		jsonData := map[string]any{
			"fieldAlias": "total",
			"exp": map[string]any{
				"value": "SUM(?)",
				"args": []map[string]any{
					{
						"name":       "amount",
						"tableAlias": "o",
					},
				},
			},
		}

		jsonBytes, err := json.Marshal(jsonData)
		require.NoError(t, err)

		var field supersaiyan.Field
		err = json.Unmarshal(jsonBytes, &field)
		require.NoError(t, err)

		assert.Equal(t, "total", field.FieldAlias)
		assert.NotNil(t, field.Exp)

		literal, ok := field.Exp.(supersaiyan.Literal)
		require.True(t, ok)
		assert.Equal(t, "SUM(?)", literal.Value)
	})
}

// TestMarshal_SQLBuilder tests marshaling and round-trip of SQLBuilder
func TestMarshal_SQLBuilder(t *testing.T) {
	t.Run("marshal and unmarshal simple query", func(t *testing.T) {
		original := supersaiyan.New("mysql", "users", "u").
			WithFields(supersaiyan.Field{Name: "id", TableAlias: "u"}).
			WithFields(supersaiyan.Field{Name: "username", TableAlias: "u"}).
			Where(supersaiyan.Eq("status", "u", "active")).
			OrderBy(supersaiyan.Desc("created_at", "u")).
			Limit(0) // No limit for cleaner comparison

		// Marshal to JSON
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		// Unmarshal from JSON
		var restored supersaiyan.SQLBuilder
		err = json.Unmarshal(jsonData, &restored)
		require.NoError(t, err)

		// Verify fields
		assert.Equal(t, original.Dialect, restored.Dialect)
		assert.Equal(t, original.Table.Name, restored.Table.Name)
		assert.Equal(t, original.Table.Alias, restored.Table.Alias)
		assert.Len(t, restored.Fields, 2)
		assert.Len(t, restored.Wheres, 1)
		assert.Len(t, restored.Sorts, 1)

		// Generate SQL from both and compare
		originalSQL, _, _ := original.Select()
		restoredSQL, _, _ := restored.Select()
		assert.Equal(t, originalSQL, restoredSQL)
	})

	t.Run("marshal and unmarshal with joins", func(t *testing.T) {
		original := supersaiyan.New("postgres", "orders", "o").
			WithFields(supersaiyan.Field{Name: "id", TableAlias: "o"}).
			WithFields(supersaiyan.Field{Name: "customer_name", TableAlias: "c"}).
			InnerJoin("customers", "c", supersaiyan.Eq("id", "c", supersaiyan.F("customer_id", supersaiyan.WithTable("o")))).
			Where(supersaiyan.Eq("status", "o", "completed")).
			Limit(0)

		// Marshal to JSON
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		// Unmarshal from JSON
		var restored supersaiyan.SQLBuilder
		err = json.Unmarshal(jsonData, &restored)
		require.NoError(t, err)

		// Verify
		assert.Len(t, restored.Table.Relations, 1)
		assert.Equal(t, "customers", restored.Table.Relations[0].Table.Name)

		// Generate SQL from both and compare
		originalSQL, _, _ := original.Select()
		restoredSQL, _, _ := restored.Select()
		assert.Equal(t, originalSQL, restoredSQL)
	})

	t.Run("marshal and unmarshal with OR conditions", func(t *testing.T) {
		original := supersaiyan.New("mysql", "products", "p").
			WithFields(supersaiyan.Field{Name: "id", TableAlias: "p"}).
			Where(
				supersaiyan.Or(
					supersaiyan.Eq("category", "p", "electronics"),
					supersaiyan.Eq("category", "p", "computers"),
				),
			).
			Limit(0)

		// Marshal to JSON
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		// Unmarshal from JSON
		var restored supersaiyan.SQLBuilder
		err = json.Unmarshal(jsonData, &restored)
		require.NoError(t, err)

		// Verify
		assert.Len(t, restored.Wheres, 1)

		// Generate SQL from both and compare
		originalSQL, _, _ := original.Select()
		restoredSQL, _, _ := restored.Select()
		assert.Equal(t, originalSQL, restoredSQL)
	})

	t.Run("marshal and unmarshal with BETWEEN", func(t *testing.T) {
		original := supersaiyan.New("mysql", "products", "p").
			WithFields(supersaiyan.Field{Name: "id", TableAlias: "p"}).
			Where(supersaiyan.Between("price", "p", 100, 1000)).
			Limit(0)

		// Marshal to JSON
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		// Unmarshal from JSON
		var restored supersaiyan.SQLBuilder
		err = json.Unmarshal(jsonData, &restored)
		require.NoError(t, err)

		// Generate SQL from both and compare
		originalSQL, _, _ := original.Select()
		restoredSQL, _, _ := restored.Select()
		assert.Equal(t, originalSQL, restoredSQL)
	})

	t.Run("marshal and unmarshal with complex field", func(t *testing.T) {
		original := supersaiyan.New("mysql", "orders", "o").
			WithFields(supersaiyan.Field{
				FieldAlias: "total",
				Exp:        supersaiyan.L("SUM(?)", supersaiyan.F("amount", supersaiyan.WithTable("o"))),
			}).
			Limit(0)

		// Marshal to JSON
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		// Unmarshal from JSON
		var restored supersaiyan.SQLBuilder
		err = json.Unmarshal(jsonData, &restored)
		require.NoError(t, err)

		// Verify
		assert.Len(t, restored.Fields, 1)
		assert.NotNil(t, restored.Fields[0].Exp)

		// Generate SQL from both and compare
		originalSQL, _, _ := original.Select()
		restoredSQL, _, _ := restored.Select()
		assert.Equal(t, originalSQL, restoredSQL)
	})

	t.Run("marshal and unmarshal with CASE expression", func(t *testing.T) {
		original := supersaiyan.New("mysql", "users", "u").
			WithFields(supersaiyan.Field{
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
			}).
			Limit(0)

		// Marshal to JSON
		jsonData, err := json.Marshal(original)
		require.NoError(t, err)

		// Unmarshal from JSON
		var restored supersaiyan.SQLBuilder
		err = json.Unmarshal(jsonData, &restored)
		require.NoError(t, err)

		// Verify
		assert.Len(t, restored.Fields, 1)
		assert.NotNil(t, restored.Fields[0].Exp)

		// Generate SQL from both and compare
		originalSQL, _, _ := original.Select()
		restoredSQL, _, _ := restored.Select()
		assert.Equal(t, originalSQL, restoredSQL)
	})
}

// TestMarshal_BoolOp tests marshaling of BoolOp
func TestMarshal_BoolOp(t *testing.T) {
	t.Run("marshal BoolOp", func(t *testing.T) {
		boolOp := supersaiyan.Eq("status", "u", "active")

		jsonData, err := json.Marshal(boolOp)
		require.NoError(t, err)
		assert.Contains(t, string(jsonData), `"op":"eq"`)
		assert.Contains(t, string(jsonData), `"fieldName":"status"`)
	})
}

// TestMarshal_RangeOp tests marshaling of RangeOp
func TestMarshal_RangeOp(t *testing.T) {
	t.Run("marshal RangeOp", func(t *testing.T) {
		rangeOp := supersaiyan.Between("price", "p", 100, 1000)

		jsonData, err := json.Marshal(rangeOp)
		require.NoError(t, err)
		assert.Contains(t, string(jsonData), `"op":"between"`)
		assert.Contains(t, string(jsonData), `"fieldName":"price"`)
	})
}

// TestMarshal_Field tests marshaling of Field
func TestMarshal_Field(t *testing.T) {
	t.Run("marshal simple field", func(t *testing.T) {
		field := supersaiyan.F("username", supersaiyan.WithTable("u"))

		jsonData, err := json.Marshal(field)
		require.NoError(t, err)
		assert.Contains(t, string(jsonData), `"name":"username"`)
		assert.Contains(t, string(jsonData), `"tableAlias":"u"`)
	})
}

// TestUnmarshal_YAML_SQLBuilder tests unmarshaling SQLBuilder from YAML
func TestUnmarshal_YAML_SQLBuilder(t *testing.T) {
	t.Run("comprehensive query from YAML file", func(t *testing.T) {
		// Read YAML file
		yamlData, err := os.ReadFile("sample_query.yaml")
		require.NoError(t, err)

		var qb supersaiyan.SQLBuilder
		err = yaml.Unmarshal(yamlData, &qb)
		require.NoError(t, err)

		// Verify basic fields
		assert.Equal(t, "mysql", qb.Dialect)
		assert.Equal(t, "users", qb.Table.Name)
		assert.Equal(t, "u", qb.Table.Alias)

		// Verify relations (joins)
		assert.Len(t, qb.Table.Relations, 2)
		assert.Equal(t, exp.InnerJoinType, qb.Table.Relations[0].JoinType)
		assert.Equal(t, "orders", qb.Table.Relations[0].Table.Name)
		assert.Equal(t, exp.LeftJoinType, qb.Table.Relations[1].JoinType)
		assert.Equal(t, "profiles", qb.Table.Relations[1].Table.Name)

		// Verify fields
		assert.Len(t, qb.Fields, 6)
		assert.Equal(t, "id", qb.Fields[0].Name)
		assert.Equal(t, "u", qb.Fields[0].TableAlias)
		assert.Equal(t, "order_count", qb.Fields[3].FieldAlias)
		assert.Equal(t, "total_spent", qb.Fields[4].FieldAlias)
		assert.Equal(t, "status_label", qb.Fields[5].FieldAlias)

		// Verify field with Literal expression
		literal, ok := qb.Fields[3].Exp.(supersaiyan.Literal)
		require.True(t, ok, "Expected Literal, got %T", qb.Fields[3].Exp)
		assert.Equal(t, "COUNT(?)", literal.Value)
		assert.Len(t, literal.Args, 1)

		// Verify field with Case expression
		caseExpr, ok := qb.Fields[5].Exp.(supersaiyan.Case)
		require.True(t, ok, "Expected Case, got %T", qb.Fields[5].Exp)
		assert.Len(t, caseExpr.Conditions, 2)
		assert.Equal(t, "Unknown", caseExpr.Else)

		// Verify wheres
		assert.Len(t, qb.Wheres, 6)

		// First where - simple BoolOp
		boolOp1, ok := qb.Wheres[0].(supersaiyan.BoolOp)
		require.True(t, ok)
		assert.Equal(t, exp.EqOp, boolOp1.Op)
		assert.Equal(t, "status", boolOp1.FieldName)
		assert.Equal(t, "active", boolOp1.Value)

		// Second where - BoolOp with number (YAML unmarshals numbers as float64)
		boolOp2, ok := qb.Wheres[1].(supersaiyan.BoolOp)
		require.True(t, ok)
		assert.Equal(t, exp.GtOp, boolOp2.Op)
		assert.Equal(t, "age", boolOp2.FieldName)
		// YAML unmarshals numbers as float64
		if floatVal, ok := boolOp2.Value.(float64); ok {
			assert.Equal(t, float64(18), floatVal)
		} else {
			assert.Equal(t, 18, boolOp2.Value)
		}

		// Third where - WhereGroup with OR
		whereGroup1, ok := qb.Wheres[2].(supersaiyan.WhereGroup)
		require.True(t, ok)
		assert.Equal(t, exp.OrType, whereGroup1.Op)
		assert.Len(t, whereGroup1.Conditions, 2)

		// Fourth where - IN operation with array
		boolOp3, ok := qb.Wheres[3].(supersaiyan.BoolOp)
		require.True(t, ok)
		assert.Equal(t, exp.InOp, boolOp3.Op)
		assert.Equal(t, "country", boolOp3.FieldName)

		// Fifth where - BETWEEN operation
		rangeOp, ok := qb.Wheres[4].(supersaiyan.RangeOp)
		require.True(t, ok)
		assert.Equal(t, exp.BetweenOp, rangeOp.Op)
		assert.Equal(t, "created_at", rangeOp.FieldName)
		assert.Equal(t, "2024-01-01", rangeOp.Start)
		assert.Equal(t, "2024-12-31", rangeOp.End)

		// Sixth where - nested AND/OR groups
		whereGroup2, ok := qb.Wheres[5].(supersaiyan.WhereGroup)
		require.True(t, ok)
		assert.Equal(t, exp.AndType, whereGroup2.Op)
		assert.Len(t, whereGroup2.Conditions, 2)

		// Verify nested OR group
		nestedOr, ok := whereGroup2.Conditions[1].(supersaiyan.WhereGroup)
		require.True(t, ok)
		assert.Equal(t, exp.OrType, nestedOr.Op)
		assert.Len(t, nestedOr.Conditions, 2)

		// Verify sorts
		assert.Len(t, qb.Sorts, 2)
		assert.Equal(t, "created_at", qb.Sorts[0].Name)
		assert.Equal(t, exp.DescSortDir, qb.Sorts[0].Order)
		assert.Equal(t, "username", qb.Sorts[1].Name)
		assert.Equal(t, exp.AscDir, qb.Sorts[1].Order)

		// Verify group by
		assert.Len(t, qb.GroupBy, 2)
		assert.Equal(t, "id", qb.GroupBy[0].Name)
		assert.Equal(t, "username", qb.GroupBy[1].Name)

		// Verify SQL generation works
		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.NotEmpty(t, sql)
		assert.NotNil(t, args)
	})

	t.Run("simple query from YAML string", func(t *testing.T) {
		yamlStr := `
dialect: postgres
table:
  name: products
  alias: p
fields:
  - name: id
    tableAlias: p
  - name: name
    tableAlias: p
  - name: price
    tableAlias: p
wheres:
  - op: in
    fieldName: category
    tableAlias: p
    value:
      - electronics
      - computers
  - op: between
    fieldName: price
    tableAlias: p
    start: 100
    end: 1000
sorts:
  - name: price
    tableAlias: p
    order: ASC
`

		var qb supersaiyan.SQLBuilder
		err := yaml.Unmarshal([]byte(yamlStr), &qb)
		require.NoError(t, err)

		assert.Equal(t, "postgres", qb.Dialect)
		assert.Equal(t, "products", qb.Table.Name)
		assert.Len(t, qb.Fields, 3)
		assert.Len(t, qb.Wheres, 2)
		assert.Len(t, qb.Sorts, 1)

		// Verify IN operation
		boolOp, ok := qb.Wheres[0].(supersaiyan.BoolOp)
		require.True(t, ok)
		assert.Equal(t, exp.InOp, boolOp.Op)

		// Verify BETWEEN operation (YAML unmarshals numbers as float64 or int depending on value)
		rangeOp, ok := qb.Wheres[1].(supersaiyan.RangeOp)
		require.True(t, ok)
		assert.Equal(t, exp.BetweenOp, rangeOp.Op)
		// YAML may unmarshal as int or float64
		if floatVal, ok := rangeOp.Start.(float64); ok {
			assert.Equal(t, float64(100), floatVal)
		} else {
			assert.Equal(t, 100, rangeOp.Start)
		}
		if floatVal, ok := rangeOp.End.(float64); ok {
			assert.Equal(t, float64(1000), floatVal)
		} else {
			assert.Equal(t, 1000, rangeOp.End)
		}

		// Verify SQL generation
		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.NotEmpty(t, sql)
		assert.NotNil(t, args)
	})

	t.Run("complex query with nested conditions", func(t *testing.T) {
		yamlStr := `
dialect: mysql
table:
  name: users
  alias: u
  relations:
    - joinType: INNER
      table:
        name: orders
        alias: o
      on:
        - op: eq
          fieldName: user_id
          tableAlias: o
          value:
            name: id
            tableAlias: u
fields:
  - name: id
    tableAlias: u
  - name: username
    tableAlias: u
  - fieldAlias: order_count
    exp:
      value: COUNT(?)
      args:
        - name: id
          tableAlias: o
wheres:
  - op: eq
    fieldName: status
    tableAlias: u
    value: active
  - op: OR
    conditions:
      - op: eq
        fieldName: role
        tableAlias: u
        value: admin
      - op: eq
        fieldName: role
        tableAlias: u
        value: moderator
  - op: between
    fieldName: age
    tableAlias: u
    start: 18
    end: 65
sorts:
  - name: created_at
    tableAlias: u
    order: DESC
groupBy:
  - name: id
    tableAlias: u
`

		var qb supersaiyan.SQLBuilder
		err := yaml.Unmarshal([]byte(yamlStr), &qb)
		require.NoError(t, err)

		// Verify structure
		assert.Equal(t, "mysql", qb.Dialect)
		assert.Equal(t, "users", qb.Table.Name)
		assert.Len(t, qb.Fields, 3)
		assert.Len(t, qb.Table.Relations, 1)
		assert.Len(t, qb.Wheres, 3)
		assert.Len(t, qb.Sorts, 1)
		assert.Len(t, qb.GroupBy, 1)

		// Verify join
		assert.Equal(t, exp.InnerJoinType, qb.Table.Relations[0].JoinType)
		assert.Equal(t, "orders", qb.Table.Relations[0].Table.Name)

		// Verify OR condition
		whereGroup, ok := qb.Wheres[1].(supersaiyan.WhereGroup)
		require.True(t, ok)
		assert.Equal(t, exp.OrType, whereGroup.Op)
		assert.Len(t, whereGroup.Conditions, 2)

		// Verify BETWEEN
		rangeOp, ok := qb.Wheres[2].(supersaiyan.RangeOp)
		require.True(t, ok)
		assert.Equal(t, exp.BetweenOp, rangeOp.Op)

		// Generate SQL to verify it works
		sql, args, err := qb.Select()
		require.NoError(t, err)
		assert.NotEmpty(t, sql)
		assert.NotNil(t, args)
	})
}
