package wfx

import "go.starlark.net/syntax"

type MakefxFile struct {
	ast *syntax.File

	// cached for your convenience, as we validated things.
	targets []*syntax.DefStmt
}

func (x *MakefxFile) ListTargets() (res []string) {
	for _, target := range x.targets {
		res = append(res, target.Name.Name)
	}
	return
}

func ParseMakefxFile(filename string, body string) (*MakefxFile, error) {
	syntaxObj, err := syntax.Parse(filename, body, syntax.RetainComments)
	if err != nil {
		return nil, err
	}
	res := &MakefxFile{ast: syntaxObj}
	res.targets, err = findTargets(res.ast)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func findTargets(ast *syntax.File) (res []*syntax.DefStmt, err error) {
	// Targets can only be top-level defs.
	// So, a simple non-recursive range suffices.
	// Thereafter, they must have a certain known signature --
	// they must have a first argument that is named exactly "fx".
	for _, stmt := range ast.Stmts {
		switch stmt2 := stmt.(type) {
		case *syntax.DefStmt:
			if len(stmt2.Params) < 1 {
				continue
			}
			if extractIdent(stmt2.Params[0]).Name == "fx" {
				res = append(res, stmt2)
			}
		}
	}
	return
}
