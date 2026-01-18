# Code Snippets & Usage Guide

This document provides quick reference for all the code snippets used in this database handler package.

---

## Table of Contents

1. [Database Schema Synchronization](#1-database-schema-synchronization)
2. [Query Builder Examples](#2-query-builder-examples)
3. [Component Management](#3-component-management)
4. [Model Definition Examples](#4-model-definition-examples)
5. [Error Handling Patterns](#5-error-handling-patterns)
6. [Performance Tips](#6-performance-tips)

---

## 1. Database Schema Synchronization

### Getting the Database Name

```go
// Get the current database name from sql.DB connection
var dbName string
if err := m.QueryRow("SELECT DATABASE()").Scan(&dbName); err != nil {
    panic("Error getting database name: " + err.Error())
}
fmt.Printf("Currently connected to database: %s\n", dbName)
```

**Usage**: Used internally to fetch index information and validate schema operations.

### Syncing Table Schema

```go
// Load current database schema for comparison
func (m *meta) SyncModelSchema() {
    // Get the active database connection
    if err := m.Ping(); err != nil {
        panic("Database not reachable: " + err.Error())
    }

    // Check if the table for this model actually exists in the database
    checkQuery := `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?`
    var count int
    
    if err := m.QueryRow(checkQuery, m.TableName).Scan(&count); err != nil {
        panic("Error checking table existence: " + err.Error())
    }
    
    if count == 0 {
        fmt.Printf("Table '%s' does not exist.\n", m.TableName)
        return
    }

    // Query the structure of the existing table
    rows, err := m.Query("SHOW COLUMNS FROM `" + m.TableName + "`")
    if err != nil {
        panic("Error getting old table structure: " + err.Error())
    }
    defer rows.Close()

    // Clear any previously cached schema info
    m.schemas = nil

    // Process each column of the table
    for rows.Next() {
        schema := schema{}
        if err := rows.Scan(&schema.field, &schema.fieldType, &schema.nullable, &schema.key, &schema.defaultVal, &schema.extra); err != nil {
            panic("Error scanning row: " + err.Error())
        }
        m.schemas = append(m.schemas, schema)
    }
}
```

**Usage**: Run this when you want to compare your model definitions with the actual database schema.

### Automatic Table Schema Modification

```go
// Automatically synchronize table schema
// This detects differences between model and database and prompts for changes
func (m *meta) syncTableSchema() {
    // This function will:
    // 1. Add missing fields
    // 2. Modify existing fields that have changed
    // 3. Sync index properties (UNIQUE, PRIMARY KEY, INDEX)
    // 4. Remove fields that no longer exist in the model
}
```

**Usage**: Called during initialization with `--migrate-model` flag or manually.

---

## 2. Query Builder Examples

### SELECT Operations

#### Get All Records

```go
// Fetch all users
users, err := Users.Get().Fetch()
if err != nil {
    log.Fatal(err)
}

for _, user := range users {
    fmt.Printf("User: %v\n", user)
}
```

#### Get with WHERE Condition

```go
// Get users where status is 'active'
activeUsers, err := Users.Get().
    Where("status").Is("active").
    Fetch()
```

#### Get with Multiple Conditions (AND)

```go
// Get active users with age greater than 18
results, err := Users.Get().
    Where("status").Is("active").
    And().Where("age").GreaterThan(18).
    Fetch()
```

#### Get with OR Conditions

```go
// Get users named 'John' OR 'Jane'
results, err := Users.Get().
    Where("firstName").Is("John").
    Or().Where("firstName").Is("Jane").
    Fetch()
```

#### Get with Pattern Matching

```go
// Get users whose email contains 'example.com'
results, err := Users.Get().
    Where("email").Like("%example.com").
    Fetch()

// Get users whose firstName starts with 'A'
results, err := Users.Get().
    Where("firstName").Like("A%").
    Fetch()
```

#### Get with IN Operator

```go
// Get users with specific IDs
userIDs := []string{"u1", "u2", "u3"}
results, err := Users.Get().
    Where("userId").In(userIDs...).
    Fetch()
```

#### Get with BETWEEN

```go
// Get users with createdAt in a date range
results, err := Users.Get().
    Where("createdAt").Between("2024-01-01", "2024-12-31").
    Fetch()
```

#### Get Single Record

```go
// Get the first user with userName = 'alice'
user, err := Users.Get().
    Where("userName").Is("alice").
    First()

if err == sql.ErrNoRows {
    fmt.Println("User not found")
} else if err != nil {
    log.Fatal(err)
}
```

#### Get with Sorting

```go
// Get users sorted by creation date (descending)
users, err := Users.Get().
    OrderBy("createdAt DESC").
    Fetch()

// Get users sorted by name (ascending)
users, err := Users.Get().
    OrderBy("userName ASC").
    Fetch()
```

#### Get with Pagination

```go
// Get 20 users starting from offset 40
results, err := Users.Get().
    Limit(20).
    Offset(40).
    Fetch()

// Using Page helper (1-indexed)
page := 2
pageSize := 20
results, err := Users.Get().
    Page(page, pageSize).
    Fetch()
```

#### Complex Query Example

```go
// Get the first 10 active users, sorted by creation date
users, err := Users.Get().
    Where("status").Is("active").
    OrderBy("createdAt DESC").
    Limit(10).
    Fetch()
```

### INSERT Operations

#### Create Single Record

```go
// Create a new user
err := Users.Create().
    Set("userId").To("u123").
    Set("userName").To("alice").
    Set("email").To("alice@example.com").
    Set("password").To("securepass").
    Set("firstName").To("Alice").
    Exec()

if err != nil {
    log.Fatal(err)
}
```

#### Create Multiple Records

```go
// Insert multiple users
users := []map[string]interface{}{
    {"userId": "u1", "userName": "user1", "email": "user1@example.com"},
    {"userId": "u2", "userName": "user2", "email": "user2@example.com"},
    {"userId": "u3", "userName": "user3", "email": "user3@example.com"},
}

for _, userData := range users {
    err := Users.Create().
        Set("userId").To(userData["userId"]).
        Set("userName").To(userData["userName"]).
        Set("email").To(userData["email"]).
        Exec()
    
    if err != nil {
        log.Printf("Error creating user: %v", err)
    }
}
```

### UPDATE Operations

#### Update Single Record

```go
// Update password for specific user
err := Users.Get().
    Where("userId").Is("u123").
    Set("password").To("newpass").
    Exec()
```

#### Update Multiple Records

```go
// Update all inactive users to archived status
err := Users.Get().
    Where("status").Is("inactive").
    Set("status").To("archived").
    Exec()
```

#### Update with Complex Condition

```go
// Update status to archived for pending users created before 2024-01-01
err := Users.Get().
    Where("status").Is("pending").
    And().Where("createdAt").LessThan("2024-01-01").
    Set("status").To("archived").
    Exec()
```

### DELETE Operations

#### Delete Single Record

```go
// Delete a specific user
err := Users.Get().
    Where("userId").Is("u123").
    Delete()
```

#### Delete Multiple Records

```go
// Delete all inactive users
err := Users.Get().
    Where("status").Is("inactive").
    Delete()

// Delete users matching multiple conditions
err := Users.Get().
    Where("status").Is("inactive").
    And().Where("lastLogin").LessThan("2023-01-01").
    Delete()
```

---

## 3. Component Management

### Define a Component Struct

```go
type Setting struct {
    Key   string `json:"Key"`
    Value string `json:"Value"`
}

type Feature struct {
    Id      string `json:"Id"`
    Name    string `json:"Name"`
    Enabled bool   `json:"Enabled"`
    Beta    bool   `json:"Beta"`
}
```

### Register a Component

```go
import "github.com/vrianta/agai/v1/model"

var SettingsComponent = model.ComponentOf[Setting, string](
    SettingsModel,  // Your model definition
    "Key",          // Primary key field name
)

var FeaturesComponent = model.ComponentOf[Feature, string](
    FeaturesModel,
    "Id",
)
```

### Initialize Components

```go
// In your main() or init() function
func init() {
    // Make sure database is set up first
    // model.SetDB(db)
    
    // Then initialize components
    model.InitializeComponents()
}
```

### Access Component Data

```go
// Get single component
setting := SettingsComponent.Val["site_name"]
fmt.Println(setting.Value)

// Check if component exists
if setting, exists := SettingsComponent.Val["theme"]; exists {
    fmt.Println("Theme:", setting.Value)
}

// Iterate all components
for key, setting := range SettingsComponent.Val {
    fmt.Printf("%s = %s\n", key, setting.Value)
}

// Count components
count := len(SettingsComponent.Val)
fmt.Printf("Total settings: %d\n", count)
```

### Reload Components from Disk

```go
// Reload all components from JSON files
model.ReloadComponents()
```

### Save Components to Disk

```go
// Save all components to JSON files
model.DumpComponentsToJSON()

// Save specific component
model.DumpComponentToJSON("settings", SettingsComponent.Val)
```

### Create JSON Component File

**File: `./components/settings.component.json`**
```json
{
  "site_name": {
    "Key": "site_name",
    "Value": "My Application"
  },
  "theme_color": {
    "Key": "theme_color",
    "Value": "#007bff"
  },
  "enable_notifications": {
    "Key": "enable_notifications",
    "Value": "true"
  }
}
```

---

## 4. Model Definition Examples

### Simple User Model

```go
var Users = model.New(
    "users",
    map[string]model.Field{
        "userId": {
            Name:     "userId",
            Type:     model.FieldTypes.VarChar,
            Length:   20,
            Nullable: false,
            Index: model.Index{
                PrimaryKey: true,
            },
        },
        "userName": {
            Name:     "userName",
            Type:     model.FieldTypes.VarChar,
            Length:   30,
            Nullable: false,
            Index: model.Index{
                Unique: true,
            },
        },
        "email": {
            Name:     "email",
            Type:     model.FieldTypes.VarChar,
            Length:   100,
            Nullable: false,
        },
    },
)
```

### Complex Model with Timestamps

```go
var Orders = model.New(
    "orders",
    map[string]model.Field{
        "orderId": {
            Name:     "orderId",
            Type:     model.FieldTypes.Int,
            Nullable: false,
            Index: model.Index{
                PrimaryKey: true,
            },
            AutoIncrement: true,
        },
        "userId": {
            Name:     "userId",
            Type:     model.FieldTypes.VarChar,
            Length:   20,
            Nullable: false,
        },
        "totalAmount": {
            Name:     "totalAmount",
            Type:     model.FieldTypes.Decimal,
            Nullable: false,
        },
        "status": {
            Name:         "status",
            Type:         model.FieldTypes.VarChar,
            Length:       20,
            DefaultValue: "pending",
            Nullable:     false,
        },
        "createdAt": {
            Name:         "createdAt",
            Type:         model.FieldTypes.Timestamp,
            DefaultValue: "CURRENT_TIMESTAMP",
            Nullable:     false,
        },
        "updatedAt": {
            Name:         "updatedAt",
            Type:         model.FieldTypes.Timestamp,
            DefaultValue: "CURRENT_TIMESTAMP",
            Nullable:     false,
        },
    },
)
```

### Model with Multiple Indexes

```go
var Products = model.New(
    "products",
    map[string]model.Field{
        "id": {
            Name:     "id",
            Type:     model.FieldTypes.VarChar,
            Length:   50,
            Nullable: false,
            Index: model.Index{
                PrimaryKey: true,
            },
        },
        "sku": {
            Name:     "sku",
            Type:     model.FieldTypes.VarChar,
            Length:   50,
            Nullable: false,
            Index: model.Index{
                Unique: true,
            },
        },
        "name": {
            Name:     "name",
            Type:     model.FieldTypes.VarChar,
            Length:   255,
            Nullable: false,
            Index: model.Index{
                Index: true,  // Regular index
            },
        },
        "category": {
            Name:     "category",
            Type:     model.FieldTypes.VarChar,
            Length:   100,
            Nullable: true,
            Index: model.Index{
                Index: true,
            },
        },
    },
)
```

---

## 5. Error Handling Patterns

### Basic Error Handling

```go
// Create with error checking
err := Users.Create().
    Set("userId").To("u123").
    Set("userName").To("alice").
    Exec()

if err != nil {
    log.Printf("Failed to create user: %v", err)
    return err
}
```

### Handle No Rows Found

```go
user, err := Users.Get().
    Where("userId").Is("u123").
    First()

if err == sql.ErrNoRows {
    fmt.Println("User not found")
    return nil
} else if err != nil {
    log.Printf("Database error: %v", err)
    return err
}

// User exists, use it
fmt.Printf("Found user: %v\n", user)
```

### Get or Create Pattern

```go
user, err := Users.Get().
    Where("userId").Is("u123").
    First()

if err == sql.ErrNoRows {
    // User doesn't exist, create it
    err := Users.Create().
        Set("userId").To("u123").
        Set("userName").To("alice").
        Set("email").To("alice@example.com").
        Exec()
    
    if err != nil {
        log.Printf("Failed to create user: %v", err)
        return err
    }
    
    fmt.Println("User created successfully")
} else if err != nil {
    log.Printf("Database error: %v", err)
    return err
} else {
    fmt.Printf("User already exists: %v\n", user)
}
```

### Transactional Error Handling

```go
// Create a new user with validation
err := Users.Create().
    Set("userId").To(userID).
    Set("userName").To(userName).
    Set("email").To(email).
    Exec()

if err != nil {
    if strings.Contains(err.Error(), "Duplicate entry") {
        fmt.Println("User already exists")
    } else if strings.Contains(err.Error(), "foreign key") {
        fmt.Println("Referenced user not found")
    } else {
        fmt.Printf("Unexpected error: %v\n", err)
    }
    return err
}
```

---

## 6. Performance Tips

### Use Limit for Large Datasets

```go
// Good - limits memory usage
results, err := Users.Get().
    Limit(100).
    Offset(0).
    Fetch()

// Bad - loads all records into memory
allUsers, err := Users.Get().Fetch()
```

### Use Indexes for Frequently Queried Columns

```go
// Define indexes on columns you frequently filter by
"email": {
    Name:   "email",
    Type:   model.FieldTypes.VarChar,
    Length: 100,
    Index: model.Index{
        Unique: true,  // Email should be unique
    },
},
"status": {
    Name:   "status",
    Type:   model.FieldTypes.VarChar,
    Length: 20,
    Index: model.Index{
        Index: true,   // Regular index for filtering
    },
},
```

### Pagination for User Interfaces

```go
// Instead of fetching all users
// results, _ := Users.Get().Fetch()

// Use pagination
const pageSize = 20
pageNum := 1

results, err := Users.Get().
    Page(pageNum, pageSize).
    OrderBy("createdAt DESC").
    Fetch()

if err != nil {
    log.Fatal(err)
}
```

### Batch Operations Instead of Individual Inserts

```go
// Collect data first, then insert
// This is better for performance than separate insert calls
for _, data := range largeDataSet {
    // Process data...
    Users.Create().
        Set("userId").To(data.UserID).
        // ... other fields ...
        Exec()
}

// Or consider using database transactions if available
```

### Use Specific Columns

```go
// Instead of fetching entire records and filtering in code
// Push filtering to database where it's faster
activeUsers, _ := Users.Get().
    Where("status").Is("active").
    Fetch()

// This is better than:
allUsers, _ := Users.Get().Fetch()
// Then manually checking status in code
```

---

## Tips & Tricks

### Type Conversions

```go
// String to Integer
maxUsersStr := SettingsComponent.Val["max_users"].Value
maxUsers, err := strconv.Atoi(maxUsersStr)

// String to Boolean
debugStr := SettingsComponent.Val["debug_mode"].Value
debugMode := strings.ToLower(debugStr) == "true"

// Integer to String
userID := 123
stringID := strconv.Itoa(userID)
```

### Counting Records

```go
// Count all records
all, _ := Users.Get().Fetch()
totalCount := len(all)

// Count matching records
active, _ := Users.Get().Where("status").Is("active").Fetch()
activeCount := len(active)
```

### Formatted Output

```go
// Pretty print results
results, _ := Users.Get().Fetch()
for _, user := range results {
    fmt.Printf("ID: %-10s | Name: %-20s | Email: %s\n",
        user.ID, user.Name, user.Email)
}
```

---

This guide covers the main code patterns and snippets. For more information, refer to `readme.md` and `component_readme.md`.
