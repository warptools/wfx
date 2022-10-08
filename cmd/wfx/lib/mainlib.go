package mainlib

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/warpsys/wfx"
)

// Main runs the complete interpreter exactly as if the full program.
func Main(args []string, stdin io.Reader, stdout, stderr io.Writer) (exitcode int) {
	fsys := os.DirFS(".")
	f, err := fsys.Open("make.fx")
	if err != nil {
		return 19
	}
	defer f.Close()
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return 18
	}
	mfxFile, err := wfx.ParseMakefxFile("make.fx", string(bs))
	if err != nil {
		return 17
	}
	for _, target := range mfxFile.ListTargets() {
		fmt.Fprintf(stdout, "%s\n", target)
	}
	return 0
}
