package supersaiyan

import (
	"encoding/json"
	"fmt"

	"github.com/doug-martin/goqu/v9/exp"
)

// unmarshalValue tries to unmarshal a value, checking if it's an expression or a simple value.
func unmarshalValue(data []byte) (any, error) {
	// Try to unmarshal as an expression first
	value, err := unmarshalExpression(data)
	if err == nil {
		return value, nil
	}

	// If it fails, try as a simple value
	var simpleValue any
	if err := json.Unmarshal(data, &simpleValue); err != nil {
		return nil, err
	}
	return simpleValue, nil
}

// unmarshalCondition detects and unmarshals different condition types.
func unmarshalCondition(data []byte) (any, error) {
	// Try to detect the type by checking for specific fields
	var typeDetector map[string]json.RawMessage
	if err := json.Unmarshal(data, &typeDetector); err != nil {
		return nil, err
	}

	// Check for BoolOp (has "op" and "fieldName")
	if _, hasOp := typeDetector["op"]; hasOp {
		if _, hasFieldName := typeDetector["fieldName"]; hasFieldName {
			// Check if it's a RangeOp (has "start" and "end")
			if _, hasStart := typeDetector["start"]; hasStart {
				var rangeOp RangeOp
				if err := json.Unmarshal(data, &rangeOp); err != nil {
					return nil, err
				}
				return rangeOp, nil
			}
			// It's a BoolOp
			var boolOp BoolOp
			if err := json.Unmarshal(data, &boolOp); err != nil {
				return nil, err
			}
			return boolOp, nil
		}
		// Check for WhereGroup (has "op" and "conditions")
		if _, hasConditions := typeDetector["conditions"]; hasConditions {
			var whereGroup WhereGroup
			if err := json.Unmarshal(data, &whereGroup); err != nil {
				return nil, err
			}
			return whereGroup, nil
		}
	}

	return nil, fmt.Errorf("unknown condition type")
}

// unmarshalExpression detects and unmarshals different expression types.
func unmarshalExpression(data []byte) (any, error) {
	// Try to detect the type by checking for specific fields
	var typeDetector map[string]json.RawMessage
	if err := json.Unmarshal(data, &typeDetector); err != nil {
		return nil, err
	}

	// Check for Case (has "conditions" array with "when"/"then")
	if conditionsRaw, hasConditions := typeDetector["conditions"]; hasConditions {
		var testConditions []map[string]any
		if err := json.Unmarshal(conditionsRaw, &testConditions); err == nil &&
			len(testConditions) > 0 {
			if _, hasWhen := testConditions[0]["when"]; hasWhen {
				var caseExpr Case
				if err := json.Unmarshal(data, &caseExpr); err != nil {
					return nil, err
				}
				return caseExpr, nil
			}
		}
	}

	// Check for Coalesce (has "fields" array)
	if _, hasFields := typeDetector["fields"]; hasFields {
		var coalesce Coalesce
		if err := json.Unmarshal(data, &coalesce); err != nil {
			return nil, err
		}
		return coalesce, nil
	}

	// Check for Literal (has "value" string)
	if _, hasValue := typeDetector["value"]; hasValue {
		var literal Literal
		if err := json.Unmarshal(data, &literal); err != nil {
			return nil, err
		}
		return literal, nil
	}

	// Check for Field
	if _, hasName := typeDetector["name"]; hasName {
		var field Field
		if err := json.Unmarshal(data, &field); err != nil {
			return nil, err
		}
		return field, nil
	}

	// Check for BoolOp
	if _, hasOp := typeDetector["op"]; hasOp {
		if _, hasFieldName := typeDetector["fieldName"]; hasFieldName {
			// Check if it's a RangeOp
			if _, hasStart := typeDetector["start"]; hasStart {
				var rangeOp RangeOp
				if err := json.Unmarshal(data, &rangeOp); err != nil {
					return nil, err
				}
				return rangeOp, nil
			}
			// It's a BoolOp
			var boolOp BoolOp
			if err := json.Unmarshal(data, &boolOp); err != nil {
				return nil, err
			}
			return boolOp, nil
		}
	}

	return nil, fmt.Errorf("unknown expression type")
}

// Helper functions to convert operations to strings
func boolOpToString(op exp.BooleanOperation) string {
	switch op {
	case exp.EqOp:
		return "eq"
	case exp.NeqOp:
		return "neq"
	case exp.IsOp:
		return "is"
	case exp.IsNotOp:
		return "isNot"
	case exp.GtOp:
		return "gt"
	case exp.GteOp:
		return "gte"
	case exp.LtOp:
		return "lt"
	case exp.LteOp:
		return "lte"
	case exp.InOp:
		return "in"
	case exp.NotInOp:
		return "notIn"
	case exp.LikeOp:
		return "like"
	case exp.NotLikeOp:
		return "notLike"
	case exp.ILikeOp:
		return "iLike"
	case exp.NotILikeOp:
		return "notILike"
	default:
		return "eq"
	}
}

func stringToBoolOp(s string) exp.BooleanOperation {
	switch s {
	case "eq":
		return exp.EqOp
	case "neq":
		return exp.NeqOp
	case "is":
		return exp.IsOp
	case "isNot":
		return exp.IsNotOp
	case "gt":
		return exp.GtOp
	case "gte":
		return exp.GteOp
	case "lt":
		return exp.LtOp
	case "lte":
		return exp.LteOp
	case "in":
		return exp.InOp
	case "notIn":
		return exp.NotInOp
	case "like":
		return exp.LikeOp
	case "notLike":
		return exp.NotLikeOp
	case "iLike":
		return exp.ILikeOp
	case "notILike":
		return exp.NotILikeOp
	default:
		return exp.EqOp
	}
}

func rangeOpToString(op exp.RangeOperation) string {
	switch op {
	case exp.BetweenOp:
		return "between"
	case exp.NotBetweenOp:
		return "notBetween"
	default:
		return "between"
	}
}

func stringToRangeOp(s string) exp.RangeOperation {
	switch s {
	case "between":
		return exp.BetweenOp
	case "notBetween":
		return exp.NotBetweenOp
	default:
		return exp.BetweenOp
	}
}
