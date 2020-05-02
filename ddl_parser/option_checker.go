package main

import (
	"github.com/pingcap/parser/ast"
	"strings"
)

var (
	reqTableOptCheck = map[ast.TableOptionType]func(r *ParseResult, def *ast.TableOption){
		ast.TableOptionCharset: checkTableOptCharsetDef,
		ast.TableOptionCollate: checkTableOptCollateDef,
		ast.TableOptionEngine:  checkTableOptEngineDef,
	}
	reqTableOptNotFoundErr = map[ast.TableOptionType]error{
		ast.TableOptionCharset: NoCharsetErr,
		ast.TableOptionCollate: NoCollateErr,
	}
)

// checkAllTableOptDef: Check if all added table options pass the validation and return converted map struct
func checkAllTableOptDef(r *ParseResult, options []*ast.TableOption) map[ast.TableOptionType]*ast.TableOption {
	optMap := make(map[ast.TableOptionType]*ast.TableOption)
	for _, option := range options {
		optName := option.Tp
		if checker, ok := reqTableOptCheck[optName]; ok {
			checker(r, option)
		}
		optMap[optName] = option
	}
	return optMap
}

// Rule: table option `CHARSET`
//       No defined rule
func checkTableOptCharsetDef(r *ParseResult, option *ast.TableOption) {
}

// Rule: table option `COLLATE`
//       Don't use `utf8mb4_general_ci` for `COLLATE`, use `utf8mb4_unicode_ci` instead
func checkTableOptCollateDef(r *ParseResult, option *ast.TableOption) {
	if strings.ToLower(option.StrValue) == badCollate {
		r.AddError(BadCollateErr)
	}
}

// Rule: table option `ENGINE`
//       Engine must be set or default as InnoDB
func checkTableOptEngineDef(r *ParseResult, option *ast.TableOption) {
	if !option.Default && strings.ToLower(option.StrValue) != EngineInnoDB {
		r.AddError(InvalidEngineErr)
	}
}
