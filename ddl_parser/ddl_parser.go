package main

import (
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/format"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"strings"
)

func main() {}

type ReturnResult struct {
	SQL      string            `json:"sql"`
	OldTable string            `json:"old_table"`
	NewTable string            `json:"new_table"`
	Columns  map[string]string `json:"columns"`
	DDLType  []string          `json:"ddl_list"`
	Error    []string          `json:"error_msg"`
	Warning  []string          `json:"warning_msg"`
}

type Parser interface {
	Parse(r *ParseResult)
}

type ParseResult struct {
	SQL      string
	OldTable string
	NewTable string
	Columns  map[string]string
	DDLType  map[string]struct{}
	Error    []*ReturnError
}

func (r *ParseResult) toReturnResult() *ReturnResult {
	returnResult := &ReturnResult{
		SQL:      r.SQL,
		OldTable: r.OldTable,
		NewTable: r.NewTable,
		Columns:  r.Columns,
	}

	for ddlType, _ := range r.DDLType {
		returnResult.DDLType = append(returnResult.DDLType, ddlType)
	}
	for _, err := range r.Error {
		if err.Level() == DDLMsgTypeError {
			returnResult.Error = append(returnResult.Error, err.Error())
		} else if err.Level() == DDLMsgTypeWarning {
			returnResult.Warning = append(returnResult.Warning, err.Error())
		}
	}
	return returnResult
}

func NewParseResult(sql string) *ParseResult {
	return &ParseResult{
		SQL:      sql,
		Columns:  make(map[string]string),
		DDLType:  make(map[string]struct{}),
	}
}

func (r *ParseResult) SetOldTable(tableName string) {
	r.OldTable = tableName
}

func (r *ParseResult) SetNewTable(tableName string) {
	r.NewTable = tableName
}

func (r *ParseResult) AddColumn(colName string, colType string) {
	r.Columns[colName] = colType
}

func (r *ParseResult) AddColumns(cols map[string]string) {
	for colName, colType := range cols {
		r.AddColumn(colName, colType)
	}
}

func (r *ParseResult) AddDDLType(ddlType string) {
	r.DDLType[ddlType] = struct{}{}
}

func (r *ParseResult) AddError(err error) {
	r.Error = append(r.Error, &ReturnError{
		errorMsg: err.Error(),
		level: DDLErrorMsgTypeMap[err],
	})
}

func parse(sql string) []*ParseResult {
	var results []*ParseResult
	p := parser.New()

	stmts, _, err := p.Parse(sql, "", "")
	if err != nil {
		result := NewParseResult(sql)
		result.AddError(SyntaxErr.Accept(err.Error()))
		results = append(results, result)
		return results
	}

	for _, stmt := range stmts {
		result := NewParseResult(stmt.Text())
		if _, ok := stmts[0].(ast.DDLNode); !ok {
			result.AddError(NoneDDLErr)
		} else {
			var ddlParser Parser
			switch impl := stmt.(type) {
			case *ast.CreateTableStmt:
				ddlParser = &CreateTableStmt{impl}
			case *ast.AlterTableStmt:
				ddlParser = &AlterTableStmt{impl}
			case *ast.RenameTableStmt:
				ddlParser = &RenameTableStmt{impl}
			case *ast.CreateIndexStmt, *ast.DropIndexStmt:
				ddlParser = &ModifyIndexStmt{impl}
			case *ast.CreateDatabaseStmt, *ast.DropDatabaseStmt, *ast.AlterDatabaseStmt:
				ddlParser = &ModifyDatabaseStmt{impl}
			case *ast.DropTableStmt, *ast.TruncateTableStmt:
				ddlParser = &DeleteTableStmt{impl}
			default:
				ddlParser = &UnsupportedDDLStmt{impl}
			}
			ddlParser.Parse(result)
		}

		results = append(results, result)
	}

	return results
}

func getUnsupportedClauseErr(node ast.Node) error {
	return UnsupportedClauseErr.Accept(restoreClause(node))
}

func restoreClause(node ast.Node) string {
	var sb strings.Builder
	_ = node.Restore(format.NewRestoreCtx(format.DefaultRestoreFlags, &sb))
	return sb.String()
}

// Rule: Table Name Definition Check
//     1: Table name cannot contains database name
//     2: Table name must in lower case
//     3: Table name cannot use any of the reserved key word
//     4: Don't include hyphen `-` in table name, use underscore `_` instead
func checkTableNameDef(r *ParseResult, tableName string) {
	if strings.Contains(tableName, DBNameSeparator) { // Rule 1
		r.AddError(TableWithDBNameErr.Accept(strings.Split(tableName, DBNameSeparator)[0]))
	}

	if strings.ToLower(tableName) != tableName { // Rule 2
		r.AddError(TableNotLowerCaseErr.Accept(tableName))
	}

	if _, ok := reservedWords[strings.ToUpper(tableName)]; ok { // Rule 3
		r.AddError(TableReservedWordErr.Accept(tableName))
	}

	if strings.Contains(tableName, Hyphen) {
		r.AddError(TableNameWithHyphenErr.Accept(tableName))
	}
}
