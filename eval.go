package wfx

import (
	"fmt"
	"io"

	"go.starlark.net/starlark"
)

func (x *MakefxFile) Eval(output io.Writer) (starlark.StringDict, error) {
	predef := starlark.StringDict{}

	prog, err := starlark.FileProgram(x.ast, predef.Has)
	if err != nil {
		return nil, err
	}

	thread := &starlark.Thread{
		Name: "eval",
		Print: func(thread *starlark.Thread, msg string) {
			fmt.Fprintln(output, msg)
		},
	}

	globals, err := prog.Init(thread, predef)
	if err != nil {
		return nil, err
	}

	return globals, err
}
