package model

import (
	"database/sql"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
)

type (
	Table[T any] struct {
		meta
		Fields T
	}

	meta struct {
		components
		*sql.DB
		TableName   string       // Name of the table in the database
		FieldTypes  FieldTypeset // Map of field names to their types
		schemas     []schema
		initialised bool   // Flag to check if the model is initialised
		primary     *Field // name of the primary elemet
		depends_on  []string
		// indexes     map[string]indexInfo // columnName -> index info
	}
)

func init() {

	var SyncDatabaseEnabled bool
	var SyncComponentsEnabled bool

	for _, arg := range os.Args[1:] {
		switch arg {
		case "--migrate-model", "-mm":
			SyncDatabaseEnabled = true
		case "--migrate-component", "-mc":
			SyncComponentsEnabled = true
		}
	}

	create_model := func(model *meta) {
		model.CreateTableIfNotExists()

		if SyncDatabaseEnabled {
			fmt.Printf("[Models] Initializing model and syncing database tables for: %s", model.TableName)
			model.syncModelSchema()
			model.syncTableSchema()

			model.initialised = true
		}
		delete(ModelsRegistry, model.TableName)
	}

	model_for_component := maps.Clone(ModelsRegistry)
	for _, model := range ModelsRegistry {
		// Wait until the database driver is initialised
		// Check if database connection is available, with retry logic
		maxRetries := 30
		retryDelay := 1 // 1 second
		retryCount := 0

		for retryCount < maxRetries {
			// Check if database connection is available
			if model.DB != nil {
				// Verify driver is ready by attempting a connection
				if err := model.Ping(); err == nil {
					break // Driver is ready, proceed with model initialization
				}
			}

			// If we reach here, driver is not ready yet
			if retryCount < maxRetries-1 {
				fmt.Printf("[Models] Waiting for driver initialization for model %s (attempt %d/%d)...\n",
					model.TableName, retryCount+1, maxRetries)
				time.Sleep(time.Duration(retryDelay) * time.Second)
			}
			retryCount++
		}

		// After all retries, check if driver is ready
		if model.DB == nil {
			panic(fmt.Sprintf("[Models] Database connection not initialized for model: %s after %d attempts",
				model.TableName, maxRetries))
		}

		if err := model.Ping(); err != nil {
			panic(fmt.Sprintf("[Models] Database driver not ready for model %s after %d attempts: %s",
				model.TableName, maxRetries, err.Error()))
		}

		// Process model and its dependencies
		for _, depends_on := range model.depends_on {
			create_model(ModelsRegistry[depends_on])
		}
		create_model(model)
		delete(ModelsRegistry, model.TableName)
	}

	fmt.Println("---------------------------------------------------------")

	for _, model := range model_for_component {

		_, err := os.Stat(filepath.Join(componentsDir, model.TableName+".component.json"))
		if !os.IsNotExist(err) {
			model.loadComponentFromDisk()
			if SyncComponentsEnabled {
				model.SyncComponentWithDB()
				model.loadComponentFromDisk()
			} else {
				// means the file exists in the disk
				model.refreshComponentFromDB()
			}
		}
	}
	initialsed = true
}

/*
 * This Package is to handle model in the database checking and creating tables and providing default functions to handle them
 * It will create the table,
 * It will update the table accordingly during the initial program startup only if the build is not true
 * So Dynaimic Table Updation will be handled during development only
 * It will provide the default functions to handle the model like Create, Read, Update, Delete
 */
func newModel(database *sql.DB, tableName string, FieldTypes FieldTypeset, depends_on []string) meta {

	for _, field := range FieldTypes {
		if field.fk == nil {
			field.table_name = tableName // Set the table name for each field
		}
	}

	_model := meta{
		DB:         database,
		components: make(components),
		TableName:  tableName,
		FieldTypes: FieldTypes,
		primary: func(FieldTypes FieldTypeset) *Field {
			for _, field := range FieldTypes {
				if field.Index.PrimaryKey {
					return field // Return the pointer directly from the map
				}
			}
			return nil
		}(FieldTypes),
		depends_on: depends_on,
	}

	_model.validate()

	return _model
}

