package main

import (
	"encoding/json"
	"fmt"
	"supersaiyan"

	"gopkg.in/yaml.v3"
)

func main() {
	fmt.Println("=== SuperSaiyan SQL Builder Examples ===")
	fmt.Println()

	// Basic Examples
	basicSelect()
	selectWithConditions()
	selectWithJoins()

	// Field Helpers
	fieldHelpers()
	literalExpressions()

	// Advanced Features
	aggregations()
	caseExpressions()
	coalesceExpressions()

	// Mutations
	insertExample()
	updateExample()
	deleteExample()

	// Unmarshaling
	jsonUnmarshal()
	yamlUnmarshal()
}

func basicSelect() {
	fmt.Println("Example: Basic SELECT")
	fmt.Println("---------------------")

	qb := supersaiyan.New("mysql", "users", "u").
		WithFields(
			supersaiyan.F("id", supersaiyan.WithTable("u")),
			supersaiyan.F("username", supersaiyan.WithTable("u")),
			supersaiyan.F("email", supersaiyan.WithTable("u")),
		).
		Where(supersaiyan.Eq("status", "u", "active")).
		OrderBy(supersaiyan.Desc("created_at", "u")).
		Limit(10)

	sql, args, _ := qb.Select()
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Args: %v\n\n", args)
}

func selectWithConditions() {
	fmt.Println("Example: Complex WHERE Conditions")
	fmt.Println("---------------------------------")

	qb := supersaiyan.New("mysql", "users", "u").
		WithFields(
			supersaiyan.F("id", supersaiyan.WithTable("u")),
			supersaiyan.F("username", supersaiyan.WithTable("u")),
		).
		Where(
			supersaiyan.Eq("status", "u", "active"),
			supersaiyan.Or(
				supersaiyan.Gt("age", "u", 18),
				supersaiyan.Eq("verified", "u", true),
			),
			supersaiyan.Like("email", "u", "%@example.com"),
		).
		Limit(10)

	sql, args, _ := qb.Select()
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Args: %v\n\n", args)
}

func selectWithJoins() {
	fmt.Println("Example: Joins")
	fmt.Println("--------------")

	qb := supersaiyan.New("mysql", "users", "u").
		WithFields(
			supersaiyan.F("id", supersaiyan.WithTable("u")),
			supersaiyan.F("username", supersaiyan.WithTable("u")),
			supersaiyan.F("order_id", supersaiyan.WithTable("o")),
			supersaiyan.F("total", supersaiyan.WithTable("o")),
		).
		InnerJoin("orders", "o", supersaiyan.Eq("user_id", "o", supersaiyan.F("id", supersaiyan.WithTable("u")))).
		LeftJoin("profiles", "p", supersaiyan.Eq("user_id", "p", supersaiyan.F("id", supersaiyan.WithTable("u")))).
		Where(supersaiyan.Eq("status", "o", "completed")).
		Limit(10)

	sql, args, _ := qb.Select()
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Args: %v\n\n", args)
}

