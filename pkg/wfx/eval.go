package wfx

import (
	"fmt"
	"io"

	"github.com/warpsys/wfx/pkg/action"

	"go.starlark.net/starlark"
)

// FirstPass performs only the first round eval -- which identifies targets.
// Starlark is evaluated here, but only whatever is used to initialize values.
// No functions are called -- that comes later.
//
// The global values at the end of the evaluation are returned,
// but are also stored in the FxFile (Eval will use them).
func (x *FxFile) FirstPass(output io.Writer) (starlark.StringDict, error) {
	predef := starlark.StringDict{
		"cmd": &action.CmdAction{},
	}

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
