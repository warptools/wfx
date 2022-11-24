package wfxapi

import (
	"github.com/serum-errors/go-serum"
)

// ErrorFxfileParse is an error constructor.
//
// Errors:
//
//   - wfx-error-fxfile-unparsable -- always this.
func ErrorFxfileParse(cause error, phase string) error {
	return serum.Error("wfx-error-fxfile-unparsable",
		serum.WithCause(cause),
		serum.WithDetail("phase", phase),
	)
}
