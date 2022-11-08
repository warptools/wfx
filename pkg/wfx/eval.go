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
		"cmd":   &action.CmdAction{},
		"pipe":  &action.PipeController{},
		"panic": &action.PanicAction{},
	}

	// We can't actually deconstruct `starlark.FileProgram()`.
	// It calls `compile.File`, and that package... is marked internal.
	// But... we can still touch up the AST before that.
	//
	// There is one fairly dark side to this: we can only recognize things by name.  Whoops.
	// ... WAIT, we can run FileProgram and it has side-effects. Such as populating all the bindings!
	// It's weird if we fuck with the ast after that and run compile again, sure, but... it...works!?

	// starlark.FileProgram(x.ast, predef.Has) // `resolve.File()` might be sufficient, but I'm not sure if that'll have effects on the syntax position info
	if err := resolve.File(x.ast, predef.Has, starlark.Universe.Has); err != nil {
		return nil, err
	}

	syntax.Walk(x.ast, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.CallExpr:
			if ident, ok := n.Fn.(*syntax.Ident); ok {
				if ident.Binding.(*resolve.Binding).Scope == 0x5 && ident.Name == "pipe" {
					fmt.Printf("::magic engaged\n")
					for _, arg := range n.Args {
						if call, ok := arg.(*syntax.CallExpr); ok {
							call.Args = append(call.Args, &syntax.BinaryExpr{
								Op: syntax.EQ,
								X:  &syntax.Ident{Name: "_pipe"},
								Y:  &syntax.Literal{Token: syntax.STRING, Value: "please"},
							})
						}
					}
				}
				// fmt.Printf(":: %#v --- %#v\n", n.Fn, ident.Binding)
				// so we can dive up through `ident.Binding`... though it's by index, into other tables, so it's a little complicated.
				// and i think that table... you have to hold onto the nearest enclosing DefStmt or File in order to have access to it.  which this walk callback doesn't really have an easy way to do.
				// (also the docs a bit wrong; you have to jump `.(*syntax.DefStmt).Function.(*resolve.Function).Locals` (or similar).)
				// You need to do all that jumping to get ... anything useful really?
				// The `*resolve.Binding` struct as a `.First` member which might be worth using, but probably doesn't remove most of the other fiddling here.
				//   wait what haha... actually this is the _only_ place you get an offramp back to `syntax.Ident`, lol?!
				//   but it's absent for globals or predeclared??  goodness gracious.  so in that case: you have to remember the name you saw earlier, and use that if e.g. resolve.Binding.Scope==0x5.  fun.
				// I think the `*resolve.Function` object probably is the root truth for unique defn, once you get there.
				//   But also it's really only got a name string and syntax position.  (and pointers back again to `Params []syntax.Expr` and `Body []syntax.Stmt`.)
				//     So that kind limits the easily accessable options if we wanted to let someone declare their own functions to be controllers and be able to detect that easily, so, take note.
				// Okay.  Uh.  Overall: this is quite a maze.  Maybe we should just start with only supporting this feature on globals.  Also if you rename them it'll fail.  Sorry.  Future work!
			}
			return true

			// Still wondering if we should just have any CallExpr that doesn't have its value be gathered... decorate it with a check for if it returns something with an EffectPlan marker method, and call the thing if it does.
			// I don't know if that's gonna fuck up syntax position info though, tbh, which is scary.

			// Oh yeah.  You should figure out the error handling story at the same time as all this.
			// If that turns out to require AST rewriting powers _anyway_, then... well.  That sets all the bars.
			//
			// Can we / should we use panics?
			// Whatever we do, it should work with explicit EffectPlan invocation too, not just with the wild powers.
		default:
			return true
		}
	})

	syntax.Walk(x.ast, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.ExprStmt:
			// I think this is the only case i want to decorate for auto-invoke.
			// Though it's a little weird, I think I can replace n.X with another call that wraps the original call, and that should do it.
			fmt.Printf("::ExprStmt found: %T\n", n.X)
			if c, ok := n.X.(*syntax.CallExpr); ok {
				n.X = &syntax.CallExpr{
					Fn:   &syntax.Ident{Name: "print"},
					Args: []syntax.Expr{c},
				}
			}
			// Yep.  Worked.
			return true
		default:
			return true
		}
	})

	prog, err := starlark.FileProgram(x.ast, predef.Has)
	if err != nil {
		return nil, err
	}

	syntax.Walk(x.ast, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.CallExpr:
			fmt.Printf(":::: callExpr now marked at: %#v\n", n) // you can eyeball the syntax position markers in this.  They don't move, despite all our fuckery above!  Neat!
			return true
		default:
			return true
		}
	})

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
func (x *FxFile) Eval(output io.Writer, targetNames []string) error {
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
		_, err := x.EvalOne(output, order[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// EvalOne calls exactly one target.  It does not call dependencies.
func (x *FxFile) EvalOne(output io.Writer, targetName string) (starlark.Value, error) {
	thread := &starlark.Thread{
		Name: "eval",
		Print: func(thread *starlark.Thread, msg string) {
			fmt.Fprintln(output, msg)
		},
	}

	return starlark.Call(thread, x.globals[targetName], []starlark.Value{starlark.None}, nil)
}
