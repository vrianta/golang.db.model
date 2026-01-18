package model

import "database/sql"

type (
	fieldType    uint16
	FieldTypeset map[string]*Field
	Result       map[string]any
	Results      map[any]Result

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
