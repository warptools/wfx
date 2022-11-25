package wfx

import (
	"fmt"
	"io"

	"github.com/serum-errors/go-serum"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"

	"github.com/warptools/wfx/pkg/action"
	"github.com/warptools/wfx/pkg/wfxapi"
)

// EvalCtx shepherds an FxFile through the evaluation process.
// It holds a reference to the FxFile itself,
// gathers other configuration like top-level IO wiring,
// and also carries some state between phases of parse and evaluation.
//
// It's frankly not a terribly clear abstraction, because the FxFile *also* ends up mutated throughout the process
// (the starlark AST doesn't seem to have a convenient deepcopy method, so we just went with a dirty approach).
//
// There is also implicitly a statemachine here with only some orderings of calls being valid, but this is not enforced in code; caveat emptor.
type EvalCtx struct {
	FxFile *FxFile

	Stdout io.Writer
	Stderr io.Writer

	Globals starlark.StringDict // Assigned at the end of FirstPass.
}

var predef = starlark.StringDict{
	"_do":   &action.Do{},
	"cmd":   &action.CmdPlanConstructor{},
	"pipe":  &action.PipeControllerConstructor{},
	"panic": &action.PanicAction{},
}

// FirstPass performs only the first round eval -- which identifies targets.
// Starlark is evaluated here, but only whatever is used to initialize values.
// No functions are called nor fx targets evaluated -- that comes later.
//
// The global values at the end of the evaluation are stored in the EvalCtx.
//
// Errors:
//
//   - wfx-error-fxfile-unparsable -- if the resolve phase fails,
//      or if the second resolve after AST modification fails.
//   - wfx-eval-error -- if the init execution (computes globals) fails.
func (ctx *EvalCtx) FirstPass() error {
	// First pass: resolve everything.
	// This gives us some early error checking; it also populates all the `resolve.Binding` data into the AST, which is handy.
	if err := resolve.File(ctx.FxFile.ast, predef.Has, starlark.Universe.Has); err != nil {
		return wfxapi.ErrorFxfileParse(err, "resolve")
	}

	// This walk finds any statements which contain just a call expression, and rewrites them so they're wrapped in a certain magic function.
	// We use this to make some very fun DSL.
	// BEWARNED: this is quite mutation-happy.  Attempting to reuse an FxFile across EvalCtx's would be unsafe.
	// (There's no easy deep-copy method I'm aware of either; I'd happily use it to compartmentalize the phases better if there were.)
	syntax.Walk(ctx.FxFile.ast, func(n syntax.Node) bool {
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
	prog, dirtyerr := starlark.FileProgram(ctx.FxFile.ast, predef.Has)
	if dirtyerr != nil {
		return wfxapi.ErrorFxfileParse(dirtyerr, "resolve2") // n.b., internally, the "compile" process can't error... the only thing going on inside that can error is `resolve.File` again.
	}

	thread := &starlark.Thread{
		Name: "exploration",
		Print: func(thread *starlark.Thread, msg string) {
			fmt.Fprintln(ctx.Stdout, "during exploratory eval: "+msg)
		},
	}

	globals, dirtyerr := prog.Init(thread, predef)
	if dirtyerr != nil {
		return serum.Error("wfx-eval-error",
			serum.WithCause(dirtyerr),
			serum.WithDetail("phase", "init"),
		)
	}
	globals.Freeze()
	ctx.Globals = globals

	return nil
}

// InvokeTargets a graph of targets, starting with their dependencies.
func (ctx *EvalCtx) InvokeTargets(targetNames []string) error {
	// walk down the topo order.  keep a set of everything that's supported to be touched.
	todo := map[string]struct{}{}
	for _, t := range targetNames {
		todo[t] = struct{}{}
	}
	order, err := toposort(ctx.FxFile.targets)
	if err != nil {
		return err
	}
	for _, stepName := range order {
		if _, exists := todo[stepName]; !exists {
			continue
		}
		for _, depName := range ctx.FxFile.targetsByName[stepName].dependsOn {
			todo[depName] = struct{}{}
		}
	}
	// Now go backwards and do the things that should be done.
	for i := len(order) - 1; i >= 0; i-- {
		if _, exists := todo[order[i]]; !exists {
			continue
		}
		_, err := ctx.invokeOneTarget(order[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// invokeOneTarget calls exactly one target.  It does not call dependencies.
func (ctx *EvalCtx) invokeOneTarget(targetName string) (starlark.Value, error) {
	thread := &starlark.Thread{
		Name: "eval",
		Print: func(thread *starlark.Thread, msg string) {
			fmt.Fprintln(ctx.Stdout, "during target invokation (target="+targetName+"): "+msg)
		},
	}
	thread.SetLocal("stdout", ctx.Stdout)
	thread.SetLocal("stderr", ctx.Stderr)

	return starlark.Call(thread, ctx.Globals[targetName], []starlark.Value{starlark.None}, nil)
}
