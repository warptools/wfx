package fixtures

import (
	"io"
	"io/fs"
	"os"
	"sort"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/warpfork/go-testmark"
	"github.com/warpfork/go-testmark/testexec"

	mainlib "github.com/warptools/wfx/cmd/wfx/lib"
)

func TestAll(t *testing.T) {
	matches, err := fs.Glob(os.DirFS("."), "*.md")
	if err != nil {
		panic(err)
	}
	sort.Strings(matches)
	for _, filename := range matches {
		testFile(t, filename)
	}
}

func testFile(t *testing.T, filename string) {
	t.Run(filename, func(t *testing.T) {
		doc, err := testmark.ReadFile(filename)
		if err != nil {
			t.Fatalf("spec file parse failed?!: %s", err)
		}

		doc.BuildDirIndex()
		tester := testexec.Tester{
			ExecFn: func(args []string, stdin io.Reader, stdout, stderr io.Writer) (exitcode int, oshit error) {
				code := mainlib.Main(args, stdin, stdout, stderr)
				return code, nil
			},
			Patches: &testmark.PatchAccumulator{},
			AssertFn: func(t *testing.T, actual, expect string) {
				qt.Assert(t, actual, qt.Equals, expect)
			},
		}
		for _, dir := range doc.DirEnt.ChildrenList {
			t.Run(dir.Name, func(t *testing.T) {
				tester.Test(t, dir)
			})
		}
	})
}
