package wfx

import (
	"fmt"
	"testing"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

func TestHello(t *testing.T) {
	filename := "thefilename.star"
	source := `
zonk = "pow"

# This is a comment.
# On several lines.
def foobar():
	pass

def frobnoz():
	pass
`

	// Build the globals map that makes our API's surfaces available in starlark.
	// (This is actually needed even to parse!)
	predef := starlark.StringDict{}

	// Parse it ourselves.  What can we do with this?
	// Not deep enough, won't retain comments: // syntaxObj, programObj, err := starlark.SourceProgram(filename, source, predef.Has)
	syntaxObj, err := syntax.Parse(filename, source, syntax.RetainComments)
	if err != nil {
		panic(err)
	}

	for i := range syntaxObj.Stmts {
		t.Logf("stmt %d: %s -- comments: %v", i, debugRef(syntaxObj.Stmts[i]), syntaxObj.Stmts[i].Comments())
	}

	// Execute Starlark program in a file.
	thread := &starlark.Thread{Name: "thethreadname"}
	globals, err := starlark.ExecFile(thread, filename, source, predef)
	if err != nil {
		panic(err)
	}

	// Retrieve a module global.  (This is probably not how we'll have warpforge's system extract results, but it's interesting.)
	fmt.Printf("result = %v\n", globals["result"])
}

func debugRef(stmt syntax.Stmt) string {
	switch stmt2 := stmt.(type) {
	case *syntax.AssignStmt:
		return fmt.Sprintf("AssignStmt^%q@%s", stmt2.LHS.(*syntax.Ident).Name, stmtStartPosStr(stmt))
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
		return "?!unrecognized"
	}
}
func stmtStartPosStr(stmt syntax.Stmt) string {
	start, _ := stmt.Span()
	return fmt.Sprintf("%d:%d", start.Line, start.Col)
}
