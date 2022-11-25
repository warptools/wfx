package wfxapi

import (
	"github.com/serum-errors/go-serum"
)

const (
	// Errors that are wfx going wrong somehow:
	// (none yet)

	// Errors that are the script author's problem:
	EcodeScriptParsefail = "wfx-script-parsefail" // For syntax errors that starlark itself will reject -- before we even get to wfx-specific features.
	EcodeScriptInvalid   = "wfx-script-invalid"   // Generally, for things being used wrong.  Whereas parse errors are "wfx-script-unparsable".  Appear at runtime, but in scenarios where we feel the error is almost certainly static errors of usage.

	// Errors that appear at runtime:
	EcodeActionCmdExit = "wfx-action-error-cmdexit" // For when subprocesses exit nonzero.
)

// ErrorFxfileParse is an error constructor.
//
// Errors:
//
//   - wfx-script-parsefail -- always this.
func ErrorScriptParsefail(cause error, phase string) error {
	return serum.Error(EcodeScriptParsefail,
		serum.WithCause(cause),
		serum.WithDetail("phase", phase),
	)
}
