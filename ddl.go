package model

import (
	"fmt"
)

// function to add the new field in the table
func (m *meta) addField(field *Field) {
	/*
		ALTER TABLE `users`
		ADD `newel` VARCHAR(20) NULL DEFAULT 'dwads' AFTER `userId`,
		ADD INDEX (`newel`);
	*/

	response := "ALTER TABLE `" + m.TableName + "`\n"
	response += "ADD " + field.columnDefinition() + field.addIndexStatement() + ";"

	if err := m.db.Ping(); err != nil {
		panic("Error While Adding new Field to the table" + err.Error())
	} else {
		if _, sql_err := m.db.Exec(response); sql_err != nil {
			panic("Error While Updating the Table Field" + sql_err.Error())
		} else {
			fmt.Printf("[AddField]      Table: %-20s | Field Added: %-20s\n", m.TableName, field.name)
		}
	}
}

// function to change the field details
func (m *meta) modifyDBField(field *Field) {
	// ALTER TABLE `users` CHANGE `userId` `userId` INT(30) NOT NULL AUTO_INCREMENT;
	response := "ALTER TABLE `" + m.TableName + "` "
	response += "CHANGE `" + field.name + "` " + field.columnDefinition() + ";"

	if err := m.db.Ping(); err != nil {
		panic("Error While Changing Field" + err.Error())
	} else {
		if _, sql_err := m.db.Exec(response); sql_err != nil {
			panic("Error While Changing the Table Field" + sql_err.Error() + "SQL queryBuilder: " + response)
		} else {
			fmt.Printf("[modifyDBField]   Table: %-20s | Field Updated: %-20s\n", m.TableName, field.name)
		}
	}
}

// Drop a field from the databasess
func (m *meta) removeDBField(fieldName string) {
	//ALTER TABLE `users` DROP `userId`;
	queryBuilder := "ALTER TABLE `" + m.TableName + "` DROP `" + fieldName + "`;"
	if err := m.db.Ping(); err != nil {
		panic("Error While Deleting Field" + err.Error())
	} else {
		if _, sql_err := m.db.Exec(queryBuilder); sql_err != nil {
			panic("Error While Deleting the Field" + sql_err.Error())
		} else {
			fmt.Printf("[removeDBField]     Table: %-20s | Field Dropped: %-20s\n", m.TableName, fieldName)
		}
	}
}
