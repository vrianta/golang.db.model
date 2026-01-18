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
		db            *sql.DB
		TableName     string       // Name of the table in the database
		FieldTypes    FieldTypeset // Map of field names to their types
		schemas       []schema
		initialised   bool   // Flag to check if the model is initialised
		initialisedDB bool   // Flag to set if the database is initialised by the user
		primary       *Field // name of the primary elemet
		depends_on    []string
		// indexes     map[string]indexInfo // columnName -> index info
	}
)

var (
	syncDatabaseEnabled   bool
	syncComponentsEnabled bool
)

func init() {

	fmt.Println("Model Initialisation Started")

	for _, arg := range os.Args[1:] {
		switch arg {
		case "--migrate-model", "-mm":
			syncDatabaseEnabled = true
		case "--migrate-component", "-mc":
			syncComponentsEnabled = true
		}
	}

	fmt.Printf("Flags are \nSyncDatabase: %v, \nSyncComponents: %v\n", syncDatabaseEnabled, syncComponentsEnabled)

}

/*
 * This Package is to handle model in the database checking and creating tables and providing default functions to handle them
 * It will create the table,
 * It will update the table accordingly during the initial program startup only if the build is not true
 * So Dynaimic Table Updation will be handled during development only
 * It will provide the default functions to handle the model like Create, Read, Update, Delete
 */
