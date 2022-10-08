package mainlib

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	cli "github.com/jawher/mow.cli"
	"go.starlark.net/starlark"

	"github.com/warpsys/wfx"
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
		mfxFile, err := wfx.ParseMakefxFile("make.fx", string(bs))
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
			// TODO probably do toposort code pretty immediately.  varying arg order shouldn't affect action here.

			globals, err := mfxFile.Eval(stdout)
			if err != nil {
				fmt.Fprintf(stderr, "%s\n", err)
				cli.Exit(14)
			}

			targetsByName := map[string]*wfx.Target{}
			for _, target := range mfxFile.ListTargets() {
				targetsByName[target.Name()] = target
			}

			// TODO shall we actually invoke targets?  manufactured statements?  or just take the globals and invoke?
			// Here we invoke the targets.  This is simple enough.
			thread := &starlark.Thread{
				Name: "eval",
				Print: func(thread *starlark.Thread, msg string) {
					fmt.Fprintln(stdout, msg)
				},
			}
			for _, targetStr := range *targets {
				target := targetsByName[targetStr]
				if target == nil {
					fmt.Fprintf(stderr, "no target named %q (try running `wfx --listtargets` to see the available targest)", targetStr)
					cli.Exit(12)
				}

				// FIXME we're blindly calling dependencies as a demo hack continues to unfold, but this is wrong and needs control
				for _, depName := range target.DependsOn() {
					dep := targetsByName[depName]
					if dep == nil {
						fmt.Fprintf(stderr, "target %q states dependency on nonexistent target %q", targetStr, depName)
						cli.Exit(13)
					}
					_, err := starlark.Call(thread, globals[dep.Name()], []starlark.Value{starlark.None}, nil)
					if err != nil {
						fmt.Fprintf(stderr, "eval error during target %q: %s", dep.Name(), err)
						cli.Exit(14)
					}
				}

				_, err := starlark.Call(thread, globals[target.Name()], []starlark.Value{starlark.None}, nil)
				if err != nil {
					fmt.Fprintf(stderr, "eval error during target %q: %s", target.Name(), err)
					cli.Exit(14)
				}
			}
		}
	}
	if err := app.Run(args); err != nil {
		panic(err)
	}

	return 0
}
