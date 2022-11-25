package action

import (
	"fmt"
	"io"

	"github.com/serum-errors/go-serum"
	"go.starlark.net/starlark"

	"github.com/warptools/wfx/pkg/wfxapi"
)

// ActionPlan is something that a wfx script is ready to "do".
// The action described by the ActionPlan can have side effects.
// Several essential action plan constructors are exported to wfx scripts by default, and form the basis for most, if not all, effects.
//
// ActionPlan are produced by various constructor functions.
// They can be forced to execute manually by giving them to `do()`;
// otherwise, simply letting them fall to the ground as the unused result of a starlark statement will cause them to be executed(!).
// (Any use of the value -- either handing it to other functions, or simply assigning it to a variable -- prevents execution.)
//
// When executed, action will resolve to produce success or an error code;
// while in motion, the action can consume a stdin stream, and produce stdout and stderr streams.
// This sounds a lot like a posix process -- and indeed, it's modeled on it! -- and often even backed by a posix process!
// But can also be implemented without full subprocess execution.
//
// In starlark, execution of the action doesn't return anything.  It only succeeds or fails.
// If it fails, the program as a whole will halt.
// Other "controller" actions can wrap an action plan and modify it in order configure behaviors other than halt-on-error.
// The "pipe", "gather", "test", and "ignorantly" controller actions are all examples of things that can tweak the halt conditions in various ways.
// ("test" is an especially interesting one, in that it actually _does_ return a result -- a boolean.)
//
// Stdout (unless rewired by a piping) goes to dev null, unless verbose mode is on; then, it's decorated and emitted to the user.
// Stderr goes to dev null, unless the action fails; then it's likely to get included in the error message.
//
// Most ActionPlan are produced by builtin constructors, but they can be defined in Starlark code too.
// The main reason to consider doing so is to take advantage of the IO streaming conventions, so that the Starlark code can be composed with "pipe" and other action controllers.
// (Not yet implemented!)
type ActionPlan struct {
	// The big question is... do we make this golang-first?  Or starlark-first?
	// Leaning towards golang-first, because... we need firstclass awareness of the IO streams, and I want them to be fast by default.
	// We can have a system where starlark supplies a bunch of callbacks and stuff to build a custom action; but that probably can come later.

	// Should this be a superset of starlark.Value though?  Yeah, probably.
	// Callable, itself?  Not sure.  Probably... not?  Seems saner to me if we export a single blessed `do` function instead (and let magic do its job, everywhere else).
	//   Since actions are almost always produced by a constructor function, having them be immediately callables themselves can produce the `()()` situation, which... users don't seem to like to see very often.
	//   It also leaves open the whole space for the action to _choose_ to be callable, which might sometimes be desirable for configureme DSLs.

	// This whole "aggressively can't return a value" thing, I'm not sure about it.
	// On the off chance you _do_ want something returned... I'd probably rather have a `x = do(foobar())` than `foobar(effect=x)` ... the latter is a bit spooky, and also can never fit in a oneliner.
	//   But... maybe this doesn't matter, especially if we're starting to go in hard on shell-like piping features?
	//     Nope, same deal.  Piping into some kind of special "collect" value... possible!  but still means a multi-liner to make it, and then later bind it into the pipe, and then yet-later use it.
	//       Okay, still consider if we want a special "collect" value though -- Do we want `x = do(...)` to return a value?  Or the collected stdout?
	//         Well, you can bridge that gap by having a "collect" action that slurps stdin and turns it into a return value.
	//            Doesn't cover the ability to do anything but strings, but if you're using this, that's fine.  You have other options too.
	//         Returning a value in general means things can do other kinds than string.  Idk, seems maybe nicetohave.
	//         What's the deal with pipe, then?  Does it just always return the value from the last thing in the pipeline?  Guess so.  Odd that it then has to drop all the others on the floor, though.
	//            Should, uh, pipe pass the result from one thing into a positional arg (or "_pipe_arg" kwarg?) to the next?  As opposed to doing stream wiring?
	// tl;dr its unclear if the amount we just ganked shell pipe concepts is actually desirable.

	// `tweak(act, label="foobar", ignoreerror="All")` ?  or `label("foobar", act)` ?

	Name_     string      // what this action considers itself named.
	Label     string      // user defined label, which may appear in output log decorations.
	Details   interface{} // used in String() if provided
	Stdin     io.ReadCloser
	Stdout    io.WriteCloser
	Stderr    io.WriteCloser
	IsExec    bool             // if two siblings in a pipe are both true for this, use MkPipe to wire them together instead of application level buffer bouncing.
	Run       func() error     // note: do be prepared for this to be run in a goroutine; it very well might be (e.g. pipe will tend to do this); or, it might not.
	Ignorable func(error) bool // NYI.  we'll see if this is a plausible idea or not.
}

func (a *ActionPlan) Name() string { return "ActionPlan" + a.Name_ }
func (a *ActionPlan) String() string {
	if a.Details != nil {
		return a.Name() + fmt.Sprintf("{%v}", a.Details)
	} else {
		return a.Name() + "{...}"
	}
}
func (a *ActionPlan) Type() string          { return "<" + a.Name() + ">" }
func (a *ActionPlan) Freeze()               {}
func (a *ActionPlan) Truth() starlark.Bool  { return starlark.True }
func (a *ActionPlan) Hash() (uint32, error) { return 0, nil }

var _ starlark.Callable = (*Do)(nil)

type Do struct{}

func (a *Do) CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	switch len(args) {
	case 1:
		if ap, ok := args[0].(*ActionPlan); ok {
			return starlark.None, ap.Run()
		}
		// Do nothing if we weren't invoked on an ActionPlan; important to be silent, since we get blindly decorated on many things.
		//   FIXME: maybe break the silent chill mode into a separate function.  give the starlark code one that's loud.
		//     Can't actually hide the silent chill function, because of how we're putting all this together, but we can at least make it less obvious.  "_do" or something.
		return starlark.None, nil
	default:
		return starlark.None, serum.Errorf(wfxapi.EcodeScriptInvalid, "`do` expects exactly one positional arg")
	}
}

func (a *Do) Name() string          { return "do()" }
func (a *Do) String() string        { return "do()" }
func (a *Do) Type() string          { return "<do!>" }
func (a *Do) Freeze()               {}
func (a *Do) Truth() starlark.Bool  { return starlark.True }
func (a *Do) Hash() (uint32, error) { return 0, nil }
