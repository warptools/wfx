package action

import (
	"fmt"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/serum-errors/go-serum"
	"go.starlark.net/starlark"
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
		panic("wfx: not yet implemented kwargs for cmd action")
	case 1:
		// Exec time.
		incantation := string(args[0].(starlark.String))
		thread.Print(thread, fmt.Sprintf("cmd: would invoke: %q\n", incantation))
		cmd := exec.Command(a.interpreter, "-c", incantation)
		return starlark.None, a.processExecError(cmd.Run(), incantation)
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
	return serum.Errorf("wfx-script-error-invalid-args", "the cmd action needs either one string arg, or kwargs: "+reason) // FIXME: tweak packages until I can use a constant here without an import cycle problem.
}

func (CmdAction) processExecError(original error, incantation string) error {
	switch e2 := original.(type) {
	case nil:
		return nil
	case *exec.ExitError:
		if e2.Exited() { // true means code; false means signal
			code := e2.ExitCode() // I don't think this exists on windows.  Ignoring for now; platform support can be "future work".
			return serum.Error("wfx-script-aborted-cmd-unhappy",
				serum.WithMessageTemplate("cmd {{cmd|q}} exited with code {{exitcode}}"),
				serum.WithDetail("cmd", incantation),
				serum.WithDetail("exitcode", strconv.Itoa(code)),
				serum.WithCause(original),
			)
		} else {
			signal := int(e2.Sys().(syscall.WaitStatus).Signal())
			return serum.Error("wfx-script-aborted-cmd-unhappy",
				serum.WithMessageTemplate("cmd {{cmd}} exited due to signal {{signal}}"), // Future: wish there was a "quote" fmt directive we could use here.
				serum.WithDetail("cmd", incantation),
				serum.WithDetail("signal", strconv.Itoa(signal)),
				serum.WithCause(original),
			)
		}
		// fun fact: you can report `e2.SystemTime()` and `e2.UserTime()`, too.  Might be worth making this loggable.
	default:
		panic(fmt.Errorf("wfx: unknown error from process exec library: %T %w", original, original))
	}
}
