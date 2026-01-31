package supersaiyan

import "github.com/doug-martin/goqu/v9/exp"

// Condition represents any type that can be used as a condition in WHERE or ON clauses.
// This interface ensures type safety while allowing flexibility.
type Condition interface {
	// toExpression converts the condition to a goqu expression.
	toExpression() exp.Expression
}

// Ensure our types implement Condition
var (
	_ Condition = BoolOp{}
	_ Condition = RangeOp{}
	_ Condition = WhereGroup{}
)

// toExpression for BoolOp
func (bo BoolOp) toExpression() exp.Expression {
	return bo.expression()
}

// toExpression for RangeOp
func (ro RangeOp) toExpression() exp.Expression {
	return ro.expression()
}

// toExpression for WhereGroup
func (wg WhereGroup) toExpression() exp.Expression {
	return wg.expression()
}
