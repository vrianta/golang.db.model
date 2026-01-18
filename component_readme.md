# Component Package Documentation

The `component` package provides a hybrid file+database approach for managing application components (configuration, static data, or dynamic content). Component data is stored as JSON files in the `./components/` directory and can also be synced with the database. This enables type-safe, ergonomic, and persistent management of configuration data.

---

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [1. Component File Structure](#1-component-file-structure)
- [2. Defining Your Component Model](#2-defining-your-component-model)
- [3. Registering a Component](#3-registering-a-component)
- [4. Using Components](#4-using-components)
- [5. Syncing Components](#5-syncing-components)
- [6. API Reference](#6-api-reference)
- [7. Best Practices](#7-best-practices)
- [8. Complete Example](#8-complete-example)
- [9. Troubleshooting](#9-troubleshooting)

---

## Features

- **File-based storage**: Each component is stored as a JSON file (`table_name.component.json`) in `./components/`
- **Type-safe access**: Data is loaded into Go structs and accessible as `map[PrimaryKey]YourStruct`
- **Automatic initialization**: On startup, loads from JSON, syncs with DB if needed, or updates JSON from DB
- **Thread-safe**: Uses mutexes for concurrent access
- **Centralized registration and initialization**
- **Hot reload**: Reload all components from disk at runtime
- **Optional write-back**: Dump in-memory data back to JSON files
- **Hybrid approach**: Works with both file and database, choosing the most appropriate source
- **Schema consistency**: Database components are synced with your model definitions

---

## Quick Start

### 1. Create JSON Component Files

Create a `./components/` directory at your project root and add JSON files:

**File: `./components/settings.component.json`**
```json
{
  "site_name": {
    "Key": "site_name",
    "Value": "My Awesome Site"
  },
  "theme": {
    "Key": "theme",
    "Value": "dark"
  },
  "max_users": {
    "Key": "max_users",
    "Value": "100"
  }
}
```

### 2. Define Your Struct

```go
type Setting struct {
    Key   string
    Value string
}
```

### 3. Define Your Model

```go
var SettingsModel = model.New("settings", map[string]model.Field{
    "Key": {
        Name:     "Key",
        Type:     model.FieldTypes.VarChar,
        Length:   100,
        Nullable: false,
        Index: model.Index{
            PrimaryKey: true,
        },
    },
    "Value": {
        Name:     "Value",
        Type:     model.FieldTypes.Text,
        Nullable: true,
    },
})
```

### 4. Register Component

```go
import "github.com/vrianta/agai/v1/model"

var SettingsComponent = model.ComponentOf[Setting, string](
    SettingsModel,  // your model
    "Key",          // primary key field name
)
```

### 5. Initialize All Components

```go
func init() {
    model.InitializeComponents()
}
```

### 6. Access Data

```go
// Access settings
siteName := SettingsComponent.Val["site_name"].Value
theme := SettingsComponent.Val["theme"].Value
```

---

## 1. Component File Structure

Component files are stored in the `./components/` directory with the naming convention: `{table_name}.component.json`

### File Format

```json
{
  "primary_key_1": {
    "field1": "value1",
    "field2": "value2",
    // ... all fields from your struct
  },
  "primary_key_2": {
    "field1": "value1",
    "field2": "value2"
  }
}
```

### Example: Settings Component

**File: `components/settings.component.json`**
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
  },
  "api_timeout": {
    "Key": "api_timeout",
    "Value": "30"
  }
}
```

### Example: Features Component

**File: `components/features.component.json`**
```json
{
  "feature_auth": {
    "Id": "feature_auth",
    "Name": "Authentication",
    "Enabled": true,
    "Beta": false
  },
  "feature_api": {
    "Id": "feature_api",
    "Name": "REST API",
    "Enabled": true,
    "Beta": true
  },
  "feature_analytics": {
    "Id": "feature_analytics",
    "Name": "Analytics Dashboard",
    "Enabled": false,
    "Beta": true
  }
}
```

---

## 2. Defining Your Component Model

Define a Go struct that matches your component's data structure:

### Simple Key-Value Component

```go
type Setting struct {
    Key   string `json:"Key"`
    Value string `json:"Value"`
}
```

### Complex Component

```go
type Feature struct {
    Id      string    `json:"Id"`
    Name    string    `json:"Name"`
    Enabled bool      `json:"Enabled"`
    Beta    bool      `json:"Beta"`
    Created time.Time `json:"Created"`
}
```

### Multi-field Component

```go
type Permission struct {
    Id          int    `json:"Id"`
    Name        string `json:"Name"`
    Description string `json:"Description"`
    Resource    string `json:"Resource"`
    Action      string `json:"Action"`
    CreatedAt   string `json:"CreatedAt"`
}
```

### Best Practices for Struct Definition

1. **Use struct tags**: Add `json` tags to match your JSON file structure
2. **Match field names**: Ensure struct fields match the JSON keys exactly (case-sensitive)
3. **Use appropriate types**: Use `string`, `int`, `bool`, `time.Time`, etc.
4. **Handle nullable fields**: Use pointers for optional fields:

```go
type User struct {
    Id       int     `json:"Id"`
    Name     string  `json:"Name"`
    Email    *string `json:"Email"`     // Optional
    LastSeen *int64  `json:"LastSeen"`  // Optional timestamp
}
```

---

## 3. Registering a Component

### Define Your Model

First, create your model definition:

```go
package models

import "github.com/vrianta/agai/v1/model"

var SettingsModel = model.New("settings", map[string]model.Field{
    "Key": {
        Name:     "Key",
        Type:     model.FieldTypes.VarChar,
        Length:   100,
        Nullable: false,
        Index: model.Index{
            PrimaryKey: true,
        },
    },
    "Value": {
        Name:     "Value",
        Type:     model.FieldTypes.Text,
        Nullable: true,
    },
})
```

### Register the Component

```go
package components

import (
    "github.com/vrianta/agai/v1/model"
    "myapp/models"
)

type Setting struct {
    Key   string `json:"Key"`
    Value string `json:"Value"`
}

// Register the component
// Generic parameters: [StructType, PrimaryKeyType]
var SettingsComponent = model.ComponentOf[Setting, string](
    models.SettingsModel,  // Your model definition
    "Key",                 // Primary key field name
)
```

### Type Parameters Explained

```go
model.ComponentOf[Setting, string](...)
                 //     ^       ^
                 //     |       +-- Type of primary key (string, int, etc.)
                 //     +-- Your struct type
```

---

## 4. Using Components

### Accessing Component Data

```go
// Get a single item by key
setting := SettingsComponent.Val["site_name"]
fmt.Println(setting.Value)  // Output: "My Awesome Site"

// Iterate over all components
for key, setting := range SettingsComponent.Val {
    fmt.Printf("%s = %s\n", key, setting.Value)
}
```

### Type-Safe Access

```go
type Feature struct {
    Id      string
    Name    string
    Enabled bool
}

var FeaturesComponent = model.ComponentOf[Feature, string](FeaturesModel, "Id")

// Access with type safety
if FeaturesComponent.Val["auth"].Enabled {
    // Authentication feature is enabled
}
```

### Checking If Component Exists

```go
if setting, exists := SettingsComponent.Val["theme"]; exists {
    fmt.Println("Theme:", setting.Value)
} else {
    fmt.Println("Theme setting not found")
}
```

### Working with Numbers

```go
// Component stored as string in JSON
maxUsersStr := SettingsComponent.Val["max_users"].Value

// Convert to integer if needed
maxUsers, err := strconv.Atoi(maxUsersStr)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Max users: %d\n", maxUsers)
```

### Filtering Components

```go
// Get all enabled features
enabledFeatures := make(map[string]Feature)
for key, feature := range FeaturesComponent.Val {
    if feature.Enabled {
        enabledFeatures[key] = feature
    }
}
```

---

## 5. Syncing Components

### How Syncing Works

The component system employs a smart sync strategy:

1. **On Startup**:
   - Load all JSON files from `./components/` directory
   - Check if the database table exists and has data
   - If DB is empty → insert JSON data as defaults
   - If DB has data → load from DB and update JSON file
   - If no JSON file → load from DB only

2. **Syncing Process**:
   - JSON is the source of truth for defaults/configuration
   - Database is the source of truth for current state
   - Changes in DB that aren't in JSON are added
   - Items in JSON that aren't in DB are added
   - Removed items are tracked

### Initialize Components

```go
// Initialize all registered components
// This loads from JSON and syncs with database
func init() {
    model.InitializeComponents()
}
```

### Manual Sync Operations

```go
// Reload all components from disk (JSON files)
model.ReloadComponents()

// Save current in-memory state to JSON files
model.DumpComponentsToJSON()

// Sync a specific component with database
SettingsComponent.SyncWithDatabase()
```

### Selective Dumping

```go
// Save only settings component to JSON
model.DumpComponentToJSON("settings", SettingsComponent.Val)
```

---

## 6. API Reference

### Component Registration

```go
model.ComponentOf[T, K](modelDef *model.Table[any], primaryKeyField string) *Component[T, K]
```

- `T`: Your component struct type
- `K`: Type of primary key field
- `modelDef`: Model definition (must match your struct)
- `primaryKeyField`: Name of the primary key field in your struct

### Component Struct

```go
type Component[T any, K comparable] struct {
    Val map[K]T         // In-memory data
    Mu  sync.RWMutex    // Thread safety
    // ... internal fields
}
```

### Methods

**Access data**:
```go
// Thread-safe read access
item := component.Val["key"]

// Check existence
if item, exists := component.Val["key"]; exists {
    // Use item
}
```

**Iterate**:
```go
for key, item := range component.Val {
    // Process each component
}
```

**Count**:
```go
count := len(component.Val)
```

### Module-Level Functions

```go
// Initialize all components
model.InitializeComponents()

// Reload from disk
model.ReloadComponents()

// Save to disk
model.DumpComponentsToJSON()

// Save specific component
model.DumpComponentToJSON(tableName string, data map[K]T)
```

---

## 7. Best Practices

### 1. Directory Structure

```
project-root/
├── components/
│   ├── settings.component.json
│   ├── features.component.json
│   └── permissions.component.json
├── models/
│   ├── settings.go
│   ├── features.go
│   └── permissions.go
├── components/
│   ├── settings.go
│   ├── features.go
│   └── permissions.go
└── main.go
```

### 2. Initialization Order

Always initialize components in your `main()` or early in `init()`:

```go
func main() {
    // 1. Set up database
    db, _ := sql.Open("mysql", "...")
    model.SetDB(db)
    
    // 2. Initialize all models
    models.Initialize()
    
    // 3. Initialize components AFTER database is ready
    component.InitializeComponents()
    
    // 4. Now you can use components
    startServer()
}
```

### 3. Consistent Struct Tags

Always use `json` tags and ensure they match your JSON files:

```go
type Setting struct {
    Key   string `json:"Key"`      // Matches JSON key
    Value string `json:"Value"`    // Matches JSON key
}
```

### 4. Error Handling

```go
// Check if component exists before accessing
if feature, exists := FeaturesComponent.Val["auth"]; exists {
    if feature.Enabled {
        // Use feature
    }
} else {
    log.Printf("Feature 'auth' not found in components")
}
```

### 5. Thread Safety

The component system uses `sync.RWMutex` internally. When updating components at runtime:

```go
// If manually modifying component data, use locks
SettingsComponent.Mu.Lock()
SettingsComponent.Val["newKey"] = Setting{...}
SettingsComponent.Mu.Unlock()

// Then sync to database and JSON
SettingsComponent.SyncWithDatabase()
model.DumpComponentToJSON("settings", SettingsComponent.Val)
```

### 6. JSON File Management

- Keep JSON files in version control for defaults
- Use `ReloadComponents()` for hot-reloading during development
- Use `DumpComponentsToJSON()` when modifying components at runtime
- Validate JSON format before adding to `./components/`

### 7. Type Conversions

```go
// Component values may be strings in JSON, convert as needed
apiTimeout := SettingsComponent.Val["api_timeout"].Value
timeout, _ := strconv.Atoi(apiTimeout)

// Or store as proper types in struct if possible
type NumericSetting struct {
    Key   string `json:"Key"`
    Value int    `json:"Value"`  // Properly typed
}
```

---

## 8. Complete Example

### Project Structure

```
myapp/
├── components/
│   └── settings.component.json
├── models/
│   └── settings.go
├── components/
│   └── settings.go
└── main.go
```

### Step 1: Create JSON File

**`components/settings.component.json`**:
```json
{
  "app_name": {
    "Key": "app_name",
    "Value": "MyApp"
  },
  "app_version": {
    "Key": "app_version",
    "Value": "1.0.0"
  },
  "debug_mode": {
    "Key": "debug_mode",
    "Value": "false"
  },
  "max_connections": {
    "Key": "max_connections",
    "Value": "100"
  }
}
```

### Step 2: Define Model

**`models/settings.go`**:
```go
package models

import "github.com/vrianta/agai/v1/model"

var SettingsModel = model.New("settings", map[string]model.Field{
    "Key": {
        Name:     "Key",
        Type:     model.FieldTypes.VarChar,
        Length:   100,
        Nullable: false,
        Index: model.Index{
            PrimaryKey: true,
        },
    },
    "Value": {
        Name:     "Value",
        Type:     model.FieldTypes.Text,
        Nullable: true,
    },
})
```

### Step 3: Register Component

**`components/settings.go`**:
```go
package components

import (
    "github.com/vrianta/agai/v1/model"
    "myapp/models"
)

type Setting struct {
    Key   string `json:"Key"`
    Value string `json:"Value"`
}

var SettingsComponent = model.ComponentOf[Setting, string](
    models.SettingsModel,
    "Key",
)
```

### Step 4: Use in Main

**`main.go`**:
```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    
    "github.com/vrianta/agai/v1/model"
    "myapp/components"
    "myapp/models"
    
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    // Setup database
    db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/mydb")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    model.SetDB(db)
    models.SettingsModel.Initialize()
    
    // Initialize components
    model.InitializeComponents()
    
    // Use components
    appName := components.SettingsComponent.Val["app_name"].Value
    appVersion := components.SettingsComponent.Val["app_version"].Value
    
    fmt.Printf("Welcome to %s v%s\n", appName, appVersion)
    
    // List all settings
    fmt.Println("\nAll Settings:")
    for key, setting := range components.SettingsComponent.Val {
        fmt.Printf("  %s = %s\n", key, setting.Value)
    }
}
```

---

## 9. Troubleshooting

### Components Directory Not Found

**Issue**: Warning about missing `./components/` directory

**Solution**: Create the directory at your project root:
```bash
mkdir components
```

### JSON File Not Loading

**Issue**: Component has no data even though JSON file exists

**Possible causes**:
- JSON syntax error (validate with `jq` or online JSON validator)
- File name doesn't match table name (should be `table_name.component.json`)
- File is in wrong directory (should be in `./components/`)

**Solution**:
```bash
# Validate JSON
jq . components/settings.component.json

# Check file naming
ls -la components/
```

### Data Not Syncing to Database

**Issue**: Components loaded from JSON but not appearing in database

**Solution**: Check if model table exists:
```go
models.SettingsModel.CreateTableIfNotExists()
```

### Type Mismatch Errors

**Issue**: "cannot unmarshal X into Y"

**Cause**: Struct field types don't match JSON data types

**Solution**: Ensure your struct matches the JSON:
```go
type Setting struct {
    Key   string `json:"Key"`      // Must be string if JSON has string
    Value string `json:"Value"`    // Must match JSON type
}
```

### Changes Not Persisting

**Issue**: Modified components don't save to JSON

**Solution**: Call dump after modifications:
```go
// Modify
components.SettingsComponent.Val["new_key"] = Setting{...}

// Save to JSON
model.DumpComponentToJSON("settings", components.SettingsComponent.Val)
```

### Thread Safety Issues

**Issue**: Race conditions when accessing components

**Solution**: Use provided locks when modifying:
```go
components.SettingsComponent.Mu.Lock()
components.SettingsComponent.Val["key"] = value
components.SettingsComponent.Mu.Unlock()
```

---

## Migration Notes

- The legacy DB-backed logic is still supported
- If no JSON file is present, the system loads from the database only
- If the database has data, it takes precedence over JSON (ensures current state is used)
- For new components, simply add a JSON file and register as shown above

---

For more advanced usage, see the main ModelsHandler documentation or source code with inline comments.
