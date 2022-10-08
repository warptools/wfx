package wfx

import (
	"fmt"

	"go.starlark.net/syntax"
)

func debugRef(stmt syntax.Stmt) string {
	switch stmt2 := stmt.(type) {
	case *syntax.AssignStmt:
		return fmt.Sprintf("AssignStmt^%q@%s", nameExpr(stmt2.LHS), stmtStartPosStr(stmt))
	case *syntax.BranchStmt:
		return "BranchStmt@" + stmtStartPosStr(stmt)
	case *syntax.DefStmt:
		return fmt.Sprintf("DefStmt^%q@%s", stmt2.Name.Name, stmtStartPosStr(stmt))
	case *syntax.ExprStmt:
		return "ExprStmt@" + stmtStartPosStr(stmt)
	case *syntax.ForStmt:
		return "ForStmt@" + stmtStartPosStr(stmt)
	case *syntax.WhileStmt:
		return "WhileStmt@" + stmtStartPosStr(stmt)
	case *syntax.IfStmt:
		return "IfStmt@" + stmtStartPosStr(stmt)
	case *syntax.LoadStmt:
		return "LoadStmt@" + stmtStartPosStr(stmt)
	case *syntax.ReturnStmt:
		return "ReturnStmt@" + stmtStartPosStr(stmt)
	default:
		return "?!unrecognized:" + fmt.Sprintf("%T", stmt)
	}
}
func stmtStartPosStr(stmt syntax.Stmt) string {
	start, _ := stmt.Span()
	return fmt.Sprintf("%d:%d", start.Line, start.Col)
}
func nameExpr(expr syntax.Expr) string { // mostly meant for use on *syntax.AssignStmt.LHS, but uses Expr as arg because it recurses.
	switch lhs := expr.(type) {
	case *syntax.Ident:
		return lhs.Name
	case *syntax.DotExpr:
		return nameExpr(lhs.X) + "." + lhs.Name.Name
	case *syntax.IndexExpr:
		return nameExpr(lhs.X) + "[" + nameExpr(lhs.Y) + "]"
	case *syntax.Literal: // not possible in `*syntax.AssignStmt.LHS`, but it is when recursing into IndexExpr.
		return lhs.Raw // happily, this still includes the quotes and all.
	default:
		return "?!unrecognized:" + fmt.Sprintf("%T", expr)
	}
}
