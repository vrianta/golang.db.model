package model

import (
	"fmt"
	"strings"
	"sync"
	"unicode"
)

func (m *meta) validate() {
	var wg sync.WaitGroup
	var mu sync.Mutex

	primaryKeyCount := 0
	fieldNames := make(map[string]struct{})
	var firstErr error
	var panicOnce sync.Once

	for _, field := range m.FieldTypes {
		wg.Add(1)
		go func(f *Field) {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					panicOnce.Do(func() {
						firstErr = fmt.Errorf("%v", r)
					})
				}
			}()

			// Check for duplicate field names
			mu.Lock()
			if _, exists := fieldNames[f.name]; exists {
				mu.Unlock()
				panic(fmt.Sprintf("[Validation Error] Duplicate field name '%s' in Table '%s'.", f.name, m.TableName))
			}
			fieldNames[f.name] = struct{}{}
			mu.Unlock()

			if f.t == FieldTypes.Enum && f.definition == nil {
				panic(fmt.Sprintf("[Validation Error] Field '%s' in Table '%s' is of type ENUM but has no definition.", f.name, m.TableName))
			} else if f.definition != nil && len(f.definition) == 0 {
				panic(fmt.Sprintf("Field '%s' of type ENUM must have Definition values", f.name))

			}
			// PRIMARY KEY and UNIQUE cannot both be true
			if f.index.PrimaryKey {
				if f.index.Unique {
					panic(fmt.Sprintf("[Validation Error] Field '%s' in Table '%s' cannot be both PRIMARY KEY and UNIQUE.", f.name, m.TableName))
				}
				// for primry key the types allowed are varchat or int
				if f.t != FieldTypes.Int && f.t != FieldTypes.VarChar {
					panic(fmt.Sprintf("[Validation Error] Field '%s' in Table '%s' cannot be both PRIMARY KEY and UNIQUE.", f.name, m.TableName))
				}

				mu.Lock()
				primaryKeyCount++
				mu.Unlock()

				if f.nullable {
					panic(fmt.Sprintf("[Validation Error] Field '%s' in Table '%s' is PRIMARY KEY but marked as nullable.", f.name, m.TableName))
				}
				if f.defaultValue != "" {
					panic(fmt.Sprintf("[Validation Error] Field '%s' in Table '%s' is PRIMARY KEY but has a default value.", f.name, m.TableName))
				}

			}

			if f.autoIncrement {
				if !f.index.PrimaryKey {
					panic(fmt.Sprintf("[Validation Error] Field '%s' in Table '%s' is AUTO_INCREMENT but not PRIMARY KEY.", f.name, m.TableName))
				}
				if !strings.HasPrefix(strings.ToLower(f.t.string()), "int") {
					panic(fmt.Sprintf("[Validation Error] Field '%s' in Table '%s' is AUTO_INCREMENT but not of integer type.", f.name, m.TableName))
				}
			}

			f.Validate()
		}(field)
	}

	wg.Wait()

	// After all field validations complete, check for multiple primary keys
	if firstErr != nil {
		panic(firstErr)
	}
	if primaryKeyCount > 1 {
		panic(fmt.Sprintf("[Validation Error] Table '%s' has more than one PRIMARY KEY field.", m.TableName))
	}
}

func (f Field) Validate() {
	if f.name == "" {
		panic("Field name cannot be empty")
	}
	if !isAlphaNumeric(f.name) {
		panic(fmt.Sprintf("Field name '%s' contains invalid characters", f.name))
	}
	if !unicode.IsLetter(rune(f.name[0])) {
		panic(fmt.Sprintf("Field name '%s' must start with a letter", f.name))
	}
	if sqlKeywords[strings.ToUpper(f.name)] {
		panic(fmt.Sprintf("Field name '%s' is a reserved SQL keyword", f.name))
	}

	switch f.t {
	case FieldTypes.TinyInt:
		if f.lenth < 3 {
			panic(fmt.Sprintf("Field '%s': TINYINT length must be at least 3", f.name))
		}
	case FieldTypes.Bool:
		if f.lenth < 0 || f.lenth > 1 {
			panic(fmt.Sprintf("Field '%s': BOOLEAN length must be 1", f.name))
		}
	case FieldTypes.SmallInt:
		if f.lenth < 5 {
			panic(fmt.Sprintf("Field '%s': SMALLINT length must be at least 5", f.name))
		}
	case FieldTypes.MediumInt:
		if f.lenth < 6 {
			panic(fmt.Sprintf("Field '%s': MEDIUMINT length must be at least 6", f.name))
		}
	case FieldTypes.Int, FieldTypes.BigInt:
		if f.lenth < 1 {
			panic(fmt.Sprintf("Field '%s': %s must have a positive length", f.name, f.t.string()))
		}
	case FieldTypes.VarChar, FieldTypes.Char:
		if f.lenth < 1 {
			panic(fmt.Sprintf("Field '%s': %s must have a positive length", f.name, f.t.string()))
		}
	case FieldTypes.Decimal:
		if f.lenth < 1 {
			panic(fmt.Sprintf("Field '%s': DECIMAL must have Length > 0", f.name))
		}
	case FieldTypes.Text, FieldTypes.Blob, FieldTypes.JSON, FieldTypes.Date, FieldTypes.Time, FieldTypes.Timestamp:
		if f.lenth > 0 {
			panic(fmt.Sprintf("Field '%s': Type %s should not have Length", f.name, f.t.string()))
		}
	}

	if f.autoIncrement && !f.t.IsNumeric() {
		panic(fmt.Sprintf("Field '%s': AutoIncrement is only allowed on numeric fields", f.name))
	}

	if f.index.PrimaryKey && f.nullable {
		panic(fmt.Sprintf("Field '%s': Primary key fields cannot be nullable", f.name))
	}

	if (f.index.Unique || f.index.Index) && (f.t == FieldTypes.Text || f.t == FieldTypes.Blob) {
		panic(fmt.Sprintf("Field '%s': Cannot use INDEX/UNIQUE on TEXT/BLOB fields", f.name))
	}

	if f.defaultValue != "" && !f.t.IsValueCompatible(f.defaultValue) {
		panic(fmt.Sprintf("Field '%s': Default value '%s' is not compatible with type %s", f.name, f.defaultValue, f.t.string()))
	}
}
