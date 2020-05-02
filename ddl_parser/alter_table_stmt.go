package main

import (
	"github.com/pingcap/parser/ast"
	"strings"
)

type AlterTableStmt struct {
	*ast.AlterTableStmt
}

// Alter Table Check:
// Rule 1: Table Option Modification Check
//      2: Column Modification Check
//      3: Constraint Modification Check
//      4: Partition Modification Check
//      5: Table Rename Check
//      6: Alter Table Approach Check
func (s *AlterTableStmt) Parse(r *ParseResult) {
	r.SetOldTable(s.Table.Name.String())
	for _, spec := range s.Specs {
		switch spec.Tp {
		case ast.AlterTableOption:
			s.checkModifyTableOption(r, spec)

		case ast.AlterTableAddColumns, ast.AlterTableDropColumn, ast.AlterTableModifyColumn,
			ast.AlterTableChangeColumn, ast.AlterTableRenameColumn, ast.AlterTableAlterColumn:
			s.checkModifyColumn(r, spec)

		case ast.AlterTableAddConstraint, ast.AlterTableDropPrimaryKey, ast.AlterTableDropIndex,
			ast.AlterTableDropForeignKey, ast.AlterTableRenameIndex, ast.AlterTableEnableKeys,
			ast.AlterTableDisableKeys, ast.AlterTableAlterCheck, ast.AlterTableDropCheck, ast.AlterTableIndexInvisible:
			s.checkModifyConstraint(r, spec)

		case ast.AlterTableAddPartitions, ast.AlterTableCoalescePartitions, ast.AlterTableDropPartition,
			ast.AlterTableTruncatePartition, ast.AlterTablePartition, ast.AlterTableRemovePartitioning,
			ast.AlterTableRebuildPartition, ast.AlterTableReorganizePartition, ast.AlterTableCheckPartitions,
			ast.AlterTableExchangePartition, ast.AlterTableOptimizePartition, ast.AlterTableRepairPartition,
			ast.AlterTableImportPartitionTablespace, ast.AlterTableDiscardPartitionTablespace:
			s.checkModifyPartition(r, spec)

		case ast.AlterTableRenameTable:
			s.checkRenameTable(r, spec)

		case ast.AlterTableLock, ast.AlterTableAlgorithm, ast.AlterTableForce, ast.AlterTableWithValidation,
			ast.AlterTableWithoutValidation:
			s.checkAlterApproach(r, spec)

		default:
			s.checkOtherAlter(r, spec)
		}
	}
}

// Rule 1: Table Option Modification Check
//      1.1: Don't use `utf8mb4_general_ci` for `COLLATE`
//      1.2: Don't specify a table's ENGINE
func (s *AlterTableStmt) checkModifyTableOption(r *ParseResult, spec *ast.AlterTableSpec) {
	r.AddDDLType(TypeModifyOption)
	checkAllTableOptDef(r, spec.Options)
}

// Rule 2: Column Modification Check
// Rule 2.1: Drop column require DE's confirmation (Warning)
//      2.2: Cannot drop column `id`, must be `AUTO_INCREMENT BIGINT UNSIGNED` after modification
//      2.3: Cannot drop column `created_at`, must be `DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP` after modification
//      2.4: Cannot drop column `updated_at`
//           must be `DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` after modification
//      2.5: Column Option Definition Check for new added column
//      2.6: Don't use FIRST or AFTER to reorder column (Warning)
func (s *AlterTableStmt) checkModifyColumn(r *ParseResult, spec *ast.AlterTableSpec) {
	r.AddDDLType(TypeModifyColumn)
	// New added columns must pass all column check
	if len(spec.NewColumns) != 0 {
		checkAllColDef(r, spec.NewColumns, TypeAlterTable)
	}

	// Reorder of column is not permitted
	if pos := spec.Position; pos != nil {
		if pos.Tp == ast.ColumnPositionFirst {
			r.AddError(ColReorderWithFirstErr.Accept(spec.NewColumns[0].Name.String()))
		} else if pos.Tp == ast.ColumnPositionAfter {
			r.AddError(ColReorderWithAfterErr.Accept(spec.NewColumns[0].Name.String()))
		}
	}

	// Removal of required columns is not permitted
	if spec.OldColumnName != nil {
		droppedCol := strings.ToLower(spec.OldColumnName.Name.String())
		r.AddError(ColDroppedErr.Accept(droppedCol))
		if err, ok := reqColDroppedErr[droppedCol]; ok {
			r.AddError(err)
		}
	}

	r.AddColumns(getAllColNames(spec.NewColumns))
}

// Rule 3: Constraint Modification Check
// Rule 3.1: Cannot drop `PRIMARY KEY`
//      3.2: Cannot drop or rename KEY `index_created_at`
//      3.3: Cannot drop or rename KEY `index_updated_at` except historical table ends with _history
//      3.4: Cannot add `FOREIGN KEY`
func (s *AlterTableStmt) checkModifyConstraint(r *ParseResult, spec *ast.AlterTableSpec) {
	r.AddDDLType(TypeModifyConstraint)
	var conDropped string
	switch spec.Tp {
	case ast.AlterTableAddConstraint:
		checkNewConstraintDef(r, spec.Constraint)
	case ast.AlterTableDropPrimaryKey:
		conDropped = primaryKey
	case ast.AlterTableDropIndex:
		conDropped = spec.Name
	case ast.AlterTableRenameIndex:
		conDropped = spec.FromKey.String()
	}

	if err, ok := reqConDroppedErr[conDropped]; ok {
		r.AddError(err)
	}
}

// Rule 4: Partition Modification Check
//         Cannot drop, truncate or remove partition (Warning)
func (s *AlterTableStmt) checkModifyPartition(r *ParseResult, spec *ast.AlterTableSpec) {
	r.AddDDLType(TypeModifyPartition)
	if spec.Tp == ast.AlterTableDropPartition || spec.Tp == ast.AlterTableTruncatePartition {
		r.AddError(PartDroppedErr)
	} else if spec.Tp == ast.AlterTableRemovePartitioning {
		r.AddError(PartRemovedErr)
	}
}

// Rule 5: Table Rename Check
//      5.1: Table name cannot contains database name
//      5.2: Table name must in lower case
//      5.3: Table name cannot use any of the reserved key word
//      5.4: Don't include hyphen `-` in table name, use underscore `_` instead
func (s *AlterTableStmt) checkRenameTable(r *ParseResult, spec *ast.AlterTableSpec) {
	r.AddDDLType(TypeRenameTable)
	newTable := spec.NewTable.Name.String()
	newSchema := spec.NewTable.Schema.String()
	if newSchema != "" {
		newTable = newSchema + DBNameSeparator + newTable
	}

	r.SetNewTable(newTable)
	checkTableNameDef(r, newTable)
}

// Rule 6: Alter Table Approach Check
//         no checking required
func (s *AlterTableStmt) checkAlterApproach(r *ParseResult, spec *ast.AlterTableSpec) {
}

// Rule 7: Other Alter Check
//         not allowed
func (s *AlterTableStmt) checkOtherAlter(r *ParseResult, spec *ast.AlterTableSpec) {
	r.AddError(getUnsupportedClauseErr(spec))
}