func newModel(tableName string, FieldTypes FieldTypeset, depends_on []string) meta {

	for _, field := range FieldTypes {
		if field.fk == nil {
			field.table_name = tableName // Set the table name for each field
		}
	}

	_model := meta{
		components: make(components),
		TableName:  tableName,
		FieldTypes: FieldTypes,
		primary: func(FieldTypes FieldTypeset) *Field {
			for _, field := range FieldTypes {
				if field.index.PrimaryKey {
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

func New[T any](tableName string, structure T) *Table[T] {

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
		meta:   newModel(tableName, FieldTypeset, depends_on),
		Fields: structure,
	}

	ModelsRegistry[tableName] = &response.meta
	return response
}

/*
 * Syncing Table Scenma and Components Syncing
 */
func (t *Table[T]) syncTable() {

	model__ := &t.meta
	create_model := func(model *meta) {
		model.CreateTableIfNotExists()

		if syncDatabaseEnabled {
			fmt.Printf("[Models] Initializing model and syncing database tables for: %s", model.TableName)
			model.syncModelSchema()
			model.syncTableSchema()

			model.initialised = true
		} else {
			fmt.Printf("[Models] Initializing model without syncing database tables for: %s", model.TableName)
			model.initialised = true
		}
		delete(ModelsRegistry, model.TableName)
	}

	if !model__.initialisedDB {
		panic("Database is not initlised in the model")
	}

	// Wait until the database driver is initialised
	// Check if database connection is available, with retry logic
	maxRetries := 30
	retryDelay := 1 // 1 second
	retryCount := 0

	for retryCount < maxRetries {
		// Check if database connection is available
		if model__.db != nil {
			// Verify driver is ready by attempting a connection
			if err := model__.db.Ping(); err == nil {
				break // Driver is ready, proceed with model initialization
			}
		}

		// If we reach here, driver is not ready yet
		if retryCount < maxRetries-1 {
			fmt.Printf("[Models] Waiting for driver initialization for model %s (attempt %d/%d)...\n",
				model__.TableName, retryCount+1, maxRetries)
			time.Sleep(time.Duration(retryDelay) * time.Second)
		}
		retryCount++
	}

	// After all retries, check if driver is ready
	if model__.db == nil {
		panic(fmt.Sprintf("[Models] Database connection not initialized for model: %s after %d attempts",
			model__.TableName, maxRetries))
	}

	if err := model__.db.Ping(); err != nil {
		panic(fmt.Sprintf("[Models] Database driver not ready for model %s after %d attempts: %s",
			model__.TableName, maxRetries, err.Error()))
	}

	// model_for_component := maps.Clone(ModelsRegistry)
	// fmt.Println("Models Registry: ", ModelsRegistry)

	// for _, model := range ModelsRegistry {
	// // Process model and its dependencies
	// WaitForDependsToBeCreated := sync.WaitGroup{}
	// for _, depends_on := range model.depends_on {
	// 	for {
	// 		_, exists := ModelsRegistry[depends_on]
	// 		if !exists {
	// 			fmt.Println("Waiting for the depends on model to be executed")
	// 			continue
	// 		}
	// 	}
	// 	create_model(ModelsRegistry[depends_on])
	// }
	create_model(model__)
	delete(ModelsRegistry, model__.TableName)
	// }

	fmt.Println("---------------------------------------------------------")

	_, err := os.Stat(filepath.Join(componentsDir, model__.TableName+".component.json"))
	if !os.IsNotExist(err) {
		model__.loadComponentFromDisk()
		if syncComponentsEnabled {
			model__.SyncComponentWithDB()
			model__.loadComponentFromDisk()
		} else {
			// means the file exists in the disk
			model__.refreshComponentFromDB()
		}
	}
}

/*
 * Opens Database Connection and have to be called on creation of the model
 */
func (t *Table[T]) InitialiseDB(driver string, DSN string) *Table[T] {
	var err error
	if t.meta.db, err = sql.Open(driver, DSN); err != nil {
		panic("Error opening database: " + err.Error())
	}

	t.meta.initialisedDB = true

	t.syncTable()

	return t
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

	if err := m.db.Ping(); err != nil {
		panic("Database Connection Not Estrablished")
	}
	_, err := m.db.Exec(sql)
	// fmt.Printf("Creating Table Sql Executed : %s", sql)
	if err != nil {
		panic("Error creating table: " + err.Error() + "\nqueryBuilder:" + sql)
	} else {
		fmt.Printf("[Models] Table '%s' ensured to exist.\n", m.TableName)
	}
}

// Handles adding/dropping PRIMARY KEY
func (m *meta) syncPrimaryKey(field *Field, schema *schema) {

	if err := m.db.Ping(); err != nil {
		fmt.Println("Error updating primary key:", err.Error())
		return
	}
	if schema.isprimary && !field.index.PrimaryKey {
		// Drop primary key
		queryBuilder := fmt.Sprintf("ALTER TABLE `%s` DROP PRIMARY KEY;", m.TableName)
		if _, err := m.db.Exec(queryBuilder); err != nil {
			fmt.Println("[Index] Error dropping PRIMARY KEY:", err)
		} else {
			fmt.Printf("[Index] PRIMARY KEY dropped for field: %s\n", field.name)
		}
	}
	if !schema.isprimary && field.index.PrimaryKey {
		// Add primary key
		queryBuilder := "ALTER TABLE " + m.TableName + " ADD PRIMARY KEY (" + field.name + ")"
		if _, err := m.db.Query(queryBuilder); err != nil {
			fmt.Println("[ERROR] failed to Add Primary Key ", err.Error())
			fmt.Println("[FAILED] Failed queryBuilder to Update Primary Key is: ", queryBuilder)
		}
	}
}

// Handles adding/dropping UNIQUE index
func (m *meta) syncUniqueIndex(field *Field, schema *schema) {

	if err := m.db.Ping(); err != nil {
		fmt.Println("Error updating unique index:", err)
		return
	}
	indexName := fmt.Sprintf("unq_%s", field.name)
	if schema.isunique && !field.index.Unique {
		// Drop unique index
		queryBuilder := fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`;", m.TableName, indexName)
		if _, err := m.db.Exec(queryBuilder); err != nil {
			fmt.Println("[Index] Error dropping UNIQUE:", err)
		} else {
			fmt.Printf("[Index] UNIQUE dropped for field: %s\n", field.name)
		}
	}
	if !schema.isunique && field.index.Unique {
		// Add unique index
		queryBuilder := fmt.Sprintf("ALTER TABLE `%s` ADD UNIQUE `%s` (`%s`);", m.TableName, indexName, field.name)
		if _, err := m.db.Exec(queryBuilder); err != nil {
			fmt.Println("[Index] Error adding UNIQUE:", err)
		} else {
			fmt.Printf("[Index] UNIQUE added for field: %s\n", field.name)
		}
	}
}

// Handles adding/dropping normal INDEX
func (m *meta) syncIndex(field *Field, schema *schema) {
	if err := m.db.Ping(); err != nil {
		fmt.Println("Error updating index:", err)
		return
	}
	indexName := fmt.Sprintf("idx_%s", field.name)
	if schema.isindex && !field.index.Index {
		// Drop index
		queryBuilder := fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`;", m.TableName, indexName)
		if _, err := m.db.Exec(queryBuilder); err != nil {
			fmt.Println("[Index] Error dropping INDEX:", err)
		} else {
			fmt.Printf("[Index] INDEX dropped for field: %s\n", field.name)
		}
	}
	if !schema.isindex && field.index.Index {
		// Add index
		queryBuilder := fmt.Sprintf("ALTER TABLE `%s` ADD INDEX `%s` (`%s`);", m.TableName, indexName, field.name)
		if _, err := m.db.Exec(queryBuilder); err != nil {
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
	maps.Copy(q.InsertRowFieldTypes, values)
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
		if field.index.PrimaryKey {
			m.primary = field
			return true // Return the pointer directly from the map
		}
	}

	return false
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
