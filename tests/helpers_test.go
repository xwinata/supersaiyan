package tests

import (
	"testing"

	"supersaiyan"

	"github.com/doug-martin/goqu/v9/exp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBoolOpHelpers tests boolean operation helper functions
func TestBoolOpHelpers(t *testing.T) {
	t.Run("Eq creates equality operation", func(t *testing.T) {
		op := supersaiyan.Eq("status", "u", "active")

		assert.Equal(t, exp.EqOp, op.Op)
		assert.Equal(t, "status", op.FieldName)
		assert.Equal(t, "u", op.TableAlias)
		assert.Equal(t, "active", op.Value)
	})

	t.Run("Neq creates not-equal operation", func(t *testing.T) {
		op := supersaiyan.Neq("status", "u", "inactive")

		assert.Equal(t, exp.NeqOp, op.Op)
		assert.Equal(t, "status", op.FieldName)
		assert.Equal(t, "inactive", op.Value)
	})

	t.Run("Gt creates greater-than operation", func(t *testing.T) {
		op := supersaiyan.Gt("age", "u", 18)

		assert.Equal(t, exp.GtOp, op.Op)
		assert.Equal(t, "age", op.FieldName)
		assert.Equal(t, 18, op.Value)
	})

	t.Run("Gte creates greater-than-or-equal operation", func(t *testing.T) {
		op := supersaiyan.Gte("age", "u", 21)

		assert.Equal(t, exp.GteOp, op.Op)
		assert.Equal(t, "age", op.FieldName)
		assert.Equal(t, 21, op.Value)
	})

	t.Run("Lt creates less-than operation", func(t *testing.T) {
		op := supersaiyan.Lt("age", "u", 65)

		assert.Equal(t, exp.LtOp, op.Op)
		assert.Equal(t, "age", op.FieldName)
		assert.Equal(t, 65, op.Value)
	})

	t.Run("Lte creates less-than-or-equal operation", func(t *testing.T) {
		op := supersaiyan.Lte("age", "u", 100)

		assert.Equal(t, exp.LteOp, op.Op)
		assert.Equal(t, "age", op.FieldName)
		assert.Equal(t, 100, op.Value)
	})

	t.Run("In creates IN operation", func(t *testing.T) {
		values := []string{"active", "pending", "verified"}
		op := supersaiyan.In("status", "u", values)

		assert.Equal(t, exp.InOp, op.Op)
		assert.Equal(t, "status", op.FieldName)
		assert.Equal(t, values, op.Value)
	})

	t.Run("NotIn creates NOT IN operation", func(t *testing.T) {
		values := []string{"banned", "deleted"}
		op := supersaiyan.NotIn("status", "u", values)

		assert.Equal(t, exp.NotInOp, op.Op)
		assert.Equal(t, "status", op.FieldName)
		assert.Equal(t, values, op.Value)
	})

	t.Run("Like creates LIKE operation", func(t *testing.T) {
		op := supersaiyan.Like("email", "u", "%@example.com")

		assert.Equal(t, exp.LikeOp, op.Op)
		assert.Equal(t, "email", op.FieldName)
		assert.Equal(t, "%@example.com", op.Value)
	})

	t.Run("ILike creates case-insensitive LIKE operation", func(t *testing.T) {
		op := supersaiyan.ILike("email", "u", "%@EXAMPLE.COM")

		assert.Equal(t, exp.ILikeOp, op.Op)
		assert.Equal(t, "email", op.FieldName)
		assert.Equal(t, "%@EXAMPLE.COM", op.Value)
	})

	t.Run("IsNull creates IS NULL operation", func(t *testing.T) {
		op := supersaiyan.IsNull("deleted_at", "u")

		assert.Equal(t, exp.IsOp, op.Op)
		assert.Equal(t, "deleted_at", op.FieldName)
		assert.Nil(t, op.Value)
	})

	t.Run("IsNotNull creates IS NOT NULL operation", func(t *testing.T) {
		op := supersaiyan.IsNotNull("email", "u")

		assert.Equal(t, exp.IsNotOp, op.Op)
		assert.Equal(t, "email", op.FieldName)
		assert.Nil(t, op.Value)
	})
}

// TestRangeOpHelpers tests range operation helper functions
func TestRangeOpHelpers(t *testing.T) {
	t.Run("Between creates BETWEEN operation", func(t *testing.T) {
		op := supersaiyan.Between("price", "p", 100, 1000)

		assert.Equal(t, exp.BetweenOp, op.Op)
		assert.Equal(t, "price", op.FieldName)
		assert.Equal(t, "p", op.TableAlias)
		assert.Equal(t, 100, op.Start)
		assert.Equal(t, 1000, op.End)
	})

	t.Run("NotBetween creates NOT BETWEEN operation", func(t *testing.T) {
		op := supersaiyan.NotBetween("age", "u", 18, 65)

		assert.Equal(t, exp.NotBetweenOp, op.Op)
		assert.Equal(t, "age", op.FieldName)
		assert.Equal(t, "u", op.TableAlias)
		assert.Equal(t, 18, op.Start)
		assert.Equal(t, 65, op.End)
	})

	t.Run("Between with dates", func(t *testing.T) {
		op := supersaiyan.Between("created_at", "u", "2024-01-01", "2024-12-31")

		assert.Equal(t, exp.BetweenOp, op.Op)
		assert.Equal(t, "created_at", op.FieldName)
		assert.Equal(t, "2024-01-01", op.Start)
		assert.Equal(t, "2024-12-31", op.End)
	})
}

// TestWhereGroupHelpers tests where group helper functions
func TestWhereGroupHelpers(t *testing.T) {
	t.Run("And creates AND group", func(t *testing.T) {
		group := supersaiyan.And(
			supersaiyan.Eq("status", "u", "active"),
			supersaiyan.Gt("age", "u", 18),
		)

		assert.Equal(t, exp.AndType, group.Op)
		assert.Len(t, group.Conditions, 2)
	})

	t.Run("Or creates OR group", func(t *testing.T) {
		group := supersaiyan.Or(
			supersaiyan.Eq("role", "u", "admin"),
			supersaiyan.Eq("role", "u", "moderator"),
		)

		assert.Equal(t, exp.OrType, group.Op)
		assert.Len(t, group.Conditions, 2)
	})

	t.Run("nested And/Or groups", func(t *testing.T) {
		group := supersaiyan.And(
			supersaiyan.Eq("status", "u", "active"),
			supersaiyan.Or(
				supersaiyan.Eq("role", "u", "admin"),
				supersaiyan.Eq("role", "u", "moderator"),
			),
		)

		assert.Equal(t, exp.AndType, group.Op)
		assert.Len(t, group.Conditions, 2)

		// Check nested OR group
		nestedOr, ok := group.Conditions[1].(supersaiyan.WhereGroup)
		require.True(t, ok)
		assert.Equal(t, exp.OrType, nestedOr.Op)
		assert.Len(t, nestedOr.Conditions, 2)
	})

	t.Run("And with single condition", func(t *testing.T) {
		group := supersaiyan.And(
			supersaiyan.Eq("status", "u", "active"),
		)

		assert.Equal(t, exp.AndType, group.Op)
		assert.Len(t, group.Conditions, 1)
	})

	t.Run("Or with multiple condition types", func(t *testing.T) {
		group := supersaiyan.Or(
			supersaiyan.Eq("status", "u", "active"),
			supersaiyan.Between("age", "u", 18, 65),
			supersaiyan.Like("email", "u", "%@example.com"),
		)

		assert.Equal(t, exp.OrType, group.Op)
		assert.Len(t, group.Conditions, 3)
	})
}

// TestSortHelpers tests sort helper functions
func TestSortHelpers(t *testing.T) {
	t.Run("Asc creates ascending sort", func(t *testing.T) {
		sort := supersaiyan.Asc("username", "u")

		assert.Equal(t, "username", sort.Name)
		assert.Equal(t, "u", sort.TableAlias)
		assert.Equal(t, exp.AscDir, sort.Order)
	})

	t.Run("Desc creates descending sort", func(t *testing.T) {
		sort := supersaiyan.Desc("created_at", "u")

		assert.Equal(t, "created_at", sort.Name)
		assert.Equal(t, "u", sort.TableAlias)
		assert.Equal(t, exp.DescSortDir, sort.Order)
	})

	t.Run("Asc without table alias", func(t *testing.T) {
		sort := supersaiyan.Asc("username", "")

		assert.Equal(t, "username", sort.Name)
		assert.Empty(t, sort.TableAlias)
		assert.Equal(t, exp.AscDir, sort.Order)
	})
}

// TestFieldHelper tests field helper function
func TestFieldHelper(t *testing.T) {
	t.Run("F creates field reference", func(t *testing.T) {
		field := supersaiyan.F("username", supersaiyan.WithTable("u"))

		assert.Equal(t, "username", field.Name)
		assert.Equal(t, "u", field.TableAlias)
		assert.Empty(t, field.FieldAlias)
		assert.Nil(t, field.Exp)
	})

	t.Run("F without table alias", func(t *testing.T) {
		field := supersaiyan.F("username")

		assert.Equal(t, "username", field.Name)
		assert.Empty(t, field.TableAlias)
	})
}

// TestParseBoolOperation tests ParseBoolOperation function
func TestParseBoolOperation(t *testing.T) {
	tests := []struct {
		input    string
		expected exp.BooleanOperation
	}{
		{"=", exp.EqOp},
		{"!=", exp.NeqOp},
		{"<>", exp.NeqOp},
		{">", exp.GtOp},
		{">=", exp.GteOp},
		{"<", exp.LtOp},
		{"<=", exp.LteOp},
		{"is", exp.IsOp},
		{"IS", exp.IsOp},
		{"is not", exp.IsNotOp},
		{"IS NOT", exp.IsNotOp},
		{"in", exp.InOp},
		{"IN", exp.InOp},
		{"not in", exp.NotInOp},
		{"NOT IN", exp.NotInOp},
		{"like", exp.LikeOp},
		{"LIKE", exp.LikeOp},
		{"not like", exp.NotLikeOp},
		{"NOT LIKE", exp.NotLikeOp},
		{"ilike", exp.ILikeOp},
		{"ILIKE", exp.ILikeOp},
		{"not ilike", exp.NotILikeOp},
		{"NOT ILIKE", exp.NotILikeOp},
		{"~", exp.RegexpLikeOp},
		{"!~", exp.RegexpNotLikeOp},
		{"~*", exp.RegexpILikeOp},
		{"!~*", exp.RegexpNotILikeOp},
		{"unknown", exp.EqOp}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := supersaiyan.ParseBoolOperation(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseRangeOperation tests ParseRangeOperation function
func TestParseRangeOperation(t *testing.T) {
	tests := []struct {
		input    string
		expected exp.RangeOperation
	}{
		{"between", exp.BetweenOp},
		{"BETWEEN", exp.BetweenOp},
		{"not between", exp.NotBetweenOp},
		{"NOT BETWEEN", exp.NotBetweenOp},
		{"not", exp.NotBetweenOp},
		{"NOT", exp.NotBetweenOp},
		{"unknown", exp.BetweenOp}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := supersaiyan.ParseRangeOperation(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseSortDirection tests ParseSortDirection function
func TestParseSortDirection(t *testing.T) {
	tests := []struct {
		input    string
		expected exp.SortDirection
	}{
		{"ASC", exp.AscDir},
		{"asc", exp.AscDir},
		{"DESC", exp.DescSortDir},
		{"desc", exp.DescSortDir},
		{"unknown", exp.AscDir}, // default
		{"", exp.AscDir},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := supersaiyan.ParseSortDirection(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseJoinType tests ParseJoinType function
func TestParseJoinType(t *testing.T) {
	tests := []struct {
		input    string
		expected exp.JoinType
	}{
		{"inner", exp.InnerJoinType},
		{"INNER", exp.InnerJoinType},
		{"left", exp.LeftJoinType},
		{"LEFT", exp.LeftJoinType},
		{"right", exp.RightJoinType},
		{"RIGHT", exp.RightJoinType},
		{"unknown", exp.InnerJoinType}, // default
		{"", exp.InnerJoinType},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := supersaiyan.ParseJoinType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBoolOperatorStrings tests the BoolOperatorStrings constant
func TestBoolOperatorStrings(t *testing.T) {
	t.Run("contains all expected operators", func(t *testing.T) {
		expectedOps := []string{
			" not ilike ",
			" not like ",
			" not in ",
			" is not ",
			"!~*",
			"~*",
			"!~",
			"<>",
			">=",
			"<=",
			"!=",
			"is",
			" in ",
			" ilike ",
			" like ",
			"~",
			">",
			"<",
			"=",
		}

		assert.Equal(t, expectedOps, supersaiyan.BoolOperatorStrings)
	})
}

// TestConditionInterface tests that types implement Condition interface
func TestConditionInterface(t *testing.T) {
	t.Run("BoolOp implements Condition", func(t *testing.T) {
		var _ supersaiyan.Condition = supersaiyan.BoolOp{}
	})

	t.Run("RangeOp implements Condition", func(t *testing.T) {
		var _ supersaiyan.Condition = supersaiyan.RangeOp{}
	})

	t.Run("WhereGroup implements Condition", func(t *testing.T) {
		var _ supersaiyan.Condition = supersaiyan.WhereGroup{}
	})
}

// TestComplexConditionBuilding tests building complex conditions
func TestComplexConditionBuilding(t *testing.T) {
	t.Run("complex nested conditions", func(t *testing.T) {
		condition := supersaiyan.And(
			supersaiyan.Eq("status", "u", "active"),
			supersaiyan.Or(
				supersaiyan.And(
					supersaiyan.Eq("role", "u", "admin"),
					supersaiyan.Gt("age", "u", 21),
				),
				supersaiyan.And(
					supersaiyan.Eq("role", "u", "user"),
					supersaiyan.Between("age", "u", 18, 65),
				),
			),
		)

		assert.Equal(t, exp.AndType, condition.Op)
		assert.Len(t, condition.Conditions, 2)

		// Verify nested structure
		nestedOr, ok := condition.Conditions[1].(supersaiyan.WhereGroup)
		require.True(t, ok)
		assert.Equal(t, exp.OrType, nestedOr.Op)
		assert.Len(t, nestedOr.Conditions, 2)
	})

	t.Run("mixed condition types in group", func(t *testing.T) {
		condition := supersaiyan.Or(
			supersaiyan.Eq("status", "u", "active"),
			supersaiyan.Between("age", "u", 18, 65),
			supersaiyan.In("country", "u", []string{"US", "CA", "UK"}),
			supersaiyan.Like("email", "u", "%@example.com"),
			supersaiyan.IsNotNull("verified_at", "u"),
		)

		assert.Equal(t, exp.OrType, condition.Op)
		assert.Len(t, condition.Conditions, 5)

		// Verify each condition type
		_, ok1 := condition.Conditions[0].(supersaiyan.BoolOp)
		assert.True(t, ok1)

		_, ok2 := condition.Conditions[1].(supersaiyan.RangeOp)
		assert.True(t, ok2)

		_, ok3 := condition.Conditions[2].(supersaiyan.BoolOp)
		assert.True(t, ok3)
	})
}
