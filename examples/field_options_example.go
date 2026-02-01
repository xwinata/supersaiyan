package main

import (
	"fmt"
	"supersaiyan"
)

func main() {
	fmt.Println("=== SuperSaiyan Field Options Examples ===")
	fmt.Println()

	// Example 1: Simple field with alias
	example1()

	// Example 2: Field with expression
	example2()

	// Example 3: Complex query with multiple field types
	example3()

	// Example 4: CASE expression
	example4()

	// Example 5: COALESCE
	example5()
}

func example1() {
	fmt.Println("Example 1: Simple field with alias")
	fmt.Println("-----------------------------------")

	qb := supersaiyan.New("mysql", "users", "u").
		WithFields(
			supersaiyan.F("id", supersaiyan.WithTable("u")),
			supersaiyan.F("username", supersaiyan.WithTable("u")),
			supersaiyan.F("created_at", supersaiyan.WithTable("u"), supersaiyan.WithAlias("registration_date")),
			supersaiyan.F("email"), // Field without table alias
		).
		Limit(10)

	sql, args, _ := qb.Select()
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Args: %v\n\n", args)
}

func example2() {
	fmt.Println("Example 2: Field with expression (COUNT)")
	fmt.Println("----------------------------------------")

	qb := supersaiyan.New("mysql", "orders", "o").
		WithFields(
			supersaiyan.F("user_id", supersaiyan.WithTable("o")),
			supersaiyan.Exp("order_count", supersaiyan.L("COUNT(?)", supersaiyan.F("id", supersaiyan.WithTable("o")))),
		).
		GroupByFields(supersaiyan.F("user_id", supersaiyan.WithTable("o"))).
		Limit(0)

	sql, args, _ := qb.Select()
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Args: %v\n\n", args)
}

func example3() {
	fmt.Println("Example 3: Complex query with aggregations")
	fmt.Println("------------------------------------------")

	qb := supersaiyan.New("mysql", "orders", "o").
		WithFields(
			supersaiyan.F("user_id", supersaiyan.WithTable("o")),
			supersaiyan.F("status", supersaiyan.WithTable("o")),
			supersaiyan.Exp("order_count", supersaiyan.L("COUNT(?)", supersaiyan.F("id", supersaiyan.WithTable("o")))),
			supersaiyan.Exp("total_amount", supersaiyan.L("SUM(?)", supersaiyan.F("amount", supersaiyan.WithTable("o")))),
			supersaiyan.Exp("avg_amount", supersaiyan.L("AVG(?)", supersaiyan.F("amount", supersaiyan.WithTable("o")))),
		).
		Where(supersaiyan.Eq("status", "o", "completed")).
		GroupByFields(
			supersaiyan.F("user_id", supersaiyan.WithTable("o")),
			supersaiyan.F("status", supersaiyan.WithTable("o")),
		).
		OrderBy(supersaiyan.Desc("total_amount", "")).
		Limit(10)

	sql, args, _ := qb.Select()
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Args: %v\n\n", args)
}

func example4() {
	fmt.Println("Example 4: CASE expression")
	fmt.Println("--------------------------")

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
					{
						When: supersaiyan.Eq("status", "u", "pending"),
						Then: "Pending User",
					},
				},
				Else: "Unknown",
			}),
		).
		Where(supersaiyan.IsNotNull("email", "u")).
		Limit(0)

	sql, args, _ := qb.Select()
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Args: %v\n\n", args)
}

func example5() {
	fmt.Println("Example 5: COALESCE")
	fmt.Println("-------------------")

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

	sql, args, _ := qb.Select()
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Args: %v\n\n", args)
}
