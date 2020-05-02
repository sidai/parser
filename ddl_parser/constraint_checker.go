package main

import (
	"github.com/pingcap/parser/ast"
	"strings"
)

var (
	reqConCheck = map[string]func(r *ParseResult, def *ast.Constraint){
		indexCreatedAt: checkIndexCreatedAtDef,
		indexUpdatedAt: checkIndexUpdatedAtDef,
	}
	reqConNotFoundErr = map[string]error{
		primaryKey:     PrimaryKeyNotFoundErr,
		indexCreatedAt: KeyCreatedAtNotFoundErr,
		indexUpdatedAt: KeyUpdatedAtNotFoundErr,
	}
	reqConDroppedErr = map[string]error{
		primaryKey:     PrimaryKeyDroppedErr,
		indexCreatedAt: KeyCreatedAtDroppedErr,
		indexUpdatedAt: KeyUpdatedAtDroppedErr,
	}
	conTypeStringMap = map[ast.ConstraintType]string{
		ast.ConstraintNoConstraint: noConstraintPrefix,
		ast.ConstraintPrimaryKey:   primaryKey,
		ast.ConstraintKey:          keyPrefix,
		ast.ConstraintIndex:        keyPrefix,
		ast.ConstraintUniq:         uniqueKeyPrefix,
		ast.ConstraintUniqKey:      uniqueKeyPrefix,
		ast.ConstraintUniqIndex:    uniqueKeyPrefix,
		ast.ConstraintForeignKey:   foreignKeyPrefix,
		ast.ConstraintSpatial:      spatialKeyPrefix,
		ast.ConstraintSpatialKey:   spatialKeyPrefix,
		ast.ConstraintSpatialIndex: spatialKeyPrefix,
		ast.ConstraintFulltext:     fullTextPrefix,
		ast.ConstraintCheck:        checkPrefix,
	}
)

// checkAllConstraintDef: Check if all added constraints pass the validation and return converted map struct
// Rule: Constraint Definition
//     1: `PRIMARY KEY` must contains column `id` and all of the columns used in `PARTITION`
//     2: `UNIQUE` key, must contains all of the columns used in `PARTITION`
//     3: `DATETIME` or `TIMESTAMP` column in `COMPOSITE KEY` must appear at the end
//     4: Don't use `FOREIGN KEY`
//     5:  All columns used in CONSTRAINT must be declared in column list before
func checkAllConstraintDef(r *ParseResult, cons []*ast.Constraint, cols []*ast.ColumnDef, part *ast.PartitionOptions) map[string]*ast.Constraint {
	conMap := make(map[string]*ast.Constraint)
	for _, con := range cons {
		declaredCols, declaredParts := getAllColNames(cols), getAllPartitionColumn(part)
		conName := checkConstraintDef(r, con, declaredCols, declaredParts)
		checkConstraintColDef(r, con, declaredCols)
		conMap[conName] = con
	}

	return conMap
}

func checkNewConstraintDef(r *ParseResult, con *ast.Constraint) string {
	return checkConstraintDef(r, con, nil, nil)
}

// checkConstraintDef: Check if the given constraints pass the validation and return the auto-generated name
func checkConstraintDef(r *ParseResult, con *ast.Constraint, colNames map[string]string, partCol map[string]struct{}) string {
	conName := con.Name
	switch conTypeStringMap[con.Tp] {
	case primaryKey:
		conName = primaryKey
		checkPrimaryKeyDef(r, con, colNames, partCol)
	case uniqueKeyPrefix:
		checkUniqueKeyDef(r, con, colNames, partCol)
	case keyPrefix:
		checkKeyDef(r, con, colNames)
	case foreignKeyPrefix:
		checkForeignKeyDef(r, con, colNames)
	}

	if checker, ok := reqConCheck[conName]; ok {
		checker(r, con)
	}

	return conName
}

// Rule: all constraint
//       All columns used in CONSTRAINT must be declared in column list before
func checkConstraintColDef(r *ParseResult, con *ast.Constraint, colNames map[string]string) {
	for colName, _ := range getAllUsedColumn(con) {
		if _, ok := colNames[colName]; !ok {
			r.AddError(ConWithUnknownColErr.Accept(colName, restoreClause(con)))
		}
	}
}

