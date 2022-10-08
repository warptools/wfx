package wfx

import (
	"fmt"

	"go.starlark.net/syntax"
)

// Reduce the expression to an Ident.  This is valid to use on expressions in syntax.DefStatement.Params.
func extractIdent(expr syntax.Expr) *syntax.Ident {
	switch lhs := expr.(type) {
	case *syntax.Ident:
		return lhs
	case *syntax.BinaryExpr:
		return extractIdent(lhs.X)
	default:
		panic("?!unrecognized:" + fmt.Sprintf("%T", expr))
	}
}
