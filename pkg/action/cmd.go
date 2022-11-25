package action

import (
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/serum-errors/go-serum"
	"go.starlark.net/starlark"

	"github.com/warptools/wfx/pkg/wfxapi"
)

var _ starlark.Callable = (*CmdPlanConstructor)(nil)

type CmdPlanConstructor struct {
	interpreter string // "/bin/bash" by default.

	// we'll probably put a callable field on this for a "customize" method.  future work.
}

func (a *CmdPlanConstructor) CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	a.init()

	switch len(args) {
	case 1:
		incantation := string(args[0].(starlark.String))
		ap := &ActionPlan{
			Name_:   "Cmd",
			Details: incantation,
			IsExec:  true,
		}
		ap.Run = func() error {
			cmd := exec.Command(a.interpreter, "-c", incantation)
			// Copy any IO handles that have been mutated onto the ActionPlan into the exec cmd var.
			// Otherwise get default IO handles from the thread locals (which currently means more or less "all the way to the user terminal").
			if ap.Stdin != nil {
				cmd.Stdin = ap.Stdin
			}
			if ap.Stdout != nil {
				cmd.Stdout = ap.Stdout
				defer ap.Stdout.Close()
			} else {
				cmd.Stdout = thread.Local("stdout").(io.Writer)
			}
			if ap.Stderr != nil {
				cmd.Stderr = ap.Stderr
			} else {
				cmd.Stderr = thread.Local("stderr").(io.Writer)
			}
			return a.processExecError(cmd.Run(), incantation)
		}
		return ap, nil
	default:
		return starlark.None, serum.Errorf(wfxapi.EcodeScriptInvalid, "`cmd` actions expect exactly one positional arg, which should be a string")
	}
}

func (a *CmdPlanConstructor) Name() string          { return "cmd()" }
func (a *CmdPlanConstructor) String() string        { return "cmd()" }
func (a *CmdPlanConstructor) Type() string          { return "<actionPlanConstructor:cmd>" }
func (a *CmdPlanConstructor) Freeze()               {}
func (a *CmdPlanConstructor) Truth() starlark.Bool  { return starlark.True }
func (a *CmdPlanConstructor) Hash() (uint32, error) { return 0, nil }

func (a *CmdPlanConstructor) init() {
	if a.interpreter == "" {
		a.interpreter = "/bin/bash"
	}
}

func (CmdPlanConstructor) processExecError(original error, incantation string) error {
	switch e2 := original.(type) {
	case nil:
		return nil
	case *exec.ExitError:
		if e2.Exited() { // true means code; false means signal
			code := e2.ExitCode() // I don't think this exists on windows.  Ignoring for now; platform support can be "future work".
			return serum.Error(wfxapi.EcodeActionCmdExit,
				serum.WithMessageTemplate("cmd {{cmd|q}} exited with code {{exitcode}}"),
				serum.WithDetail("cmd", incantation),
				serum.WithDetail("exitcode", strconv.Itoa(code)),
			)
		} else {
			signal := int(e2.Sys().(syscall.WaitStatus).Signal())
			return serum.Error(wfxapi.EcodeActionCmdExit,
				serum.WithMessageTemplate("cmd {{cmd|q}} exited due to signal {{signal}}"),
				serum.WithDetail("cmd", incantation),
				serum.WithDetail("signal", strconv.Itoa(signal)),
			)
		}
		// fun fact: you can report `e2.SystemTime()` and `e2.UserTime()`, too.  Might be worth making this loggable.
	default:
		panic(fmt.Errorf("wfx: unknown error from process exec library: %T %w", original, original))
	}
}
