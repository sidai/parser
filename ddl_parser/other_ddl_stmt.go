package main

import (
	"github.com/pingcap/parser/ast"
)

type RenameTableStmt struct{ ast.StmtNode }
type ModifyIndexStmt struct{ ast.StmtNode }
type ModifyDatabaseStmt struct{ ast.StmtNode }
type DeleteTableStmt struct{ ast.StmtNode }
type UnsupportedDDLStmt struct{ ast.StmtNode }

func (s *RenameTableStmt) Parse(r *ParseResult) {
	r.AddError(RenameTableErr)
}

func (s *ModifyIndexStmt) Parse(r *ParseResult) {
	r.AddError(ModifyIndexErr)
}

func (s *ModifyDatabaseStmt) Parse(r *ParseResult) {
	r.AddError(ModifyDatabaseErr)
}

func (s *DeleteTableStmt) Parse(r *ParseResult) {
	r.AddError(DeleteTableErr)
}

func (s *UnsupportedDDLStmt) Parse(r *ParseResult) {
	r.AddError(getUnsupportedClauseErr(s.StmtNode))
}