func New[T any](database *sql.DB, tableName string, structure T) *Table[T] {

	if err := database.Ping(); err != nil {
		panic(fmt.Sprintf("Database connection failed: %s for Table: %s", err.Error(), tableName))
	}

	t := reflect.TypeOf(structure)
	v := reflect.ValueOf(structure)

	if t.Kind() != reflect.Struct {
		panic("structure passed to New must be a struct")
	}

	FieldTypeset := make(FieldTypeset, t.NumField())
	depends_on := []string{}
	for i := 0; i < t.NumField(); i++ {
		structField := t.Field(i)
		valueField := v.Field(i)

		// Handle pointer to Field
		fieldPtr, ok := valueField.Interface().(*Field)
		if !ok {
			panic(fmt.Sprintf("[Model Error] Field '%s' is not of type *model.Field of Model %s", structField.Name, tableName))
		}
		if fieldPtr == nil {
			panic(fmt.Sprintf("[Validation Error] Field '%s' in Talble %s Body is not Defined", structField.Name, tableName))
		}
		// Update metadata
		fieldPtr.name = structField.Name
		if fieldPtr.fk == nil {
			fieldPtr.table_name = tableName
		}

		FieldTypeset[structField.Name] = fieldPtr
	}

	response := &Table[T]{
		meta:   newModel(database, tableName, FieldTypeset, depends_on),
		Fields: structure,
	}

	ModelsRegistry[tableName] = &response.meta
	return response
}

func (m *meta) CreateTableIfNotExists() {
	sql := "CREATE TABLE IF NOT EXISTS " + m.TableName + " (\n"
	fieldDefs := []string{}

	for _, field := range m.FieldTypes {
		fieldDefs = append(fieldDefs, field.columnDefinition())
	}

	for _, field := range m.FieldTypes {
		indexStatements := field.addIndexStatement()
		if indexStatements != "" {
			fieldDefs = append(fieldDefs, indexStatements)
		}
	}

	sql += strings.Join(fieldDefs, ",\n")
	sql += "\n);"

	if err := m.Ping(); err != nil {
		panic("Database Connection Not Estrablished")
	}
	_, err := m.Exec(sql)
	// log.Info("Creating Table Sql Executed : %s", sql)
	if err != nil {
		panic("Error creating table: " + err.Error() + "\nqueryBuilder:" + sql)
	}
}

// Handles adding/dropping PRIMARY KEY
func (m *meta) syncPrimaryKey(field *Field, schema *schema) {

	if err := m.Ping(); err != nil {
		fmt.Println("Error updating primary key:", err.Error())
		return
	}
	if schema.isprimary && !field.Index.PrimaryKey {
		// Drop primary key
		queryBuilder := fmt.Sprintf("ALTER TABLE `%s` DROP PRIMARY KEY;", m.TableName)
		if _, err := m.Exec(queryBuilder); err != nil {
			fmt.Println("[Index] Error dropping PRIMARY KEY:", err)
		} else {
			fmt.Printf("[Index] PRIMARY KEY dropped for field: %s\n", field.name)
		}
	}
	if !schema.isprimary && field.Index.PrimaryKey {
		// Add primary key
		queryBuilder := "ALTER TABLE " + m.TableName + " ADD PRIMARY KEY (" + field.name + ")"
		if _, err := m.Query(queryBuilder); err != nil {
			fmt.Println("[ERROR] failed to Add Primary Key ", err.Error())
			fmt.Println("[FAILED] Failed queryBuilder to Update Primary Key is: ", queryBuilder)
		}
	}
}

func logSection(header string) {
	fmt.Println("---------------------------------------------------------")
	fmt.Println(header)
	fmt.Println("---------------------------------------------------------")
}

