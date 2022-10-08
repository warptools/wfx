package wfx

import (
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type MakefxFile struct {
	ast *syntax.File

	// cached for your convenience, as we validated things.
	targets []*Target
}

func (x *MakefxFile) ListTargets() []*Target {
	return x.targets
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

type Target struct {
	name   string  // mostly the def name, but for file-based targets, may be a generated mangle.
	parent *Target // usually nil, but for file-based targets that have been manifested, points to the def that made them.

	stmt     *syntax.DefStmt
	callable starlark.Callable // nil until MakefxFile.Eval has prepared us.
}

func (t *Target) Name() string {
	return t.name
}

func findTargets(ast *syntax.File) (res []*Target, err error) {
	// Targets can only be top-level defs.
	// So, a simple non-recursive range suffices.
	// Thereafter, they must have a certain known signature --
	// they must have a first argument that is named exactly "fx".
	// (This rule may expand in the future -- for example, "fx_files=*" for other target forms.)
	// Any defs not matching the pattern are simply regular functions.
	for _, stmt := range ast.Stmts {
		switch stmt2 := stmt.(type) {
		case *syntax.DefStmt:
			if len(stmt2.Params) < 1 {
				continue
			}
			if extractIdent(stmt2.Params[0]).Name == "fx" {
				res = append(res, &Target{
					name: stmt2.Name.Name,
					stmt: stmt2,
				})
			}
		}
	}
	return
}