func fieldHelpers() {
	fmt.Println("Example: Field Helpers")
	fmt.Println("----------------------")

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

func literalExpressions() {
	fmt.Println("Example: Literal Expressions")
	fmt.Println("----------------------------")

	qb := supersaiyan.New("mysql", "orders", "o").
		WithFields(
			supersaiyan.F("user_id", supersaiyan.WithTable("o")),
			supersaiyan.Exp("order_count", supersaiyan.L("COUNT(?)", supersaiyan.F("id", supersaiyan.WithTable("o")))),
		).
		GroupByFields(supersaiyan.F("user_id", supersaiyan.WithTable("o"))).
		Limit(10)

	sql, args, _ := qb.Select()
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Args: %v\n\n", args)
}

func aggregations() {
	fmt.Println("Example: Aggregations with GROUP BY")
	fmt.Println("-----------------------------------")

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

func caseExpressions() {
	fmt.Println("Example: CASE Expressions")
	fmt.Println("-------------------------")

	qb := supersaiyan.New("mysql", "users", "u").
		WithFields(
			supersaiyan.F("id", supersaiyan.WithTable("u")),
			supersaiyan.F("username", supersaiyan.WithTable("u")),
			supersaiyan.Exp("status_label", supersaiyan.C(
				"Unknown",
				supersaiyan.WT(supersaiyan.Eq("status", "u", "active"), "Active User"),
				supersaiyan.WT(supersaiyan.Eq("status", "u", "inactive"), "Inactive User"),
				supersaiyan.WT(supersaiyan.Eq("status", "u", "pending"), "Pending User"),
			)),
		).
		Where(supersaiyan.IsNotNull("email", "u")).
		Limit(10)

	sql, args, _ := qb.Select()
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Args: %v\n\n", args)
}

func coalesceExpressions() {
	fmt.Println("Example: COALESCE")
	fmt.Println("-----------------")

	qb := supersaiyan.New("mysql", "users", "u").
		WithFields(
			supersaiyan.F("id", supersaiyan.WithTable("u")),
			supersaiyan.Exp("display_name", supersaiyan.Coal(
				"Anonymous",
				supersaiyan.F("nickname", supersaiyan.WithTable("u")),
				supersaiyan.F("username", supersaiyan.WithTable("u")),
				supersaiyan.F("email", supersaiyan.WithTable("u")),
			)),
		).
		Limit(10)

	sql, args, _ := qb.Select()
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Args: %v\n\n", args)
}

func insertExample() {
	fmt.Println("Example: INSERT")
	fmt.Println("---------------")

	qb := supersaiyan.New("mysql", "users", "u")

	entry := map[string]any{
		"username": "john_doe",
		"email":    "john@example.com",
		"age":      30,
		"status":   "active",
	}

	sql, args, _ := qb.Add(entry)
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Args: %v\n\n", args)
}

func updateExample() {
	fmt.Println("Example: UPDATE")
	fmt.Println("---------------")

	qb := supersaiyan.New("mysql", "users", "u").
		Where(supersaiyan.Eq("id", "u", 123))

	entry := map[string]any{
		"username": "jane_doe",
		"email":    "jane@example.com",
		"status":   "inactive",
	}

	sql, args, _ := qb.Edit(entry)
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Args: %v\n\n", args)
}

func deleteExample() {
	fmt.Println("Example: DELETE")
	fmt.Println("---------------")

	qb := supersaiyan.New("mysql", "users", "u").
		Where(
			supersaiyan.Eq("status", "u", "deleted"),
			supersaiyan.Lt("last_login", "u", "2020-01-01"),
		)

	sql, args, _ := qb.Delete()
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Args: %v\n\n", args)
}

func jsonUnmarshal() {
	fmt.Println("Example: JSON Unmarshaling")
	fmt.Println("--------------------------")

	jsonQuery := `{
		"dialect": "postgres",
		"table": {
			"name": "users",
			"alias": "u"
		},
		"fields": [
			{"name": "id", "tableAlias": "u"},
			{"name": "username", "tableAlias": "u"}
		],
		"wheres": [
			{
				"op": "eq",
				"fieldName": "status",
				"tableAlias": "u",
				"value": "active"
			}
		],
		"sorts": [
			{
				"name": "created_at",
				"tableAlias": "u",
				"order": "DESC"
			}
		]
	}`

	var qb supersaiyan.SQLBuilder
	json.Unmarshal([]byte(jsonQuery), &qb)

	sql, args, _ := qb.Select()
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Args: %v\n\n", args)
}

func yamlUnmarshal() {
	fmt.Println("Example: YAML Unmarshaling")
	fmt.Println("--------------------------")

	yamlQuery := `
dialect: mysql
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
	yaml.Unmarshal([]byte(yamlQuery), &qb)

	sql, args, _ := qb.Select()
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Args: %v\n\n", args)
}
