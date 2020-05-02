package main

import (
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/tidb/types"
	"strings"
)

var (
	reqColCheck = map[string]func(r *ParseResult, def *ast.ColumnDef){
		columnID:        checkColIDDef,
		columnCreatedAt: checkColCreatedAtDef,
		columnUpdatedAt: checkColUpdatedAtDef,
	}
	reqColNotFoundErr = map[string]error{
		columnID:        ColIDNotFoundErr,
		columnCreatedAt: ColCreatedAtNotFoundErr,
		columnUpdatedAt: ColUpdatedAtNotFoundErr,
	}
	reqColDroppedErr = map[string]error{
		columnID:        ColIDDroppedErr,
		columnCreatedAt: ColCreatedAtDroppedErr,
		columnUpdatedAt: ColUpdatedAtDroppedErr,
	}
)

// checkAllColDef: Check if all added columns pass the validation and return converted map struct
func checkAllColDef(r *ParseResult, cols []*ast.ColumnDef, ddlType string) map[string]*ast.ColumnDef {
	colMap := make(map[string]*ast.ColumnDef)
	for _, col := range cols {
		checkColDef(r, col, ddlType) // Rule 4.4

		colName := strings.ToLower(col.Name.Name.String())
		if checker, ok := reqColCheck[colName]; ok {
			checker(r, col)
		}

		colMap[colName] = col
	}
	return colMap
}

// Rule: column `id`
//       `BIGINT UNSIGNED` type with `AUTO INCREMENT`
func checkColIDDef(r *ParseResult, col *ast.ColumnDef) {
	colInfo := col.Tp
	if colInfo == nil || !mysql.HasUnsignedFlag(colInfo.Flag) { // UNSIGNED
		r.AddError(ColIDNotUnsignedErr)
	}
	if colInfo == nil || colInfo.Tp != mysql.TypeLonglong { // BIGINT
		r.AddError(ColIDNotBigIntErr)
	}

	autoIncSet := false
	for _, option := range col.Options {
		if option.Tp == ast.ColumnOptionAutoIncrement { // Auto Increment
			autoIncSet = true
		}
	}
	if !autoIncSet {
		r.AddError(ColIDNotAutoIncErr)
	}
}

// Rule: column `created_at`
//       `DATETIME` type with `NOT NULL DEFAULT CURRENT_TIMESTAMP`
func checkColCreatedAtDef(r *ParseResult, col *ast.ColumnDef) {
	if col.Tp == nil || col.Tp.Tp != mysql.TypeDatetime { // DATETIME
		r.AddError(ColCreatedAtNotDateTimeErr)
	}
	notNullSet := false
	defaultValueSet := false
	for _, option := range col.Options {
		if option.Tp == ast.ColumnOptionNotNull { // NOT NULL
			notNullSet = true
		} else if option.Tp == ast.ColumnOptionDefaultValue { // DEFAULT CURRENT_TIMESTAMP
			defaultValueSet = checkCurrTimestampExpr(option.Expr)
		}
	}
	if !notNullSet {
		r.AddError(ColCreatedAtNotNotNullErr)
	}
	if !defaultValueSet {
		r.AddError(ColCreatedAtInvalidDefValErr)
	}
}

// Rule: column `updated_at`
//       `DATETIME` type with `NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` except historical table ends with _history
func checkColUpdatedAtDef(r *ParseResult, col *ast.ColumnDef) {
	if col.Tp == nil || col.Tp.Tp != mysql.TypeDatetime { // DATETIME
		r.AddError(ColUpdatedAtNotDateTimeErr)
	}
	notNullSet := false
	defaultValueSet := false
	onUpdateSet := false
	for _, option := range col.Options {
		if option.Tp == ast.ColumnOptionNotNull { // NOT NULL
			notNullSet = true
		} else if option.Tp == ast.ColumnOptionDefaultValue { // DEFAULT CURRENT_TIMESTAMP
			defaultValueSet = checkCurrTimestampExpr(option.Expr)
		} else if option.Tp == ast.ColumnOptionOnUpdate { // ON UPDATE CURRENT_TIMESTAMP
			onUpdateSet = checkCurrTimestampExpr(option.Expr)
		}
	}
	if !notNullSet {
		r.AddError(ColUpdatedAtNotNotNullErr)
	}
	if !defaultValueSet {
		r.AddError(ColUpdatedAtInvalidDefValErr)
	}
	if !onUpdateSet {
		r.AddError(ColUpdatedAtInvalidOnUpdateErr)
	}
}

