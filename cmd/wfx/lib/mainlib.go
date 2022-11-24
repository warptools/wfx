package mainlib

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	cli "github.com/jawher/mow.cli"

	"github.com/warptools/wfx/pkg/wfx"
)

// Main runs the complete interpreter exactly as if the full program.
func Main(args []string, stdin io.Reader, stdout, stderr io.Writer) (exitcode int) {
	// Large TODO: this CLI library ignores our stdout and stderr params, and also tries to control rather than return exitcode.  We can't test anything off the happy path for args parsing until it does.
	app := cli.App("wfx", "the effect system for warpforge")
	app.Spec = "[[--dryrun] TARGETS... | --listtargets]"
	var (
		targets     = app.StringsArg("TARGETS", []string{}, "targets to refresh")
		dryrun      = app.BoolOpt("dryrun", false, "instead of acting, print names of targets that would be run, given the other arguments.")
		listtargets = app.BoolOpt("listtargets", false, "instead of acting, only list the available targets (one per line).")
	)
	app.Action = func() {
		fsys := os.DirFS(".")
		f, err := fsys.Open("make.fx")
		if err != nil {
			cli.Exit(19)
		}
		defer f.Close()
		bs, err := ioutil.ReadAll(f)
		if err != nil {
			cli.Exit(18)
		}

		// Evaluation happens in roughly three passes, each with their own opportunities to discover deeper kinds of errors:
		//  Pass 1: The syntax is parsed, and very high-level issues may be found -- then we attempt to discover all the targets.
		//  Pass 2: The syntax is interpreted more completely -- undefined references will now be noticed, if possible; but evaluation itself still does not yet occur (e.g. dynamic references won't be checked).
		//  Pass 3: Full evaluation -- now any remaining errors that are within the flow of execution will be found.
		mfxFile, err := wfx.ParseFxFile("make.fx", string(bs))
		if err != nil {
			cli.Exit(17)
		}

		if *listtargets {
			for _, target := range mfxFile.ListTargets() {
				fmt.Fprintf(stdout, "%s\n", target.Name())
			}
		} else {
			_ = dryrun // TODO support dryrun mode
			_ = targets

			evalCtx := wfx.EvalCtx{
				FxFile: mfxFile,
				Stdout: stdout,
				Stderr: stderr,
			}

			err := evalCtx.FirstPass()
			if err != nil {
				fmt.Fprintf(stderr, "%s\n", err)
				cli.Exit(14)
			}

			err = evalCtx.InvokeTargets(*targets)
			if err != nil {
				fmt.Fprintf(stderr, "%s\n", err)
				cli.Exit(12)
			}

		}
	}
	if err := app.Run(args); err != nil {
		panic(err)
	}

	return 0
}