// Handles adding/dropping UNIQUE index
func (m *meta) syncUniqueIndex(field *Field, schema *schema) {

	if err := m.Ping(); err != nil {
		fmt.Println("Error updating unique index:", err)
		return
	}
	indexName := fmt.Sprintf("unq_%s", field.name)
	if schema.isunique && !field.Index.Unique {
		// Drop unique index
		queryBuilder := fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`;", m.TableName, indexName)
		if _, err := m.Exec(queryBuilder); err != nil {
			fmt.Println("[Index] Error dropping UNIQUE:", err)
		} else {
			fmt.Printf("[Index] UNIQUE dropped for field: %s\n", field.name)
		}
	}
	if !schema.isunique && field.Index.Unique {
		// Add unique index
		queryBuilder := fmt.Sprintf("ALTER TABLE `%s` ADD UNIQUE `%s` (`%s`);", m.TableName, indexName, field.name)
		if _, err := m.Exec(queryBuilder); err != nil {
			fmt.Println("[Index] Error adding UNIQUE:", err)
		} else {
			fmt.Printf("[Index] UNIQUE added for field: %s\n", field.name)
		}
	}
}

// Handles adding/dropping normal INDEX
func (m *meta) syncIndex(field *Field, schema *schema) {
	if err := m.Ping(); err != nil {
		fmt.Println("Error updating index:", err)
		return
	}
	indexName := fmt.Sprintf("idx_%s", field.name)
	if schema.isindex && !field.Index.Index {
		// Drop index
		queryBuilder := fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`;", m.TableName, indexName)
		if _, err := m.Exec(queryBuilder); err != nil {
			fmt.Println("[Index] Error dropping INDEX:", err)
		} else {
			fmt.Printf("[Index] INDEX dropped for field: %s\n", field.name)
		}
	}
	if !schema.isindex && field.Index.Index {
		// Add index
		queryBuilder := fmt.Sprintf("ALTER TABLE `%s` ADD INDEX `%s` (`%s`);", m.TableName, indexName, field.name)
		if _, err := m.Exec(queryBuilder); err != nil {
			fmt.Println("[Index] Error adding INDEX:", err)
		} else {
			fmt.Printf("[Index] INDEX added for field: %s\n", field.name)
		}
	}
}

// get the table name
func (m *meta) GetTableName() string {
	return m.TableName
}

// Convert the Fetched Data to a of objects
// This function will convert the Table to a map[string]any for easy access and manipulation
// func (m *Table) ToMap() map[string]any {
// 	response := make(map[string]any, len(m.FieldTypes))
// 	var wg sync.WaitGroup
// 	var mu sync.Mutex

// 	for _, field := range m.FieldTypes {
// 		wg.Add(1)

// 		go func(f *Field) {
// 			defer wg.Done()

// 			var value any
// 			if f.value != nil {
// 				value = f.value
// 			} else {
// 				value = nil
// 			}

// 			mu.Lock()
// 			response[f.Name] = value
// 			mu.Unlock()
// 		}(&field)
// 	}

// 	wg.Wait()
// 	return response
// }

// InsertRow InsertRows a new record into the table using the provided values map.
// This is a dedicated Create/InsertRow function that does not overlap with table creation or schema management.
func (m *meta) InsertRow(values map[string]any) error {
	q := m.Create()
	for k, v := range values {
		q.InsertRowFieldTypes[k] = v
	}
	return q.Exec()
}

func (m *meta) GetPrimaryKey() *Field {
	if !m.HasPrimaryKey() {
		panic("Primary Key is Required for but the Model(" + m.TableName + ") ")
	}
	return m.primary
}

/*
To check if the model has primary key or not

true ->  if exists
false -> if not exists
*/
func (m *meta) HasPrimaryKey() bool {
	if m.primary != nil {
		return true
	}
	for _, field := range m.FieldTypes {
		if field.Index.PrimaryKey {
			m.primary = field
			return true // Return the pointer directly from the map
		}
	}

	return false
}

/*
GetField(fieldname) -> return pointer of the field
*/

func (m *meta) GetField(field_name string) *Field {
	field, ok := m.FieldTypes[field_name]
	if !ok {
		return nil
	}
	return field
}

/*
GetField(fieldname) -> return pointer of the field
*/

func (m *meta) GetFieldTypes() *FieldTypeset {
	return &m.FieldTypes
}

// Print the Objects of the models as the good for debug perpose
func (r *Results) PrintAsTable() {
	if len(*r) == 0 {
		return
	}

	// Collect all unique column names across all rows
	colSet := map[string]struct{}{}
	for _, row := range *r {
		for col := range row {
			colSet[col] = struct{}{}
		}
	}

	// Sort column names for consistent display
	var colNames []string
	for col := range colSet {
		colNames = append(colNames, col)
	}
	sort.Strings(colNames)

	// Print header
	for _, col := range colNames {
		fmt.Printf("| %-15s", col)
	}
	fmt.Println("|")

	// Print separator
	fmt.Println(strings.Repeat("-", len(colNames)*18))

	// Print each row
	for _, row := range *r {
		for _, col := range colNames {
			val := row[col]
			fmt.Printf("| %-15v", val)
		}
		fmt.Println("|")
	}
}