// Rule: Column Definition
//    1. Don't include hyphen `-` in column name, use underscore `_` instead
//    2: No inline key declaration. All Primary/Unique/Reference key must use INDEX line
//    3: Use `DATETIME` instead of `TIMESTAMP`
//    4: Don't specify a display length for BIGINT, INT, MEDIUMINT, SMALLINT, and TINYINT etc.
//    5: Don't use 'DEFAULT' for TEXT, JSON, BLOB (LARGE OBJECT) data type
//    6: Don't use ENUM fields for storing input
//    7: Don't use reserved keyword for column name
//    8: Don't use upper case for column name
//    9: `NOT NULL` column must provided `DEFAULT` unless has `AUTO_INCREMENT` or violate RULE 5 (Warning) ALTER TABLE only
func checkColDef(r *ParseResult, col *ast.ColumnDef, ddlType string) {
	colInfo := col.Tp
	colName := col.Name.Name.String()
	if strings.Contains(colName, Hyphen) { // Rule 1
		r.AddError(ColNameWithHyphenErr.Accept(colName))
	}

	if _, ok := reservedWords[strings.ToUpper(colName)]; ok { // Rule 7
		r.AddError(ColReservedWordErr.Accept(colName))
	}
	if strings.ToLower(colName) != colName { // Rule 8
		r.AddError(ColNameNotLowerCaseErr.Accept(colName))
	}

	hasNotNull, hasDefault, canSkipDefault, autoIncSet := false, false, false, false
	for _, option := range col.Options {
		if option.Tp == ast.ColumnOptionPrimaryKey || option.Tp == ast.ColumnOptionUniqKey || option.Tp == ast.ColumnOptionReference { // Rule 2
			r.AddError(ColInlineKeyErr.Accept(colName))
		} else if option.Tp == ast.ColumnOptionDefaultValue {
			hasDefault = true
		} else if option.Tp == ast.ColumnOptionNotNull {
			hasNotNull = true
		} else if option.Tp == ast.ColumnOptionAutoIncrement {
			autoIncSet = true
		}
	}

	if colInfo != nil {
		if colInfo.Tp == mysql.TypeTimestamp { // Rule 3
			r.AddError(ColTimeStampTypeErr.Accept(colName))
		} else if colInfo.EvalType() == types.ETInt { // Rule 4
			// Note: bool will be parsed as tinyint(1)
			if colInfo.Tp != mysql.TypeTiny && colInfo.Flen != types.UnspecifiedLength {
				r.AddError(ColDisplayLengthIntErr.Accept(colName))
			}
		} else if colInfo.Tp == mysql.TypeJSON || colInfo.Tp == mysql.TypeGeometry || types.IsTypeBlob(colInfo.Tp) {
			if hasDefault { // Rule 5
				r.AddError(ColDefaultBlobJsonErr.Accept(colName, types.TypeStr(colInfo.Tp)))
			}
			canSkipDefault = true
		} else if colInfo.Tp == mysql.TypeEnum { // Rule 6
			r.AddError(ColEnumTypeErr.Accept(colName))
		}
	}


	if ddlType == TypeAlterTable && hasNotNull && !canSkipDefault && !hasDefault && !autoIncSet { // Rule 9
		r.AddError(ColNotNullWithoutDefaultErr.Accept(colName))
	}
}

// return if the expr evaluates to `CURRENT_TIMESTAMP`
func checkCurrTimestampExpr(expr ast.ExprNode) bool {
	if expr == nil {
		return false
	}

	funcCallExpr, ok := expr.(*ast.FuncCallExpr)
	return ok && strings.ToLower(funcCallExpr.FnName.String()) == ast.CurrentTimestamp
}

func getAllColNames(cols []*ast.ColumnDef) map[string]string {
	colNames := make(map[string]string)
	for _, col := range cols {
		colInfo := col.Tp
		colName := col.Name.Name.String()

		if colInfo != nil {
			colNames[colName] = strings.ToUpper(types.TypeToStr(colInfo.Tp, colInfo.Charset))
		}
	}
	return colNames
}
