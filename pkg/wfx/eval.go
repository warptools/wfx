package wfx

import (
	"fmt"
	"io"

	"github.com/warpsys/wfx/pkg/action"

	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// FirstPass performs only the first round eval -- which identifies targets.
// Starlark is evaluated here, but only whatever is used to initialize values.
// No functions are called -- that comes later.
//
// The global values at the end of the evaluation are returned,
// but are also stored in the FxFile (Eval will use them).
func (x *FxFile) FirstPass(output io.Writer) (starlark.StringDict, error) {
	predef := starlark.StringDict{
		"_do":   &action.Do{},
		"cmd":   &action.CmdPlanConstructor{},
		"pipe":  &action.PipeControllerConstructor{},
		"panic": &action.PanicAction{},
	}

	// First pass: resolve everything.
	// This gives us some early error checking; it also populates all the `resolve.Binding` data into the AST, which is handy.
	if err := resolve.File(x.ast, predef.Has, starlark.Universe.Has); err != nil {
		return nil, err
	}

	// This walk finds any statements which contain just a call expression, and rewrites them so they're wrapped in a certain magic function.
	// We use this to make some very fun DSL.
	syntax.Walk(x.ast, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.ExprStmt:
			//fmt.Printf("::ExprStmt found: %T\n", n.X)
			if c, ok := n.X.(*syntax.CallExpr); ok {
				n.X = &syntax.CallExpr{
					Fn:   &syntax.Ident{Name: "_do"},
					Args: []syntax.Expr{c},
				}
			}
		}
		return true
	})

	// Resolves the whole AST again (we've modified it!) and compiles the program.  Almost ready to run.
	prog, err := starlark.FileProgram(x.ast, predef.Has)
	if err != nil {
		return nil, err
	}

	thread := &starlark.Thread{
		Name: "exploration",
		Print: func(thread *starlark.Thread, msg string) {
			fmt.Fprintln(output, msg)
		},
	}

	globals, err := prog.Init(thread, predef)
	if err != nil {
		return nil, err
	}
	globals.Freeze()
	x.globals = globals

	return globals, err
}

// Eval calls each target, and their dependencies.
func (x *FxFile) Eval(stdout, stderr io.Writer, targetNames []string) error {
	// walk down the topo order.  keep a set of everything that's supported to be touched.
	todo := map[string]struct{}{}
	for _, t := range targetNames {
		todo[t] = struct{}{}
	}
	order, err := toposort(x.targets)
	if err != nil {
		return err
	}
	for _, stepName := range order {
		if _, exists := todo[stepName]; !exists {
			continue
		}
		for _, depName := range x.targetsByName[stepName].dependsOn {
			todo[depName] = struct{}{}
		}
	}
	// Now go backwards and do the things that should be done.
	for i := len(order) - 1; i >= 0; i-- {
		if _, exists := todo[order[i]]; !exists {
			continue
		}
		_, err := x.EvalOne(stdout, stderr, order[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// EvalOne calls exactly one target.  It does not call dependencies.
func (x *FxFile) EvalOne(stdout, stderr io.Writer, targetName string) (starlark.Value, error) {
	thread := &starlark.Thread{
		Name: "eval",
		Print: func(thread *starlark.Thread, msg string) {
			fmt.Fprintln(stdout, msg)
		},
	}
	thread.SetLocal("stdout", stdout)
	thread.SetLocal("stderr", stderr)

	return starlark.Call(thread, x.globals[targetName], []starlark.Value{starlark.None}, nil)
}
