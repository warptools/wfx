package action

import (
	"go.starlark.net/starlark"
)

var _ starlark.Callable = (*PanicAction)(nil)

type PanicAction struct {
}

func (a *PanicAction) CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	panic("panic!")
	// We can just return errors, too.
	// return starlark.None, fmt.Errorf("an error")
	// I don't yet have a good grasp of why that might be preferable to just panicking.
}

func (a *PanicAction) Name() string          { return "panic()" }
func (a *PanicAction) String() string        { return "panic()" }
func (a *PanicAction) Type() string          { return "<action:panic>" }
func (a *PanicAction) Freeze()               {}
func (a *PanicAction) Truth() starlark.Bool  { return starlark.True }
func (a *PanicAction) Hash() (uint32, error) { return 0, nil }