// Rule: primary key
//       must contains `id`
//       must contains all columns used in `PARTITION KEY`
//       range column must at the end
func checkPrimaryKeyDef(r *ParseResult, con *ast.Constraint, colNames map[string]string, partCol map[string]struct{}) {
	conUsed := getAllUsedColumn(con)
	if _, ok := conUsed[columnID]; !ok  {
		r.AddError(PrimaryKeyIDNotFoundErr)
	}
	checkUniqueKeyPartKeyDef(r, con, conUsed, partCol)
	checkIndexRangeKeyDef(r, con, conUsed, colNames)
}

// Rule: unique key
//       must prefix with `uk_`
//       must contains all columns used in `PARTITION KEY`
//       range column must at the end
func checkUniqueKeyDef(r *ParseResult, con *ast.Constraint, colNames map[string]string, partCol map[string]struct{}) {
	conUsed := getAllUsedColumn(con)
	checkUniqueKeyPartKeyDef(r, con, conUsed, partCol)
	checkIndexRangeKeyDef(r, con, conUsed, colNames)
	if !strings.HasPrefix(con.Name, uniqueKeyPrefix + NameSeparator) {
		r.AddError(UniqueKeyPrefixErr.Accept(restoreClause(con)))
	}
}

// Rule: key
//       must prefix with `index_`
//       range column must at the end
func checkKeyDef(r *ParseResult, con *ast.Constraint, colNames map[string]string) {
	conUsed := getAllUsedColumn(con)
	checkIndexRangeKeyDef(r, con, conUsed, colNames)
	if !strings.HasPrefix(con.Name, keyPrefix + NameSeparator) {
		r.AddError(IndexNamePrefixErr.Accept(restoreClause(con)))
	}
}

// Rule: foreign key
//       not allowed
func checkForeignKeyDef(r *ParseResult, con *ast.Constraint, colNames map[string]string) {
	r.AddError(ForeignKeyErr.Accept(restoreClause(con)))
}

func checkUniqueKeyPartKeyDef(r *ParseResult, con *ast.Constraint, conUsed map[string]int, partCol map[string]struct{}) {
	for colName, _ := range partCol {
		if _, ok := conUsed[colName]; !ok {
			r.AddError(UniqueKeyPartKeyNotFoundErr.Accept(colName, restoreClause(con)))
		}
	}
}

func checkIndexRangeKeyDef(r *ParseResult, con *ast.Constraint, conUsed map[string]int, colNames map[string]string) {
	for colName, colType := range colNames {
		if colType == columnTypeTimeStamp || colType == columnTypeDateTime {
			if index, ok := conUsed[colName]; ok && index != len(conUsed)-1 {
				r.AddError(CompKeyNoEndRangeKeyErr.Accept(colName, colType, restoreClause(con)))
			}
		}
	}
}

// Rule: index `index_created_at`
//       must be `INDEX/KEY` with column (`created_at`) only
func checkIndexCreatedAtDef(r *ParseResult, con *ast.Constraint) {
	if conTypeStringMap[con.Tp] != keyPrefix || len(con.Keys) != 1 || con.Keys[0].Column.Name.String() != columnCreatedAt {
		r.AddError(KeyCreatedAtFormatErr)
	}
}

// Rule: index `index_updated_at`
//       must be `INDEX/KEY` with column (`updated_at`) only
func checkIndexUpdatedAtDef(r *ParseResult, con *ast.Constraint) {
	if conTypeStringMap[con.Tp] != keyPrefix || len(con.Keys) != 1 || con.Keys[0].Column.Name.String() != columnUpdatedAt {
		r.AddError(KeyUpdatedAtFormatErr)
	}
}

// Return all of the column used with index in the given constraint
func getAllUsedColumn(con *ast.Constraint) map[string]int {
	colUsed := make(map[string]int)
	for index, key := range con.Keys {
		colUsed[key.Column.Name.String()] = index
	}
	return colUsed
}

// Return the matching conventional constraint name provided the constraint type and columns
func formStandardConName(con *ast.Constraint) string {
	tokens := []string{conTypeStringMap[con.Tp]}
	for _, key := range con.Keys {
		tokens = append(tokens, key.Column.Name.String())
	}
	return strings.Join(tokens, NameSeparator)
}