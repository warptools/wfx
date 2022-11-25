package wfx

import (
	"github.com/serum-errors/go-serum"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"

	"github.com/warptools/wfx/pkg/wfxapi"
)

// FxFile stores a parsed starlark syntax AST plus cached info about targets.
type FxFile struct {
	ast *syntax.File

	// cached for your convenience, as we validated things.
	targets       []*Target
	targetsByName map[string]*Target
}

func (x *FxFile) ListTargets() []*Target {
	return x.targets
}

// ParseFxFile parses a an "fx file", which is expected to contain starlark code matching certain conventions.
// The filename argument is advisory; the body argument is data that has already been loaded.
//
// Aside from parsinng the file as starlark syntax (and returning any immediate errors from that),
// it also looks for the "fx" conventions that denote functions that are "targets",
// and builds a map of those.
//
// Note that what is checked by this function is purely syntax parse.
// It does not check that references in the code resolve, for example -- that comes later.
func ParseFxFile(filename string, body string) (*FxFile, error) {
	syntaxObj, err := syntax.Parse(filename, body, syntax.RetainComments)
	if err != nil {
		return nil, err
	}
	res := &FxFile{ast: syntaxObj}
	res.targets, err = findTargets(res.ast)
	if err != nil {
		return nil, err
	}
	res.targetsByName = make(map[string]*Target, len(res.targets))
	for _, t := range res.targets {
		res.targetsByName[t.name] = t
	}
	return res, nil
}

type Target struct {
	name      string   // mostly the def name, but for file-based targets, may be a generated mangle.
	parent    *Target  // usually nil, but for file-based targets that have been manifested, points to the def that made them.
	dependsOn []string // dependencies are by string name.

	stmt     *syntax.DefStmt
	callable starlark.Callable // nil until FxFile.Eval has prepared us.

	// future: not entirely clear if these will alwaysalways have stmt and callable.
}

func (t *Target) Name() string {
	return t.name
}
func (t *Target) DependsOn() []string {
	return t.dependsOn
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
			// First, the checklist to see if we have a target.  Continue otherwise.
			if len(stmt2.Params) < 1 {
				continue
			}
			if extractIdent(stmt2.Params[0]).Name != "fx" {
				continue
			}
			// Alright; start forming a target!  Neato.
			tgt := &Target{
				name: stmt2.Name.Name,
				stmt: stmt2,
			}
			// Process any other additional Known Arguments that are data holders.
			// For most of these, the "default" value will be examined; we can read those literals from here.
			// Unrecognized arguments are ignored, for future-proofness.
			for _, param := range stmt2.Params[1:] {
				switch extractIdent(param).Name {
				case "depends_on":
					expr2, ok := param.(*syntax.BinaryExpr)
					if !ok {
						continue
					}
					switch v := expr2.Y.(type) {
					case *syntax.ListExpr:
						for _, item := range v.List {
							lit, ok := item.(*syntax.Literal)
							if !ok || lit.Token != syntax.STRING {
								return nil, errDependsOnValueRestriction()
							}
							tgt.dependsOn = append(tgt.dependsOn, lit.Value.(string))
						}
					case *syntax.Literal:
						if v.Token != syntax.STRING {
							return nil, errDependsOnValueRestriction()
						}
						tgt.dependsOn = []string{v.Value.(string)}
					default:
						return nil, errDependsOnValueRestriction()
					}
				}
			}
			res = append(res, tgt)
		}
	}
	return
}

func errDependsOnValueRestriction() error {
	return serum.Errorf(wfxapi.EcodeScriptInvalid, "depends_on clause in target declaration may only use lists of string literals, or a single string literal")
}
