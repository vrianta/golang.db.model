package model

import "database/sql"

type (
	fieldType    uint16
	FieldTypeset map[string]*Field
	Result       map[string]any
	Results      map[any]Result

	Field struct {
		// user do not have to pass the name of the field it will automatically populate the name
		name          string
		Type          fieldType
		Length        int
		Nullable      bool
		Definition    []any // Used for ENUM types, e.g., []any{"value1", "value2"}
		DefaultValue  string
		AutoIncrement bool
		Index         Index // Index type (e.g., "UNIQUE", "INDEX")

		// table name
		table_name string

		fk *foreignKey // unexported foreign key metadata
	}

	foreignKey struct {
		referenceTable  string
		referenceColumn string
		onDelete        string
		onUpdate        string
	}

	Index struct {
		PrimaryKey bool
		Unique     bool
		Index      bool
		FullText   bool
		Spatial    bool
	}

	schema struct {
		field      string
		fieldType  string
		nullable   string
		key        string
		extra      string
		defaultVal sql.NullString

		// Add these for precise index detection (from `information_schema.statistics`)
		// indexName string
		isunique  bool
		isindex   bool
		isprimary bool
	}

	// InsertRowBuilder is a dedicated struct for InsertRow operations (CREATE), separate from the general queryBuilder struct.
	InsertRowBuilder struct {
		model               *meta
		InsertRowFieldTypes map[string]any
		lastSet             string
	}
)
