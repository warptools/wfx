package action

import (
	"fmt"

	"github.com/serum-errors/go-serum"
	"go.starlark.net/starlark"
	// "github.com/warpsys/wfx/pkg/wfx"
)

var _ starlark.Callable = (*CmdAction)(nil)

type CmdAction struct {
	interpreter string // "/bin/bash" by default.
}

func (a *CmdAction) CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	a.init()

	// This can be called in two modes:
	// 1. if it's got a string parameter, we're gonna exec something.
	// 2. if it's got only kwargs, we're gonna customize the cmd instance, and return a new one with those customizations.
	switch len(args) {
	case 0:
		// Then there'd better be some kwargs!
		if len(kwargs) < 1 {
			return starlark.None, a.errInvalidArgs("no args received")
		}
		// TODO
	case 1:
		// Exec time.
		thread.Print(thread, fmt.Sprintf("cmd: would invoke: %q\n", args[0]))
	default:
		return starlark.None, a.errInvalidArgs("received too many args")
	}

	return starlark.None, nil
}

func (a *CmdAction) Name() string          { return "cmd()" }
func (a *CmdAction) String() string        { return "cmd()" }
func (a *CmdAction) Type() string          { return "<action:cmd>" }
func (a *CmdAction) Freeze()               {}
func (a *CmdAction) Truth() starlark.Bool  { return starlark.True }
func (a *CmdAction) Hash() (uint32, error) { return 0, nil }

func (a *CmdAction) init() {
	if a.interpreter == "" {
		a.interpreter = "/bin/bash"
	}
}

func (CmdAction) errInvalidArgs(reason string) error {
	return serum.Errorf("wfx-error-invalid-args", "the cmd action needs either one string arg, or kwargs: "+reason) // FIXME: tweak packages until I can use a constant here without an import cycle problem.
}
