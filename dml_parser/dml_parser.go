package main

import "C"
import (
	"encoding/json"
	"fmt"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/format"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"strings"
)

func main() {}

const (
	dmlDelete = "DELETE"
	dmlInsert = "INSERT"
	dmlUpdate = "UPDATE"
)

//export Parse
func Parse(sql string) *C.char {
	return toCGOReturn(parse(sql))
}

func toCGOReturn(parseResults []*ParseResult) *C.char {
	jsonStr, _ := json.Marshal(parseResults)
	return stringToCString(string(jsonStr))
}

func stringToCString(str string) *C.char {
	cs := C.CString(str)
	return cs
}

type ParseResult struct {
	SQL        string   `json:"sql"`
	TableNames []string `json:"tables"`
	DMLType    string   `json:"dml_type"`
	ScanClause string   `json:"scan_clause"`
	Count      int      `json:"insert_count"`
	Error      string   `json:"error_msg"`
}

func parse(sql string) []*ParseResult {
	var results []*ParseResult
	p := parser.New()

	stmts, _, err := p.Parse(sql, "", "")
	if err != nil {
		result := NewParseResult(sql).SetError(fmt.Sprintf("Syntax error: %+v\n", err))
		results = append(results, result)
		return results
	}

	for _, stmt := range stmts {
		result := NewParseResult(stmt.Text())
		switch impl := stmt.(type) {
		case *ast.DeleteStmt:
			result.SetType(dmlDelete).SetTables(extractTableNameFromExpr(impl.TableRefs)).SetScanClause(restoreClause(
				&ast.SelectStmt{
					SelectStmtOpts: &ast.SelectStmtOpts{
						Distinct:   true,
						SQLCache:   true,
						TableHints: impl.TableHints,
					},
					Distinct:   true,
					From:       impl.TableRefs,
					Where:      impl.Where,
					Fields:     &ast.FieldList{
						Fields: []*ast.SelectField{
							{
								WildCard:  new(ast.WildCardField),
							},
						},
					},
					OrderBy:    impl.Order,
					Limit:      impl.Limit,
					TableHints: impl.TableHints,
				}))
		case *ast.UpdateStmt:
			result.SetType(dmlUpdate).SetTables(extractTableNameFromExpr(impl.TableRefs)).SetScanClause(restoreClause(
				&ast.SelectStmt{
					SelectStmtOpts: &ast.SelectStmtOpts{
						Distinct:   true,
						SQLCache:   true,
						TableHints: impl.TableHints,
					},
					Distinct:   true,
					From:       impl.TableRefs,
					Where:      impl.Where,
					Fields:     &ast.FieldList{
						Fields: []*ast.SelectField{
							{
								WildCard:  new(ast.WildCardField),
							},
						},
					},
					OrderBy:    impl.Order,
					Limit:      impl.Limit,
					TableHints: impl.TableHints,
				}))
		case *ast.InsertStmt:
			result.SetType(dmlInsert).SetTables(extractTableNameFromExpr(impl.Table)).
				SetCount(len(impl.Lists)).SetScanClause(restoreClause(impl.Select))
		default:
			result.SetError("statement provided is not a valid DELETE, UPDATE or INSERT query")
		}

		results = append(results, result)
	}

	return results
}

func NewParseResult(sql string) *ParseResult {
	return &ParseResult{SQL: strings.TrimSpace(sql)}
}

func (p *ParseResult) SetError(err string) *ParseResult {
	p.Error = err
	return p
}

func (p *ParseResult) SetTables(tableNames []string) *ParseResult {
	exist := make(map[string]struct{})
	for _, name := range tableNames {
		if _, ok := exist[name]; !ok {
			p.TableNames = append(p.TableNames, name)
			exist[name] = struct{}{}
		}
	}
	return p
}

func (p *ParseResult) SetType(dmlType string) *ParseResult {
	p.DMLType = dmlType
	return p
}

func (p *ParseResult) SetCount(count int) *ParseResult {
	p. Count = count
	return p
}

func (p *ParseResult) SetScanClause(clause string) *ParseResult {
	p.ScanClause = clause
	return p
}

func extractTableNameFromExpr(expr *ast.TableRefsClause) []string {
	if expr == nil {
		return []string{}
	}
	visitor := new(tableNameVisitor)
	expr.Accept(visitor)
	return visitor.tableName
}

func restoreClause(node ast.Node) string {
	if node == nil {
		return ""
	}
	var sb strings.Builder
	_ = node.Restore(format.NewRestoreCtx(format.DefaultRestoreFlags, &sb))
	return sb.String()
}

type tableNameVisitor struct {
	tableName []string
}

func (v *tableNameVisitor) Enter(in ast.Node) (out ast.Node, skipChildren bool) {
	if tableNameNode, ok := in.(*ast.TableName); ok {
		v.tableName = append(v.tableName, restoreClause(tableNameNode))
	}
	return in, false
}

func (v *tableNameVisitor) Leave(in ast.Node) (out ast.Node, ok bool) {
	return in, true
}
