# ModelsHandler: Go ORM-like Query Builder

ModelsHandler is a human-friendly, chainable query builder for working with database tables using Go structs. It allows you to easily create models, build queries, and interact with your database in a readable, maintainable way—similar to popular Object-Relational Mappers (ORMs).

The package provides automatic database schema synchronization, making it simple to manage database migrations alongside your code.

---

## Table of Contents
- [Features](#features)
- [Installation & Setup](#installation--setup)
- [1. Defining a Model](#1-defining-a-model)
- [2. Understanding the Model Structure](#2-understanding-the-model-structure)
- [3. Initializing Models](#3-initializing-models)
- [4. Building and Executing Queries](#4-building-and-executing-queries)
  - [Creating Records (INSERT)](#creating-records-insert)
  - [Fetching Data (SELECT)](#fetching-data-select)
  - [Fetching a Single Row](#fetching-a-single-row)
  - [Updating Data (UPDATE)](#updating-data-update)
  - [Deleting Data (DELETE)](#deleting-data-delete)
- [5. Query Builder API Reference](#5-query-builder-api-reference)
- [6. Schema Synchronization](#6-schema-synchronization)
- [7. Advanced Features](#7-advanced-features)
- [8. Best Practices](#8-best-practices)
- [9. Contributing](#9-contributing)
- [10. License](#10-license)

---

## Features

- **Chainable, fluent API** for building queries intuitively
- **Supports SELECT, CREATE, UPDATE, DELETE operations** with a familiar ORM-like syntax
- **Rich WHERE conditions**: AND, OR, IN, NOT IN, BETWEEN, LIKE, IS NULL, comparison operators, etc.
- **Pagination & Sorting**: LIMIT, OFFSET, ORDER BY, GROUP BY
- **Automatic Schema Synchronization**: Compare model definitions with database schema and auto-migrate
- **Type-safe field definitions** with validation
- **Easy to extend and understand** with clear, documented code
- **Automatic Database Migration**: When the `Build` flag is set to `false`, ModelsHandler automatically migrates your database schema to match your model definitions

> **Note:** Migration only happens if the `Build` flag is `false`. In development, set `Build: false` to enable auto-migration. In production, set it to `true` to prevent accidental schema changes.

---

## Installation & Setup

1. **Import the package** in your Go project:
   ```go
   import "github.com/vrianta/golang/model"
   ```

2. **Initialize your database connection** before using models:
   ```go
   import (
       _ "github.com/go-sql-driver/mysql"
       "github.com/vrianta/golang/model"
       
   )
   ```

3. **Define your models** (see section 1 below)

---

## 1. Defining a Model

To use ModelsHandler, define your model as a struct with pointer fields. For example, to create a `Users` model:

```go
package models

import "github.com/vrianta/golang/model"

var Users = model.New(db, "users", struct {
    UserId    *model.Field
    UserName  *model.Field
    Password  *model.Field
    FirstName *model.Field
    LastName  *model.Field
}{
    UserId: &model.Field{
        Type:     model.FieldTypes.VarChar,
        Length:   100,
        Nullable: false,
        Index: model.Index{
            PrimaryKey: true,
            Unique:     false,
            Index:      true,
        },
    },
    UserName: &model.Field{
        Type:     model.FieldTypes.VarChar,
        Length:   30,
        Nullable: false,
        Index: model.Index{
            Unique: true,
            Index:  true,
        },
    },
    Password: &model.Field{
        Type:     model.FieldTypes.Text,
        Nullable: false,
    },
    FirstName: &model.Field{
        Type:     model.FieldTypes.VarChar,
        Length:   20,
        Nullable: false,
    },
    LastName: &model.Field{
        Type:     model.FieldTypes.VarChar,
        Length:   20,
        Nullable: false,
    },
})
```

### Field Options

- **Type** (FieldType): Data type (VarChar, Int, Text, Timestamp, etc.)
- **Length** (int): For VARCHAR types, the maximum length (0 = unspecified)
- **Nullable** (bool): Whether NULL values are allowed
- **DefaultValue** (string): Default value for the column
- **AutoIncrement** (bool): Auto-increment the column
- **Index** (Index struct): Configure PRIMARY KEY, UNIQUE, or regular INDEX

### Supported Field Types

- `VarChar`, `Text` - String types
- `Int`, `BigInt`, `SmallInt` - Integer types
- `Decimal`, `Float` - Decimal types
- `Boolean` - Boolean type
- `Date`, `DateTime`, `Timestamp` - Temporal types
- `JSON` - JSON type

### Model Field Access

After defining your model, you can access fields using the `.Fields` property:

```go
Users.Fields.UserId      // Access UserId field
Users.Fields.UserName    // Access UserName field
Users.Fields.Password    // Access Password field
```

---

## 2. Understanding the Model Structure

The model system works as follows:

1. **Model Definition**: You define a model with `model.New()`, passing:
   - Table name (string)
   - A struct with pointer fields to `model.Field`

2. **Field Access**: After definition, access fields via `ModelName.Fields.FieldName`

3. **Type-Safe Queries**: Use field references in queries instead of string column names:
   ```go
   // Good - Type-safe with field references
   Users.Get().Where(Users.Fields.UserId).Is("u123").First()
   
   // The field reference includes metadata about the column
   ```

4. **Query Results**: Results are returned as `map[string]interface{}` where:
   - Keys are column names from the database
   - Values are the data

---

## 3. Initializing Models

Before you can use your models, you need to initialize them. The initialization process:

1. Creates the table if it doesn't exist
2. Compares your model definitions with the database schema
3. Optionally synchronizes any schema differences

```go
func init() {
    Users.Initialize()  // Creates table and syncs schema if needed
}
```

**Automatic Migration**: If you run your application with the `--migrate-model` or `-mm` flag, the system will automatically sync the database schema with your model definitions:

```bash
go run main.go --migrate-model
```

Or:
```bash
go run main.go -mm
```

---

## 4. Building and Executing Queries

### Creating Records (INSERT)

```go
// Create a new user
err := Users.Create().
    Set(Users.Fields.UserId).To("u123").
    Set(Users.Fields.UserName).To("alice").
    Set(Users.Fields.Password).To("securepass").
    Set(Users.Fields.FirstName).To("Alice").
    Set(Users.Fields.LastName).To("Smith").
    Exec()

if err != nil {
    log.Fatal(err)
}
```

### Fetching Data (SELECT)

```go
// Get all users
users, err := Users.Get().Fetch()

// Get users with a specific userName
results, err := Users.Get().
    Where(Users.Fields.UserName).Is("alice").
    Fetch()

// Get multiple users by UserId
results, err := Users.Get().
    Where(Users.Fields.UserId).In("u1", "u2", "u3").
    Fetch()
```

### Fetching a Single Row

```go
// Get the first user with userName = 'alice'
user, err := Users.Get().
    Where(Users.Fields.UserName).Is("alice").
    First()

if err == sql.ErrNoRows {
    fmt.Println("User not found")
} else if err != nil {
    log.Fatal(err)
}
```

### Updating Data (UPDATE)

```go
// Update a specific user's password
err := Users.Get().
    Where(Users.Fields.UserId).Is("u123").
    Set(Users.Fields.Password).To("newpass").
    Exec()

// Update multiple fields
err := Users.Get().
    Where(Users.Fields.UserId).Is("u123").
    Set(Users.Fields.FirstName).To("John").
    Set(Users.Fields.LastName).To("Doe").
    Exec()

// Using Update method (pass nil if no initial field)
err := Users.Update(nil).
    Where(Users.Fields.UserId).Is("u123").
    Set(Users.Fields.Password).To("updated_pass").
    Exec()
```

### Deleting Data (DELETE)

```go
// Delete a specific user
err := Users.Get().
    Where(Users.Fields.UserId).Is("u123").
    Delete()

// Delete users matching a condition
err := Users.Get().
    Where(Users.Fields.UserName).Is("inactive_user").
    Delete()
```

---

## 5. Query Builder API Reference

### Query Initiation

- `.Get()` — Start a new SELECT query (chain with WHERE, ORDER BY, etc.)
- `.Create()` — Start a new INSERT query
- `.Update(field)` — Start a new UPDATE query (pass nil or a field reference)

### WHERE Conditions

- `.Where(field)` — Add a WHERE clause for the specified field (use `Model.Fields.FieldName`)
- `.Is(value)` — WHERE field = value
- `.IsNot(value)` — WHERE field != value
- `.GreaterThan(value)` — WHERE field > value
- `.LessThan(value)` — WHERE field < value
- `.GreaterThanOrEqual(value)` — WHERE field >= value
- `.LessThanOrEqual(value)` — WHERE field <= value
- `.Like(pattern)` — WHERE field LIKE pattern (SQL pattern matching)
- `.In(values...)` — WHERE field IN (value1, value2, ...)
- `.NotIn(values...)` — WHERE field NOT IN (...)
- `.Between(min, max)` — WHERE field BETWEEN min AND max
- `.IsNull()` — WHERE field IS NULL
- `.IsNotNull()` — WHERE field IS NOT NULL

### Combining Conditions

- `.And()` — Add AND operator for next condition
- `.Or()` — Add OR operator for next condition

### Sorting & Grouping

- `.OrderBy(clause)` — ORDER BY clause (e.g., "name ASC", "createdAt DESC")
- `.GroupBy(clause)` — GROUP BY clause

### Pagination

- `.Limit(n)` — Limit to n results
- `.Offset(n)` — Skip first n results
- `.Page(page, pageSize)` — Helper for pagination (1-indexed page number)

### Setting Values (for INSERT/UPDATE)

- `.Set(field).To(value)` — Set field value (use `Model.Fields.FieldName` for field)
- `.Exec()` — Execute the query

### Execution

- `.Fetch()` — Execute SELECT and return all results as slice
- `.First()` — Execute SELECT and return first result
- `.Exec()` — Execute INSERT or UPDATE
- `.Delete()` — Execute DELETE

---

## 6. Schema Synchronization

### What is Schema Synchronization?

Schema synchronization automatically compares your Go model definitions with your database schema and applies necessary changes. This allows you to evolve your database structure alongside your code without writing manual migration scripts.

### How It Works

The `SyncModelSchema()` and `syncTableSchema()` functions detect and fix:

1. **Add New Fields**: Detects fields in your model that don't exist in the database and creates them
2. **Modify Existing Fields**: Identifies type mismatches, length changes, nullable/auto-increment differences
3. **Sync Indexes**: Ensures PRIMARY KEY, UNIQUE, and INDEX definitions match
4. **Remove Unused Fields**: Detects fields in the database that no longer exist in your model (with confirmation)

### Using Schema Synchronization

**Enable auto-sync on startup**:

```go
func init() {
    Users.SyncModelSchema()  // Load and compare current database schema
    Users.Initialize()       // Create table if not exists
}
```

**Run migration manually**:

```bash
# Add the --migrate-model flag to sync all models
go run main.go --migrate-model
```

### What Changes Are Detected?

The system detects and can fix:
- **Type mismatches**: e.g., INT vs VARCHAR
- **Length changes**: VARCHAR(50) → VARCHAR(100)
- **Nullable constraints**: Column was NOT NULL now needs to be NULL
- **Default values**: Different default value assignments
- **Auto-increment**: Field should be auto-incrementing but isn't
- **Index changes**: UNIQUE, PRIMARY KEY, or INDEX properties

### Example: Schema Evolution

**Original model**:
```go
"email": {
    Name:     "email",
    Type:     model.FieldTypes.VarChar,
    Length:   50,
    Nullable: false,
}
```

**Updated model** (length changed):
```go
"email": {
    Name:     "email",
    Type:     model.FieldTypes.VarChar,
    Length:   100,  // Changed from 50 to 100
    Nullable: false,
}
```

When you run with `--migrate-model`, the system automatically alters the column:
```sql
ALTER TABLE users MODIFY COLUMN email VARCHAR(100) NOT NULL;
```

---

## 7. Advanced Features

### Pagination

Use `Limit()`, `Offset()`, or `Page()` for efficient pagination:

```go
// Get 20 users starting from offset 40
results, err := Users.Get().Limit(20).Offset(40).Fetch()

// Or use the Page() helper (page 1-indexed)
results, err := Users.Get().Page(3, 20).Fetch()  // Page 3 with 20 items per page
```

### Complex WHERE Conditions

Combine multiple conditions with `And()` and `Or()`:

```go
// Get users with specific userName AND UserId
results, err := Users.Get().
    Where(Users.Fields.UserName).Is("alice").
    And().Where(Users.Fields.UserId).Is("u123").
    Fetch()

// Get users named 'Alice' OR 'Bob'
results, err := Users.Get().
    Where(Users.Fields.FirstName).Is("Alice").
    Or().Where(Users.Fields.FirstName).Is("Bob").
    Fetch()
```

### Pattern Matching with LIKE

```go
// Get all users whose firstName starts with 'A'
results, err := Users.Get().
    Where(Users.Fields.FirstName).Like("A%").
    Fetch()

// Get users whose userName contains 'john'
results, err := Users.Get().
    Where(Users.Fields.UserName).Like("%john%").
    Fetch()
```

### Using IN and BETWEEN

```go
// Get users with specific IDs
results, err := Users.Get().
    Where(Users.Fields.UserId).In("u1", "u2", "u3").
    Fetch()

// Get users NOT in a list
results, err := Users.Get().
    Where(Users.Fields.UserId).NotIn("u1", "u2").
    Fetch()
```

### Batch Operations

Create or update multiple records efficiently:

```go
// Insert multiple users
userData := []struct {
    UserId    string
    UserName  string
    Password  string
    FirstName string
    LastName  string
}{
    {"u1", "user1", "pass1", "John", "Doe"},
    {"u2", "user2", "pass2", "Jane", "Smith"},
    {"u3", "user3", "pass3", "Bob", "Johnson"},
}

for _, data := range userData {
    err := Users.Create().
        Set(Users.Fields.UserId).To(data.UserId).
        Set(Users.Fields.UserName).To(data.UserName).
        Set(Users.Fields.Password).To(data.Password).
        Set(Users.Fields.FirstName).To(data.FirstName).
        Set(Users.Fields.LastName).To(data.LastName).
        Exec()
    
    if err != nil {
        log.Printf("Error creating user: %v", err)
    }
}
```

### Conditional Updates

```go
// Update multiple users matching a condition
err := Users.Get().
    Where(Users.Fields.UserName).Is("oldname").
    Set(Users.Fields.FirstName).To("NewFirstName").
    Set(Users.Fields.LastName).To("NewLastName").
    Exec()

// Complex conditional update
query := Users.Update(nil).
    Where(Users.Fields.UserId).Is("u123")

query.
    Set(Users.Fields.Password).To("newpassword").
    Set(Users.Fields.FirstName).To("UpdatedFirst").
    Set(Users.Fields.LastName).To("UpdatedLast")

if err := query.Exec(); err != nil {
    log.Fatal(err)
}
```

---

## 8. Best Practices

### Error Handling

Always check error returns after database operations:

```go
user, err := Users.Get().Where("userId").Is("u123").First()
if err == sql.ErrNoRows {
    fmt.Println("User not found")
} else if err != nil {
    log.Fatal(err)
}
```

### Connection Pooling

The underlying `sql.DB` handles connection pooling automatically. Share the same `sql.DB` instance across goroutines—it's thread-safe:

```go
// Do this (good)
var db *sql.DB

func init() {
    db, _ = sql.Open("mysql", "...")
}

// Don't do this (bad)
// Create a new connection for each operation
```

### Performance

- Use `Limit()` for large result sets to avoid loading entire tables into memory
- Use appropriate indexes on frequently queried columns
- Use pagination for large datasets

### Schema Safety

- Use `--migrate-model` in development to auto-sync schema
- Set `Build: true` in production to prevent accidental schema changes
- Always review schema migrations before applying to production databases

### Type Safety

Leverage Go's type system by using strongly-typed field definitions:

```go
// Good - type-safe
"age": {
    Name:     "age",
    Type:     model.FieldTypes.Int,
    Nullable: false,
}

// Avoid - less safe
"metadata": {
    Type: model.FieldTypes.Text,  // JSON in string?
}
```

### Common Patterns

**Get or create**:
```go
user, err := Users.Get().Where(Users.Fields.UserId).Is("u123").First()
if err == sql.ErrNoRows {
    // User doesn't exist, create it
    err := Users.Create().
        Set(Users.Fields.UserId).To("u123").
        Set(Users.Fields.UserName).To("alice").
        Exec()
}
```

**Update with WHERE clause**:
```go
err := Users.Get().
    Where(Users.Fields.UserName).Is("oldname").
    Set(Users.Fields.FirstName).To("newname").
    Exec()
```

**Update multiple fields**:
```go
query := Users.Update(nil).
    Where(Users.Fields.UserId).Is("u123")

query.
    Set(Users.Fields.FirstName).To("John").
    Set(Users.Fields.LastName).To("Doe").
    Set(Users.Fields.Password).To("newpass")

if err := query.Exec(); err != nil {
    log.Fatal(err)
}
```

**Count records**:
```go
results, _ := Users.Get().Fetch()
count := len(results)
```

**Delete with condition**:
```go
err := Users.Get().
    Where(Users.Fields.UserId).Is("u123").
    Delete()
```

**Access retrieved data**:
```go
user, err := Users.Get().Where(Users.Fields.UserId).Is("u123").First()
if err != nil {
    log.Fatal(err)
}

// user is a map[string]interface{} with all field values
userName := user[Users.Fields.UserName.Name]
firstName := user[Users.Fields.FirstName.Name]
```

---

## 9. Contributing

Pull requests and suggestions are welcome! Please:
- Document your code with clear comments
- Keep the API intuitive and chainable
- Add tests for new features
- Follow Go conventions and idioms

---

## 10. License

MIT
