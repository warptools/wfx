package main

import (
	"os"

	"github.com/warptools/wfx/cmd/wfx/lib"
)

func main() {
	os.Exit(mainlib.Main(os.Args, os.Stdin, os.Stdout, os.Stderr))
}
