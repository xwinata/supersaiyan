# SuperSaiyan SQL Builder

![coverage](https://raw.githubusercontent.com/xwinata/supersaiyan/refs/heads/badges/.badges/main/coverage.svg)

A powerful, type-safe SQL query builder for Go that extends [goqu](https://github.com/doug-martin/goqu) with struct-based query construction and JSON/YAML unmarshaling support.

## Features

- 🔧 **Struct-Based Query Construction** - Build SQL queries using Go structs with a fluent, chainable API
- 📦 **JSON/YAML Support** - Unmarshal queries from JSON or YAML configuration files
- 🔒 **Type-Safe** - Leverage Go's type system for compile-time safety
- 🛡️ **SQL Injection Protection** - All queries use prepared statements by default
- 🎯 **Multiple Dialects** - Support for MySQL, PostgreSQL, SQLite, and SQL Server
- 🔗 **Complex Queries** - Joins, subqueries, aggregations, CASE expressions, and more
- ⚡ **Built on goqu** - Extends the battle-tested goqu library

## Installation

```bash
go get github.com/xwinata/supersaiyan
```

## Quick Start

```go
package main

import (
    "fmt"
    "supersaiyan"
)

func main() {
    qb := supersaiyan.New("mysql", "users", "u").
        WithFields(
            supersaiyan.F("id", supersaiyan.WithTable("u")),
            supersaiyan.F("username", supersaiyan.WithTable("u")),
        ).
        Where(supersaiyan.Eq("status", "u", "active")).
        OrderBy(supersaiyan.Desc("created_at", "u")).
        Limit(10)

    sql, args, err := qb.Select()
    if err != nil {
        panic(err)
    }

    fmt.Println(sql)
    // SELECT "u"."id", "u"."username" FROM "users" AS "u" 
    // WHERE ("u"."status" = ?) ORDER BY "u"."created_at" DESC LIMIT ?
    
    fmt.Println(args)
    // [active 10]
}
```

## API Reference

### Creating a Query Builder

```go
qb := supersaiyan.New(dialect, tableName, tableAlias)
```

Supported dialects: `mysql`, `postgres`, `sqlite3`, `sqlserver`

### Helper Functions

#### `F()` - Field References

```go
F("username")                                    // Simple field
F("username", WithTable("u"))                    // With table alias
F("created_at", WithTable("u"), WithAlias("reg_date"))  // With field alias
```

#### `L()` - Literal Expressions

```go
L("COUNT(*)")                                    // Simple literal
L("COUNT(?)", F("id", WithTable("o")))          // With field argument
L("CONCAT(?, ' ', ?)", "Hello", "World")        // With multiple args
```

#### `Exp()` - Expression Fields

```go
Exp("order_count", L("COUNT(?)", F("id", WithTable("o"))))
```

#### `Coal()` - COALESCE

```go
Coal(nil, F("nickname", WithTable("u")), F("username", WithTable("u")))
Coal("Anonymous", F("nickname", WithTable("u")), F("username", WithTable("u")))
```

#### `C()` and `WT()` - CASE Expressions

```go
C("Unknown",
    WT(Eq("status", "u", "active"), "Active"),
    WT(Eq("status", "u", "inactive"), "Inactive"),
)
```

### Query Building

#### Fields

```go
qb.WithFields(
    F("id", WithTable("u")),
    F("username", WithTable("u")),
    F("email", WithTable("u"), WithAlias("contact_email")),
)
```

#### Conditions

```go
// Basic conditions
qb.Where(
    Eq("status", "u", "active"),      // =
    Neq("role", "u", "guest"),        // !=
    Gt("age", "u", 18),               // >
    Gte("score", "u", 100),           // >=
    Lt("price", "p", 1000),           // <
    Lte("quantity", "p", 50),         // <=
)

// Pattern matching
qb.Where(
    Like("email", "u", "%@example.com"),
    ILike("name", "u", "%john%"),     // Case-insensitive
)

// NULL checks
qb.Where(
    IsNull("deleted_at", "u"),
    IsNotNull("email", "u"),
)

// IN / NOT IN
qb.Where(
    In("status", "u", []string{"active", "pending"}),
    NotIn("role", "u", []string{"banned", "deleted"}),
)

// BETWEEN
qb.Where(
    Between("price", "p", 100, 1000),
    NotBetween("age", "u", 18, 65),
)

// Logical operators
qb.Where(
    Or(
        Eq("role", "u", "admin"),
        Eq("role", "u", "moderator"),
    ),
)

qb.Where(
    And(
        Eq("status", "u", "active"),
        Or(
            Gt("age", "u", 18),
            Eq("verified", "u", true),
        ),
    ),
)
```

#### Joins

```go
qb.InnerJoin("orders", "o", Eq("user_id", "o", F("id", WithTable("u"))))
qb.LeftJoin("profiles", "p", Eq("user_id", "p", F("id", WithTable("u"))))
qb.RightJoin("departments", "d", Eq("id", "d", F("department_id", WithTable("u"))))
```

#### Sorting

```go
qb.OrderBy(
    Desc("created_at", "u"),
    Asc("username", "u"),
)
```

#### Grouping

```go
qb.GroupByFields(
    F("user_id", WithTable("o")),
    F("status", WithTable("o")),
)
```

#### Pagination

```go
qb.Limit(10).Offset(20)
```

### SQL Generation

```go
// SELECT
sql, args, err := qb.Select()

// COUNT
sql, args, err := qb.Count()

// INSERT
sql, args, err := qb.Add(map[string]any{
    "username": "john_doe",
    "email":    "john@example.com",
})

// UPDATE (requires WHERE)
qb.Where(Eq("id", "u", 123))
sql, args, err := qb.Edit(map[string]any{
    "username": "jane_doe",
})

// DELETE (requires WHERE)
qb.Where(Eq("id", "u", 123))
sql, args, err := qb.Delete()
```

## Advanced Features

### CASE Expressions

```go
qb.WithFields(
    F("id", WithTable("u")),
    Exp("status_label", C(
        "Unknown",
        WT(Eq("status", "u", "active"), "Active User"),
        WT(Eq("status", "u", "inactive"), "Inactive User"),
    )),
)
```

### COALESCE

```go
qb.WithFields(
    F("id", WithTable("u")),
    Exp("display_name", Coal(
        "Anonymous",
        F("nickname", WithTable("u")),
        F("username", WithTable("u")),
        F("email", WithTable("u")),
    )),
)
```

### Aggregations

```go
qb.WithFields(
    F("user_id", WithTable("o")),
    Exp("order_count", L("COUNT(?)", F("id", WithTable("o")))),
    Exp("total_amount", L("SUM(?)", F("amount", WithTable("o")))),
).GroupByFields(F("user_id", WithTable("o")))
```

## JSON/YAML Unmarshaling

SuperSaiyan supports unmarshaling queries from JSON and YAML:

```go
// From JSON
var qb supersaiyan.SQLBuilder
json.Unmarshal([]byte(jsonQuery), &qb)

// From YAML
yaml.Unmarshal([]byte(yamlQuery), &qb)
```

See [examples/](examples/) for complete JSON/YAML examples.

## Safety Features

### Prepared Statements

All queries use prepared statements by default to prevent SQL injection.

### Required WHERE for Mutations

`Edit()` and `Delete()` require WHERE conditions to prevent accidental data loss:

```go
qb := supersaiyan.New("mysql", "users", "u")

// This will return an error
_, _, err := qb.Delete()
// Error: WHERE condition is required for Edit and Delete operations

// This is safe
qb.Where(Eq("id", "u", 1))
_, _, err = qb.Delete()
// OK
```

## Examples

See the [examples/](examples/) directory for comprehensive examples:

- [examples.go](examples/examples.go) - Complete examples covering all features including field helpers, expressions, CASE, COALESCE, joins, aggregations, and more

## Credits

Built on top of [goqu](https://github.com/doug-martin/goqu) by Doug Martin.
