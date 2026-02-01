# SuperSaiyan SQL Builder

![coverage](https://raw.githubusercontent.com/xwinata/supersaiyan/refs/heads/badges/.badges/main/coverage.svg)

A powerful, type-safe SQL query builder for Go that extends [goqu](https://github.com/doug-martin/goqu) with struct-based query construction and JSON/YAML unmarshaling support. Build complex SQL queries using Go structs and seamlessly serialize/deserialize them.

## Features

- 🔧 **Struct-Based Query Construction** - Build SQL queries using Go structs with a fluent, chainable API
- 📦 **JSON/YAML Support** - Unmarshal queries from JSON or YAML configuration files
- 🔒 **Type-Safe** - Leverage Go's type system for compile-time safety
- 🛡️ **SQL Injection Protection** - All queries use prepared statements by default
- 🎯 **Multiple Dialects** - Support for MySQL, PostgreSQL, SQLite, and SQL Server
- 🔗 **Complex Queries** - Joins, subqueries, aggregations, CASE expressions, and more
- ⚡ **Built on goqu** - Extends the battle-tested goqu library
- ✅ **Well Tested** - 74.5% test coverage with 234+ test cases

## Installation

```bash
go get github.com/xwinata/supersaiyan
```

## Quick Start

### Basic Query Construction

```go
package main

import (
    "fmt"
    "supersaiyan"
)

func main() {
    // Create a query builder
    qb := supersaiyan.New("mysql", "users", "u").
        WithFields(
            supersaiyan.F("id", supersaiyan.WithTable("u")),
            supersaiyan.F("username", supersaiyan.WithTable("u")),
            supersaiyan.F("email", supersaiyan.WithTable("u")),
        ).
        Where(
            supersaiyan.Eq("status", "u", "active"),
            supersaiyan.Gt("age", "u", 18),
        ).
        OrderBy(supersaiyan.Desc("created_at", "u")).
        Limit(10)

    // Generate SQL
    sql, args, err := qb.Select()
    if err != nil {
        panic(err)
    }

    fmt.Println(sql)
    // Output: SELECT "u"."id", "u"."username", "u"."email" 
    //         FROM "users" AS "u" 
    //         WHERE (("u"."status" = ?) AND ("u"."age" > ?)) 
    //         ORDER BY "u"."created_at" DESC 
    //         LIMIT ?
    
    fmt.Println(args)
    // Output: [active 18 10]
}
```

### Unmarshal from JSON

```go
package main

import (
    "encoding/json"
    "supersaiyan"
)

func main() {
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
        ]
    }`

    var qb supersaiyan.SQLBuilder
    json.Unmarshal([]byte(jsonQuery), &qb)

    sql, args, _ := qb.Select()
    // Use sql and args with your database
}
```

### Unmarshal from YAML

```go
package main

import (
    "supersaiyan"
    "gopkg.in/yaml.v3"
)

func main() {
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
    // Use sql and args with your database
}
```

## Core Concepts

### SQLBuilder

The main struct for building queries. Create one using `New()`:

```go
qb := supersaiyan.New(dialect, tableName, tableAlias)
```

### Chaining Methods

All methods return `*SQLBuilder` for fluent chaining:

```go
qb := supersaiyan.New("mysql", "users", "u").
    WithFields(supersaiyan.F("id", "u")).
    Where(supersaiyan.Eq("status", "u", "active")).
    OrderBy(supersaiyan.Desc("created_at", "u")).
    Limit(10).
    Offset(20)
```

## API Reference

### Query Construction

#### Creating a Builder

```go
// New(dialect, tableName, tableAlias) creates a new query builder
qb := supersaiyan.New("mysql", "users", "u")
```

Supported dialects: `mysql`, `postgres`, `sqlite3`, `sqlserver`

#### Field Helper Functions

The `F()` function creates simple field references using functional options:

```go
// Simple field without table alias
F("username")

// Field with table alias
F("username", WithTable("u"))

// Field with table alias and field alias
F("created_at", WithTable("u"), WithAlias("reg_date"))

// Field with field alias but no table alias
F("username", WithAlias("user_name"))
```

The `Exp()` function creates expression/computed fields (aggregations, CASE, COALESCE, etc.):

```go
// Aggregation
Exp("order_count", Literal{
    Value: "COUNT(?)",
    Args:  []any{F("id", "o")},
})

// CASE expression
Exp("status_label", Case{
    Conditions: []WhenThen{
        {When: Eq("status", "u", "active"), Then: "Active"},
        {When: Eq("status", "u", "inactive"), Then: "Inactive"},
    },
    Else: "Unknown",
})

// COALESCE
Exp("display_name", Coalesce{
    Fields: []Field{
        F("nickname", "u"),
        F("username", "u"),
    },
    DefaultValue: "Anonymous",
})
```

#### Field Selection

```go
// Select specific fields
qb.WithFields(
    supersaiyan.F("id", "u"),
    supersaiyan.F("username", "u"),
)

// Field with alias
qb.WithFields(
    supersaiyan.F("created_at", "u", "registration_date"),
)

// Expression field using Exp() - for computed fields
qb.WithFields(
    supersaiyan.Exp("total", supersaiyan.Literal{
        Value: "SUM(?)",
        Args:  []any{supersaiyan.F("amount", "o")},
    }),
)

// Multiple aggregations
qb.WithFields(
    supersaiyan.F("user_id", "o"),
    supersaiyan.Exp("order_count", supersaiyan.Literal{
        Value: "COUNT(?)",
        Args:  []any{supersaiyan.F("id", "o")},
    }),
    supersaiyan.Exp("total_amount", supersaiyan.Literal{
        Value: "SUM(?)",
        Args:  []any{supersaiyan.F("amount", "o")},
    }),
)
```

#### WHERE Conditions

```go
// Basic conditions
qb.Where(
    supersaiyan.Eq("status", "u", "active"),      // =
    supersaiyan.Neq("role", "u", "guest"),        // !=
    supersaiyan.Gt("age", "u", 18),               // >
    supersaiyan.Gte("score", "u", 100),           // >=
    supersaiyan.Lt("price", "p", 1000),           // <
    supersaiyan.Lte("quantity", "p", 50),         // <=
)

// Pattern matching
qb.Where(
    supersaiyan.Like("email", "u", "%@example.com"),
    supersaiyan.ILike("name", "u", "%john%"),     // Case-insensitive
)

// NULL checks
qb.Where(
    supersaiyan.IsNull("deleted_at", "u"),
    supersaiyan.IsNotNull("email", "u"),
)

// IN / NOT IN
qb.Where(
    supersaiyan.In("status", "u", []string{"active", "pending"}),
    supersaiyan.NotIn("role", "u", []string{"banned", "deleted"}),
)

// BETWEEN
qb.Where(
    supersaiyan.Between("price", "p", 100, 1000),
    supersaiyan.NotBetween("age", "u", 18, 65),
)
```

#### Logical Operators

```go
// AND conditions (default)
qb.Where(
    supersaiyan.Eq("status", "u", "active"),
    supersaiyan.Gt("age", "u", 18),
)

// OR conditions
qb.Where(supersaiyan.Or(
    supersaiyan.Eq("role", "u", "admin"),
    supersaiyan.Eq("role", "u", "moderator"),
))

// Complex nested conditions
qb.Where(
    supersaiyan.Eq("status", "u", "active"),
    supersaiyan.Or(
        supersaiyan.And(
            supersaiyan.Eq("role", "u", "admin"),
            supersaiyan.Gt("age", "u", 21),
        ),
        supersaiyan.Eq("verified", "u", true),
    ),
)
```

#### Joins

```go
// Inner join
qb.InnerJoin("orders", "o", 
    supersaiyan.Eq("user_id", "o", supersaiyan.F("id", "u")),
)

// Left join
qb.LeftJoin("profiles", "p",
    supersaiyan.Eq("user_id", "p", supersaiyan.F("id", "u")),
)

// Right join
qb.RightJoin("departments", "d",
    supersaiyan.Eq("id", "d", supersaiyan.F("department_id", "u")),
)

// Multiple join conditions
qb.InnerJoin("orders", "o",
    supersaiyan.Eq("user_id", "o", supersaiyan.F("id", "u")),
    supersaiyan.Eq("status", "o", "completed"),
)
```

#### Sorting

```go
// Ascending
qb.OrderBy(supersaiyan.Asc("username", "u"))

// Descending
qb.OrderBy(supersaiyan.Desc("created_at", "u"))

// Multiple sorts
qb.OrderBy(
    supersaiyan.Desc("created_at", "u"),
    supersaiyan.Asc("username", "u"),
)
```

#### Grouping

```go
qb.GroupByFields(
    supersaiyan.F("user_id", "o"),
    supersaiyan.F("status", "o"),
)
```

#### Pagination

```go
// Limit
qb.Limit(25)

// Offset
qb.Offset(50)

// Remove limit
qb.Limit(0)
```

### SQL Generation

#### SELECT

```go
sql, args, err := qb.Select()
// Returns: SQL string, prepared statement args, error
```

#### COUNT

```go
sql, args, err := qb.Count()
// Generates: SELECT COUNT(*) FROM ...
```

#### INSERT

```go
entry := map[string]any{
    "username": "john_doe",
    "email":    "john@example.com",
    "age":      30,
}

sql, args, err := qb.Add(entry)
// Generates: INSERT INTO users ...
```

#### UPDATE

```go
// Requires WHERE condition for safety
qb.Where(supersaiyan.Eq("id", "u", 1))

entry := map[string]any{
    "username": "jane_doe",
    "email":    "jane@example.com",
}

sql, args, err := qb.Edit(entry)
// Generates: UPDATE users SET ... WHERE ...
// Returns error if no WHERE conditions
```

#### DELETE

```go
// Requires WHERE condition for safety
qb.Where(supersaiyan.Eq("id", "u", 1))

sql, args, err := qb.Delete()
// Generates: DELETE FROM users WHERE ...
// Returns error if no WHERE conditions
```

## Advanced Features

### CASE Expressions

```go
qb.WithFields(supersaiyan.Exp("status_label", supersaiyan.Case{
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
}))
```

### COALESCE

```go
qb.WithFields(supersaiyan.Exp("display_name", supersaiyan.Coalesce{
    Fields: []supersaiyan.Field{
        supersaiyan.F("nickname", "u"),
        supersaiyan.F("username", "u"),
        supersaiyan.F("email", "u"),
    },
    DefaultValue: "Anonymous",
}))
```

### Raw SQL Literals

```go
qb.WithFields(supersaiyan.Exp("total_orders", supersaiyan.Literal{
    Value: "COUNT(DISTINCT ?)",
    Args:  []any{supersaiyan.F("id", "o")},
}))
```

### Aggregations

```go
qb := supersaiyan.New("mysql", "orders", "o").
    WithFields(
        supersaiyan.F("user_id", "o"),
        supersaiyan.Exp("order_count", supersaiyan.Literal{
            Value: "COUNT(?)",
            Args:  []any{supersaiyan.F("id", "o")},
        }),
        supersaiyan.Exp("total_amount", supersaiyan.Literal{
            Value: "SUM(?)",
            Args:  []any{supersaiyan.F("amount", "o")},
        }),
    ).
    Where(supersaiyan.Eq("status", "o", "completed")).
    GroupByFields(supersaiyan.F("user_id", "o")).
    OrderBy(supersaiyan.Desc("total_amount", ""))
```

## JSON/YAML Schema

### Complete Example

```json
{
  "dialect": "mysql",
  "table": {
    "name": "users",
    "alias": "u",
    "relations": [
      {
        "joinType": "INNER",
        "table": {
          "name": "orders",
          "alias": "o"
        },
        "on": [
          {
            "op": "eq",
            "fieldName": "user_id",
            "tableAlias": "o",
            "value": {
              "name": "id",
              "tableAlias": "u"
            }
          }
        ]
      }
    ]
  },
  "fields": [
    {
      "name": "id",
      "tableAlias": "u"
    },
    {
      "name": "username",
      "tableAlias": "u"
    },
    {
      "fieldAlias": "order_count",
      "exp": {
        "value": "COUNT(?)",
        "args": [
          {
            "name": "id",
            "tableAlias": "o"
          }
        ]
      }
    }
  ],
  "wheres": [
    {
      "op": "eq",
      "fieldName": "status",
      "tableAlias": "u",
      "value": "active"
    },
    {
      "op": "OR",
      "conditions": [
        {
          "op": "like",
          "fieldName": "email",
          "tableAlias": "u",
          "value": "%@gmail.com"
        },
        {
          "op": "like",
          "fieldName": "email",
          "tableAlias": "u",
          "value": "%@yahoo.com"
        }
      ]
    }
  ],
  "sorts": [
    {
      "name": "created_at",
      "tableAlias": "u",
      "order": "DESC"
    }
  ],
  "groupBy": [
    {
      "name": "id",
      "tableAlias": "u"
    }
  ]
}
```

### Field Types

#### Boolean Operations
- `eq` - Equal (=)
- `neq` - Not equal (!=)
- `gt` - Greater than (>)
- `gte` - Greater than or equal (>=)
- `lt` - Less than (<)
- `lte` - Less than or equal (<=)
- `in` - IN
- `notIn` - NOT IN
- `like` - LIKE
- `notLike` - NOT LIKE
- `iLike` - Case-insensitive LIKE
- `notILike` - Case-insensitive NOT LIKE
- `is` - IS (for NULL)
- `isNot` - IS NOT (for NULL)

#### Range Operations
- `between` - BETWEEN
- `notBetween` - NOT BETWEEN

#### Logical Operations
- `AND` - AND group
- `OR` - OR group

#### Join Types
- `INNER` - INNER JOIN
- `LEFT` - LEFT JOIN
- `RIGHT` - RIGHT JOIN

#### Sort Directions
- `ASC` - Ascending
- `DESC` - Descending

## Complete Example

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    "supersaiyan"
    
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    // Open database connection
    db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/dbname")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Build complex query
    qb := supersaiyan.New("mysql", "users", "u").
        WithFields(
            supersaiyan.F("id", "u"),
            supersaiyan.F("username", "u"),
            supersaiyan.F("email", "u"),
            supersaiyan.Field{
                FieldAlias: "order_count",
                Exp: supersaiyan.Literal{
                    Value: "COUNT(?)",
                    Args:  []any{supersaiyan.F("id", "o")},
                },
            },
        ).
        InnerJoin("orders", "o",
            supersaiyan.Eq("user_id", "o", supersaiyan.F("id", "u")),
        ).
        Where(
            supersaiyan.Eq("status", "u", "active"),
            supersaiyan.Gt("age", "u", 18),
        ).
        GroupByFields(supersaiyan.F("id", "u")).
        OrderBy(supersaiyan.Desc("order_count", "")).
        Limit(10)

    // Generate SQL
    sql, args, err := qb.Select()
    if err != nil {
        log.Fatal(err)
    }

    // Execute query
    rows, err := db.Query(sql, args...)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    // Process results
    for rows.Next() {
        var id int
        var username, email string
        var orderCount int
        
        err := rows.Scan(&id, &username, &email, &orderCount)
        if err != nil {
            log.Fatal(err)
        }
        
        fmt.Printf("User: %s (%s) - Orders: %d\n", username, email, orderCount)
    }
}
```

## Safety Features

### Prepared Statements

All queries use prepared statements by default to prevent SQL injection:

```go
sql, args, err := qb.Select()
// sql contains placeholders: SELECT * FROM users WHERE status = ?
// args contains values: ["active"]
```

### Required WHERE for Mutations

`Edit()` and `Delete()` require WHERE conditions to prevent accidental data loss:

```go
qb := supersaiyan.New("mysql", "users", "u")

// This will return an error
_, _, err := qb.Delete()
// Error: WHERE condition is required for Edit and Delete operations

// This is safe
qb.Where(supersaiyan.Eq("id", "u", 1))
_, _, err = qb.Delete()
// OK
```

## Testing

Run the test suite:

```bash
# Run all tests
go test ./tests/...

# Run with coverage
go test ./tests/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific test
go test ./tests/... -run TestSelect
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Add your license here]

## Credits

Built on top of [goqu](https://github.com/doug-martin/goqu) by Doug Martin.
