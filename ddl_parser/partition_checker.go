package main

import (
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/model"
)

// checkAllPartitionDef: Check if all added partitions pass the validation
// Rule 1: all columns used must be declared in column list before
//      2: Don't use HASH partition (Warning)
func checkAllPartitionDef(r *ParseResult, part *ast.PartitionOptions, cols []*ast.ColumnDef) {
	checkPartColumnsDef(r, part, cols)
	checkPartTypeDef(r, part)
}

// Rule: all constraint columns
//       All columns used in PARTITION must be declared in column list before
func checkPartColumnsDef(r *ParseResult, part *ast.PartitionOptions, cols []*ast.ColumnDef) {
	colUsed := getAllPartitionColumn(part)
	colDeclared := getAllColNames(cols)

	for colName, _ := range colUsed {
		if _, ok := colDeclared[colName]; !ok {
			r.AddError(PartWithUnknownColErr.Accept(colName, restoreClause(part)))
		}
	}
}


// Rule: all constraint type
//       Don't use HASH partition (Warning)
func checkPartTypeDef(r *ParseResult, part *ast.PartitionOptions) {
	containsHash := false
	if part != nil {
		if part.Tp == model.PartitionTypeHash {
			containsHash = true
		}
		if part.Sub != nil && part.Sub.Tp == model.PartitionTypeHash {
			containsHash = true
		}
	}
	if containsHash {
		r.AddError(PartWithHashErr.Accept(part))
	}
}

// Return all columns used in PARTITION as a map
func getAllPartitionColumn(part *ast.PartitionOptions) map[string]struct{} {
	if part == nil {
		return nil
	}
	colNames := make(map[string]struct{})

	// For RANGE, LIST and HASH partition
	for _, colName := range extractColNameFromExpr(part.Expr) {
		colNames[colName] = struct{}{}
	}
	if part.Sub != nil {
		for _, colName := range extractColNameFromExpr(part.Sub.Expr) {
			colNames[colName] = struct{}{}
		}
	}

	// For KEY, RANGE COLUMNS and LIST COLUMNS partition
	for _, col := range part.ColumnNames {
		colNames[col.Name.String()] = struct{}{}
	}
	if part.Sub != nil {
		for _, col := range part.Sub.ColumnNames {
			colNames[col.Name.String()] = struct{}{}
		}
	}
	return colNames
}

func extractColNameFromExpr(expr ast.ExprNode) []string {
	if expr == nil {
		return []string{}
	}
	visitor := colNameVisitor{}
	expr.Accept(&visitor)
	return visitor.colName
}

type colNameVisitor struct{
	colName []string
}

func (v *colNameVisitor) Enter(in ast.Node) (out ast.Node, skipChildren bool) {
	if colNameNode, ok := in.(*ast.ColumnName); ok {
		v.colName = append(v.colName, colNameNode.Name.String())
	}
	return in, false
}

func (v *colNameVisitor) Leave(in ast.Node) (out ast.Node, ok bool) {
	return in, true
}
