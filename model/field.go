package model

/*
CREATE TABLE IF NOT EXISTS employees (
	id INT AUTO_INCREMENT,
	name VARCHAR(100),
	position VARCHAR(50),
	salary DECIMAL(10, 2),
	hire_date DATE,
	PRIMARY KEY `id` (id),
	INDEX `idx_name` (name)
);
*/

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type (
	index struct {
		PrimaryKey bool
		Unique     bool
		Index      bool
		FullText   bool
		Spatial    bool
	}

	Field struct {
		// user do not have to pass the name of the field it will automatically populate the name
		name          string
		t             fieldType //type of the field
		lenth         int
		nullable      bool
		definition    []any // Used for ENUM types, e.g., []any{"value1", "value2"}
		defaultValue  string
		autoIncrement bool
		index         index // Index type (e.g., "UNIQUE", "INDEX")

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
)

func CreateField() *Field {
	return &Field{
		nullable: true,
	}
}

// ---------- Numeric ----------

func (f *Field) AsTinyInt() *Field   { f.t = FieldTypes.TinyInt; return f }
func (f *Field) AsSmallInt() *Field  { f.t = FieldTypes.SmallInt; return f }
func (f *Field) AsMediumInt() *Field { f.t = FieldTypes.MediumInt; return f }
func (f *Field) AsInt() *Field       { f.t = FieldTypes.Int; return f }
func (f *Field) AsBigInt() *Field    { f.t = FieldTypes.BigInt; return f }

func (f *Field) AsFloat() *Field  { f.t = FieldTypes.Float; return f }
func (f *Field) AsDouble() *Field { f.t = FieldTypes.Double; return f }
func (f *Field) AsReal() *Field   { f.t = FieldTypes.Real; return f }

func (f *Field) AsDecimal(precision int) *Field {
	f.t = FieldTypes.Decimal
	f.lenth = precision
	return f
}

// ---------- String / Text ----------

func (f *Field) AsChar(n int) *Field {
	f.t = FieldTypes.Char
	f.lenth = n
	return f
}

func (f *Field) AsVarchar(n int) *Field {
	f.t = FieldTypes.VarChar
	f.lenth = n
	return f
}

func (f *Field) AsTinyText() *Field   { f.t = FieldTypes.TinyText; return f }
func (f *Field) AsText() *Field       { f.t = FieldTypes.Text; return f }
func (f *Field) AsMediumText() *Field { f.t = FieldTypes.MediumText; return f }
func (f *Field) AsLongText() *Field   { f.t = FieldTypes.LongText; return f }

// ---------- Binary / Blob ----------

func (f *Field) AsBlob() *Field       { f.t = FieldTypes.Blob; return f }
func (f *Field) AsTinyBlob() *Field   { f.t = FieldTypes.TinyBlob; return f }
func (f *Field) AsMediumBlob() *Field { f.t = FieldTypes.MediumBlob; return f }
func (f *Field) AsLongBlob() *Field   { f.t = FieldTypes.LongBlob; return f }

// ---------- Date & Time ----------

func (f *Field) AsDate() *Field {
	f.t = FieldTypes.Date
	return f
}

func (f *Field) AsTime() *Field {
	f.t = FieldTypes.Time
	return f
}

func (f *Field) AsTimestamp() *Field {
	f.t = FieldTypes.Timestamp
	return f
}

func (f *Field) AsYear() *Field {
	f.t = FieldTypes.Year
	return f
}

// ---------- JSON / ENUM / SET ----------

func (f *Field) AsJSON() *Field {
	f.t = FieldTypes.JSON
	return f
}

func (f *Field) AsEnum(values ...any) *Field {
	f.t = FieldTypes.Enum
	f.definition = values
	return f
}

func (f *Field) AsSet(values ...any) *Field {
	f.t = FieldTypes.Set
	f.definition = values
	return f
}

// ---------- Geometry ----------

func (f *Field) AsGeometry() *Field {
	f.t = FieldTypes.Geometry
	return f
}

func (f *Field) AsPoint() *Field {
	f.t = FieldTypes.Point
	return f
}

func (f *Field) AsLineString() *Field {
	f.t = FieldTypes.LineString
	return f
}

func (f *Field) AsPolygon() *Field {
	f.t = FieldTypes.Polygon
	return f
}

// ---------- Misc ----------

func (f *Field) AsUUID() *Field {
	f.t = FieldTypes.UUID
	return f
}

func (f *Field) AsBool() *Field {
	f.t = FieldTypes.Bool
	return f
}

func (f *Field) NotNull() *Field {
	f.nullable = false
	return f
}

func (f *Field) Default(value string) *Field {
	f.defaultValue = value
	return f
}

func (f *Field) DefaultNull() *Field {
	f.defaultValue = "NULL"
	return f
}

func (f *Field) DefaultNow() *Field {
	f.defaultValue = "CURRENT_TIMESTAMP"
	return f
}

func (f *Field) IsPrimary() *Field {
	f.index.PrimaryKey = true
	return f
}

func (f *Field) IsUnique() *Field {
	f.index.Unique = true
	return f
}

func (f *Field) IsIndex() *Field {
	f.index.Index = true
	return f
}

func (f *Field) columnDefinition() string {
	var response string

	// ENUM support
	if f.t == FieldTypes.Enum {

		enumValues := make([]string, len(f.definition))
		for i, val := range f.definition {
			enumValues[i] = "'" + fmt.Sprintf("%v", val) + "'"
		}
		response = fmt.Sprintf("%s ENUM(%s)", f.name, strings.Join(enumValues, ","))
	} else {
		response = f.name + " " + f.t.string()

		// if the length is greater than 0 then we are setting the length of the field
		// this is mostly used for VARCHAR, CHAR, TEXT, etc.
		if f.lenth > 0 {
			response += "(" + fmt.Sprint(f.lenth) + ")"
		}
	}

	// if f.lenth > 0 {
	// 	response += "(" + fmt.Sprint(f.lenth) + ") "
	// }

	if f.nullable {
		response += " NULL "
	} else if !f.nullable {
		response += " NOT NULL "
	}

	if f.defaultValue != "" {
		switch f.t {
		case FieldTypes.String, FieldTypes.Text:
			response += "DEFAULT '" + f.defaultValue + "' "
		case FieldTypes.Bool:
			if f.defaultValue == "true" || f.defaultValue == "1" {
				response += "DEFAULT TRUE "
			} else {
				response += "DEFAULT FALSE "
			}
		case FieldTypes.Enum:
			response += "DEFAULT '" + f.defaultValue + "' "
		default:
			response += "DEFAULT " + f.defaultValue + " "
		}
	}

	// AUTO_INCREMENT support
	if f.autoIncrement {
		response += "AUTO_INCREMENT "
	}

	return strings.TrimSpace(response)
}

/*
 * @return - Array of Index Statements
 */
func (f *Field) createIndexStatements() string {
	responseArray := []string{}
	if f.index.PrimaryKey {
		responseArray = append(responseArray, "Primary Key pk_"+f.table_name+"_"+string(f.name)+" ("+string(f.name)+")")
	}
	if f.index.Index {
		responseArray = append(responseArray, "INDEX idx_"+f.table_name+"_"+string(f.name)+" ("+string(f.name)+")")
	}
	if f.index.FullText {
		responseArray = append(responseArray, "FULLTEXT ftxt_"+f.table_name+"_"+string(f.name)+" ("+string(f.name)+")")

	}
	if f.index.Spatial {
		responseArray = append(responseArray, "SPATIAL sp_"+f.table_name+"_"+string(f.name)+" ("+string(f.name)+")")
	}
	if f.index.Unique {
		responseArray = append(responseArray, "UNIQUE unq_"+f.table_name+"_"+string(f.name)+" ("+string(f.name)+")")
	}

	if f.fk != nil {
		responseArray = append(responseArray, f.foreignKeyConstraint())
	}

	return strings.Join(responseArray, ",\n")
}

// Index Statements with ADD in it
func (f *Field) addIndexStatement() string {
	responseArray := []string{}
	if f.index.PrimaryKey {
		responseArray = append(responseArray, "Primary Key pk_"+f.table_name+"_"+string(f.name)+" ("+string(f.name)+")")
	}
	if f.index.Index {
		responseArray = append(responseArray, "INDEX idx_"+f.table_name+"_"+string(f.name)+" ("+string(f.name)+")")
	}
	if f.index.FullText {
		responseArray = append(responseArray, "FULLTEXT ftxt_"+f.table_name+"_"+string(f.name)+" ("+string(f.name)+")")

	}
	if f.index.Spatial {
		responseArray = append(responseArray, "SPATIAL sp_"+f.table_name+"_"+string(f.name)+" ("+string(f.name)+")")
	}
	if f.index.Unique {
		responseArray = append(responseArray, "UNIQUE unq_"+f.table_name+"_"+string(f.name)+" ("+string(f.name)+")")
	}

	if f.fk != nil {
		responseArray = append(responseArray, f.foreignKeyConstraint())
	}

	if len(responseArray) > 0 {
		return ", ADD " + strings.Join(responseArray, ", ADD \n")
	}

	return ""
}

func (f *Field) Name() string {
	return f.name
}

// check if the value is compatible with the field type
func (ft fieldType) IsValueCompatible(val string) bool {
	switch ft {
	case FieldTypes.Int, FieldTypes.TinyInt, FieldTypes.SmallInt, FieldTypes.MediumInt, FieldTypes.BigInt:
		_, err := strconv.Atoi(val)
		return err == nil
	case FieldTypes.Float, FieldTypes.Decimal, FieldTypes.Double, FieldTypes.Real:
		_, err := strconv.ParseFloat(val, 64)
		return err == nil
	case FieldTypes.Date, FieldTypes.Time:
		_, err1 := time.Parse("2006-01-02", val)
		_, err2 := time.Parse("2006-01-02 15:04:05", val)
		return err1 == nil || err2 == nil
	case FieldTypes.Timestamp:
		_, err := time.Parse("2006-01-02 15:04:05", val)
		return err == nil || val == "CURRENT_TIMESTAMP" || val == "NOW()"
	case FieldTypes.JSON:
		var js json.RawMessage
		return json.Unmarshal([]byte(val), &js) == nil
	case FieldTypes.VarChar, FieldTypes.Text, FieldTypes.String, FieldTypes.Char:
		return true
	case FieldTypes.Bool:
		return val == "0" || val == "1" || val == "true" || val == "false"
	default:
		return true
	}
}

func isAlphaNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func (f *Field) Compare(fieldTypeStr string) bool {

	// edge case for enum if is returning then it will return all the default values also have to make a condition to split functions
	_type := strings.Split(fieldTypeStr, "(")[0] // so if any type return TYPE() then it will return TYPE

	switch _type {
	case "TINYINT":
		switch f.t {
		case FieldTypes.TinyInt, FieldTypes.Bool:
			return true
		}
	case "SMALLINT":
		switch f.t {
		case FieldTypes.SmallInt:
			return true
		}
	case "MEDIUMINT":
		switch f.t {
		case FieldTypes.MediumInt:
			return true
		}
	case "INT", "INTEGER":
		switch f.t {
		case FieldTypes.Int:
			return true
		}
	case "BIGINT":
		switch f.t {
		case FieldTypes.BigInt:
			return true
		}
	case "VARCHAR":
		switch f.t {
		case FieldTypes.VarChar, FieldTypes.String:
			return true
		}
	case "CHAR":
		switch f.t {
		case FieldTypes.Char, FieldTypes.String:
			return true
		}
	case "TEXT":
		switch f.t {
		case FieldTypes.Text, FieldTypes.String:
			return true
		}
	case "LONGTEXT":
		switch f.t {
		case FieldTypes.LongText, FieldTypes.JSON:
			return true
		}
	case "BOOL", "BOOLEAN":
		switch f.t {
		case FieldTypes.Bool, FieldTypes.TinyInt:
			return true
		}
	case "DECIMAL":
		switch f.t {
		case FieldTypes.Decimal:
			return true
		}
	case "FLOAT":
		switch f.t {
		case FieldTypes.Float:
			return true
		}
	case "DOUBLE", "REAL":
		switch f.t {
		case FieldTypes.Double, FieldTypes.Real:
			return true
		}
	case "JSON":
		switch f.t {
		case FieldTypes.JSON:
			return true
		}
	case "BLOB":
		switch f.t {
		case FieldTypes.Blob:
			return true
		}
	case "DATE":
		switch f.t {
		case FieldTypes.Date:
			return true
		}
	case "TIME":
		switch f.t {
		case FieldTypes.Time:
			return true
		}
	case "TIMESTAMP":
		switch f.t {
		case FieldTypes.Timestamp:
			return true
		}
	case "ENUM":
		// Enum is coming as a functions with values - ENUM('MALE','FEMALE','EXTRA','OTHER')
		enumValues := make([]string, len(f.definition))
		for i, val := range f.definition {
			enumValues[i] = "'" + strings.ToUpper(fmt.Sprintf("%v", val)+"'")
		}
		enum := fmt.Sprintf("ENUM(%s)", strings.Join(enumValues, ","))
		// println(enum)
		if fieldTypeStr == enum {
			return true
		}
	case "UUID":
		switch f.t {
		case FieldTypes.UUID:
			return true
		}
	default:
		return false
	}

	return false
}

func (f *Field) foreignKeyConstraint() string {
	if f.fk == nil {
		return ""
	}

	stmt := fmt.Sprintf("CONSTRAINT fk_%s_%s FOREIGN KEY (%s) REFERENCES %s(%s)",
		f.table_name, f.name, f.name, f.fk.referenceTable, f.fk.referenceColumn)

	if f.fk.onDelete != "" {
		stmt += " ON DELETE " + f.fk.onDelete
	}
	if f.fk.onUpdate != "" {
		stmt += " ON UPDATE " + f.fk.onUpdate
	}

	return stmt
}

/*
 * ToForeignKey converts a field to a foreign key with the given onDelete and onUpdate actions.
 * It requires the field to be a primary key and sets the foreign key metadata.
 * It also sets the index and unique properties based on the parameters.
 * @parameters:
 * onDelete, onUpdate - actions for foreign key constraints
 * is_primary_key - whether the field is a primary key
 * is_index - whether the field should be indexed
 * is_unique - whether the field should be unique
 * @returns: a new Field with foreign key constraints set
 */
func (f *Field) ToForeignKey(onDelete, onUpdate string, is_primary_key, is_index, is_unique bool) *Field {

	if !f.index.PrimaryKey {
		panic("Field " + f.name + " must be a primary key to set as foreign key")
	}

	// TODO: Find a way to get the table name before the field is created
	fmt.Printf("Table Name of the foreingkey: %s", f.table_name)
	return &Field{
		name:          f.name,
		t:             f.t,
		lenth:         f.lenth,
		nullable:      f.nullable,
		defaultValue:  f.defaultValue,
		autoIncrement: f.autoIncrement,
		index: index{
			PrimaryKey: is_primary_key,
			Index:      is_index,
			Unique:     is_unique,
		},
		table_name: f.table_name,
		fk: &foreignKey{
			referenceTable:  f.table_name,
			referenceColumn: f.name,
			onDelete:        onDelete,
			onUpdate:        onUpdate,
		},
	}
}

// ForeignKey sets the foreign key constraint for a field
// leave onDelete and onUpdate empty if you do not want to set them
// example onDelete: "CASCADE", "SET NULL", "RESTRICT", etc.
// example onUpdate: "CASCADE", "SET NULL", "RESTRICT", etc.
func ForeignKey(field Field, onDelete, onUpdate string, is_primary_key bool, is_index bool, is_unique bool) Field {
	field.index.PrimaryKey = is_primary_key
	field.index.Index = is_index
	field.index.Unique = is_unique

	field.fk = &foreignKey{
		referenceTable:  field.table_name,
		referenceColumn: field.name,
		onDelete:        onDelete,
		onUpdate:        onUpdate,
	}
	return field
}
