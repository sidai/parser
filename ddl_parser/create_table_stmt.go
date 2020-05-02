package main

import (
	"github.com/pingcap/parser/ast"
)

type CreateTableStmt struct {
	*ast.CreateTableStmt
}

// Create Table Check:
// Rule 1: Cannot create temporary table
//      2: Cannot create table with `IF NOT EXISTS`
//      3: Cannot populate table while creating
//      4: Cannot use create table like statement
//      5: Table Column Definition Check
//      6: Table Constraint Definition Check
//      7: Table Option Definition Check
//      8: Table Name Definition Check
//      9: Table Partition Definition Check
func (s *CreateTableStmt) Parse(r *ParseResult) {
	if s.IsTemporary { // Rule 1
		r.AddError(TempTableErr)
	}

	if s.IfNotExists { // Rule 2
		r.AddError(UseIfNotExistsErr)
	}

	if s.Select != nil { // Rule 3
		r.AddError(CreateWithSelectErr)
	}

	if s.ReferTable != nil { // Rule 4
		r.AddError(CreateWithLikeErr)
	}

	s.checkTableColumnsDef(r) // Rule 5
	s.checkTableConstraintsDef(r) // Rule 6
	s.checkTableOptionsDef(r) // Rule 7
    s.checkTableNameDef(r) // Rule 8
	s.checkTablePartitionDef(r) // Rule 9

	r.AddDDLType(TypeCreateTable)
}

// Rule 5: Table Column Definition Check
//      5.1: Must hav e `AUTO INCREMENT` column `id` of `BIGINT UNSIGNED` type
//      5.2: Must have column `created_at` of `DATETIME` type with `NOT NULL DEFAULT CURRENT_TIMESTAMP`
//      5.3: Must have column `updated_at` of `DATETIME` type with `NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE
//           CURRENT_TIMESTAMP` except historical table ends with _history
//      5.4: Column Option Definition Check
func (s *CreateTableStmt) checkTableColumnsDef(r *ParseResult) {
	colMap := checkAllColDef(r, s.Cols, TypeCreateTable)

	for reqCol, err := range reqColNotFoundErr {
		if _, ok := colMap[reqCol]; !ok {
			r.AddError(err)
		}
	}

	r.AddColumns(getAllColNames(s.Cols))
}

// Rule 6: Table Constraint Definition Check
//      6.1: Must add `PRIMARY KEY` containing column `id`
//      6.2: `UNIQUE` key must contains all of the columns used in `PARTITION`
//      6.3: `DATETIME` or `TIMESTAMP` column in `COMPOSITE KEY` must appear at the end (Warning)
//      6.2: Must add `index_created_at` as Index using `created_at` column
//      6.3: Must add `index_updated_at` as Index using `updated_at` column
//      6.4: Don't use `FOREIGN KEY`
//      6.7: All columns used in CONSTRAINT must be declared in column list before
func (s *CreateTableStmt) checkTableConstraintsDef(r *ParseResult) {
	conMap := checkAllConstraintDef(r, s.Constraints, s.Cols, s.Partition)

	for reqCon, err := range reqConNotFoundErr {
		if _, ok := conMap[reqCon]; !ok {
			r.AddError(err)
		}
	}
}

// Rule 7: Table Option Definition Check
//      7.1: Do specify the `CHARSET` when creating new table
//      7.2: Do specify the `COLLATE` when creating new table
//      7.3: Don't use `utf8mb4_general_ci` for `COLLATE`, use `utf8mb4_unicode_ci` instead
//      7.4: Don't specify a table's ENGINE
func (s *CreateTableStmt) checkTableOptionsDef(r *ParseResult) {
	optMap := checkAllTableOptDef(r, s.Options)

	for reqOpt, err := range reqTableOptNotFoundErr {
		if _, ok := optMap[reqOpt]; !ok {
			r.AddError(err)
		}
	}
}

// Rule 8: Table Name Definition Check
//      8.1: Table name cannot contains database name
//      8.2: Table name must in lower case
//      8.3: Table name cannot use any of the reserved key word
//      8.4: Don't include hyphen `-` in table name, use underscore `_` instead
func (s *CreateTableStmt) checkTableNameDef(r *ParseResult) {
	tableName := s.Table.Name.String()
	schemaName := s.Table.Schema.String()
	if schemaName != "" {
		tableName = schemaName + DBNameSeparator + tableName
	}

	r.SetOldTable(tableName)
	checkTableNameDef(r, tableName)
}

// Rule 9: Table Partition Definition Check
//      9.1: All columns used in PARTITION must be declared in column list before
//      9.2: Don't use HASH partition (Warning)
func (s *CreateTableStmt) checkTablePartitionDef(r *ParseResult) {
	checkAllPartitionDef(r, s.Partition, s.Cols)
}