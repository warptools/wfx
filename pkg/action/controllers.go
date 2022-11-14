package action

import (
	"io"
	"sync"

	"github.com/serum-errors/go-serum"
	"go.starlark.net/starlark"
)

/*
"Controllers" are actions that take other actions as parameters, and influence their execution in some way.

For example:

	pipe(
		cmd("something"),
		cmd("grep 'some words'"),
	)

... will act like a single action, where both cmd actions are run under the control of the pipe action;
and the pipe action will wire the inputs and outputs of those commands together before it launches them.

*/

/*
Some hypothetical features (not yet all implemented):

	- pipe(a,b,...) -- equiv of shell `a | b | ...`.
	- gather(a,b,...) -- equiv of shell `{ a ; b ; ... }`.
	- pipe(gather(a, pipe(b,c)), d) -- a valid construction, just like shell `{a; b|c;} | d` is!

*/

/*
	Okay, now on how to actually pull this off:

	Option1: "_chain=true" as kwargs in the depths.
		pro: simple.
		con: syntax boilerplate heavy. Easy to write invalid combinations by skipping some keystrokes.
		pro: one step depth effect (no chance of spooky-action-at-a-distance).
	Option2: controller funcs cause each arg to be wrapped in a "flipper thunk" (e.g. "pipe" will cause each arg that's a func to get wrapped in "__pipe__flipper") and then you do TLS or whatever you want in there.
		pro: dank. General.
		con: complex af to implement.
		con: leaves TLS reading and protocol negotiation in the hands of other library functions. (Doesn't smack frontline users; does smack reusable part developers. Maybe an okay trade, but notable.)
		con: can be a source of spooky action at SIGNIFICANT distance, if someone does the TLS protocol "wrong". Major badness for debugability. (Especially given that our major intention with this is changing when majorly side-effecting actions actually occur!)
			maybe mitigate by providing core lib funcs that peek properly, but, still only partial/hopeful/you-have-to-hold-it-right mitigation!
	Option3: decorate each top level statement in a wfx target with an implicit "do(...)".
		con: pretty one-off.
		con: means any use beyond the basics will surprise users (by not doing anything).
		pro: probably the second-easiest to implemented.
		pro: every action can return a "plan" object.  consistent.
	Option4: hybrid of Option1, but caused by mechanism like Option2!
		pro: syntax goals attained
		pro: consistently puts the info in all child positions (unlike Option1!)
		pro: one step depth effect (no chance of spooky-action-at-a-distance).
		pro: general.

	Okay... So, there's a clear winner, actually!  Option4.

	...
	...
	... and, now that I've actually tried to implement it, the winds feel different again.
	Option4 is possible, and a proof-of-concept has been made.  It seems viable.  But it still pushes a decent amount of complexity into quite a few directions at once.
	Option3 actually turned out to be VERY easy to implement, and also implementable in a very general and consistent way (it was possible to make it work on _any_ dangling call, not just on statements in top-level targets) which makes it more desiable overall.

	Gonna wanna sleep on it again, but Option3 now seems likely to be more compelling.  Option4 can stay in our pocket (and perhaps still be usable for other purposes, later).
*/

var _ starlark.Callable = (*PipeControllerConstructor)(nil)

type PipeControllerConstructor struct{}

func (a *PipeControllerConstructor) CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// Prep all the wiring first.
	var prev *ActionPlan
	for i, arg := range args {
		ap, ok := arg.(*ActionPlan)
		if !ok {
			return starlark.None, serum.Errorf("wfx-script-error-invalid-args", "`pipe` expects all positional args to be an ActionPlan")
		}
		if i > 0 {
			r, w := io.Pipe()
			ap.Stdin = r
			prev.Stdout = w
		}
		prev = ap
	}

	// Okay: let's go.
	results := make([]error, len(args))
	var firstError error
	var mutex sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(args))
	for i, arg := range args {
		i, arg := i, arg
		go func() {
			//fmt.Printf("::: launching %s\n", arg.(*ActionPlan).String())
			results[i] = arg.(*ActionPlan).Run()
			// We attempt to keep the first error.
			// This is pretty best-effort.  Fundamentally, there's a lack of synchronization at the kernel interface which lets us reliably know which process exited first.
			// We *hope* to get it close enough, and we *hope* that the first error we see is the most meaningful one (e.g., the one that's not complaining about pipes that are broken by other commands already exiting unexpectedly!),
			// but it's impossible to know for sure.
			mutex.Lock()
			if firstError == nil {
				firstError = results[i]
			}
			mutex.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()
	// TODO this should compose a list of errors.  Also represent them... how?  Do we return a serum error into starlark value?
	return starlark.None, firstError
}

func (a *PipeControllerConstructor) Name() string          { return "pipe()" }
func (a *PipeControllerConstructor) String() string        { return "pipe()" }
func (a *PipeControllerConstructor) Type() string          { return "<action:pipe>" }
func (a *PipeControllerConstructor) Freeze()               {}
func (a *PipeControllerConstructor) Truth() starlark.Bool  { return starlark.True }
func (a *PipeControllerConstructor) Hash() (uint32, error) { return 0, nil }
