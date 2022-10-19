WarpForge eXtended / Warpforge Effects
======================================

**This is a very early prototype.  Please disregard.**

... unless you're really really really interested, and maybe wanna contribute.  Then feel free to read on.

---

`wfx` (it stands for either "Warpforge Extensions" or "Warpforge Effects" (mentally pronounce "fx" as "effects"))...

... is a "Make-like" tool, which is configured in a Starlark dialect.

Use it to:

1) invoke commands, and
2) efficiently maintain a filesystem that contains generated content that should be kept up-to-date.

---


Features
--------

- Declarative configuration file.
	- Feels a bit like `make`: you write targets; you say what actions to take to produce and maintain each target; like `make`, `wfx` will parse this declaration file, and make those targets easy to run.
	- Write Starlark (it's a python dialect).  Any function with a param named "fx" is a ==target== for `wfx`.  (E.g. `def install(fx):` means `wfx install` is gonna do whatever you say next.)
- Declare dependencies: Execution is a DAG -- evaluating a target causes its dependencies to be evaluated first; and all targets are evaluated exactly once, no matter how many times they might be depended on.
	- tl;dr: this is probably what you want -- it's the kind of behavior `make` gives you, too.
- Self-analyzing: run `wfx --listtargets` to get a list of all the possible actions you can take with the current config file.
- FUTURE: Run anything.  `cmd("foo --bar && baz | frob")` invokes a shell, and executes the `foo`, `baz`, and `frob` processes within it.
- FUTURE: Customize anything.  `cmd = cmd.customize(shell="/bin/fish")`, if you want to use the Fish shell instead of the default Bash, for example.
- FUTURE: Easily fetch data, so that bootstrapping other systems is easy.  Downloading (both from URLs, and from content-addressed sources!) is natively supported.  (No more worrying about whether `wget` or `curl` is installed!)
- FUTURE: Keep things up-to-date easily: targets can "own" some output filesystem paths, and can be trusted to keep them updated in the most efficient way possible (e.g., updating them when appropriate, while also no-op'ing _fast_ whenever possible).
	- Combined with the dependency DAG: each target having the ability to decide that it's already satisfactorily up-to-date means that whole graphs of dependencies can be very fast to (partially!) evaluate when repeated.
- FUTURE: Easily invoke `warpforge` -- use this anytime you have a task you want done in a sandbox, rather than having host effects!  (Or, if you just want the content-addressed memoization superpower!)

---


What's it look like?
--------------------

This is the smallest possible `make.fx` file:

```python
def hello(fx):
	cmd("echo hello!")
```

You can run that as `wfx hello`, and it'll... do what you expect. :)

A `make.fx` file containing multiple targets, which depend on each other, looks like this:

```python
def task_a(fx, depends_on=['task_b', 'task_c']):
	pass

def task_b(fx, depends_on=['task_c']):
	pass

def task_c(fx):
	pass
```

In that example, runing `wfx task_a` will run all three tasks; running `wfx task_b` will only run two.

You can declare that a target "owns" some files, and so should be invoked only when they're out of date:

```python
def owns_a_file(fx_files=['foo.a']):
	pass
```

(Note: not all features shown here are fully implemented (yet).  The examples are for syntax only.)

---


Goals
-----

1. Be a useful tool.  Help people get shit done.
	- `wfx` is aiming for productivity and pragmatism.
2. Be legible.
	- Even an unfamiliar person should be able to look at a `make.fx` file and say "oh.  Yes, I get it" and find the targets by eyeball, with minimal explanation.
3. Be easy to author.
	- Writing the minimal hello-world `make.fx` file should be so few keystrokes that you can bang one out from memory in 3 seconds or less.
4. Be predictable.  Be deterministic.
	- Absolutely no randomization behaviors allowed.
5. Be powerful.  Be a gateway.
	- Lots of other tools do lots of powerful things.  `wfx` should be good at invoking them to do what they do best (and not necessarily try to gain every feature itself!).
	- Looking at a `make.fx` file should be a good way to rapidly discover what actions can be taken on a filesystem, even (or especially) if you don't have existing familiarity with the other tools being invoked.

And also:

- Play nice with computation-addressable and content-addressable friends!
	- `wfx` is a tool for having effects on a filesystem, where the user controls the naming: e.g., `wfx` wants to let the user say "put the foobar tool in the path '/stuff/apps/foobar', then run it".
	- By contrast, [`warpforge`](http://warpforge.io/) is a tool for running things inside a container (no effects on the host), and works almost exclusively with content-addressed storage (meaning it doesn't accept user instructions about placing things in named paths).
	- ... This is a match made in heaven!  `wfx` can be used to invoke `warpforge`, take the outputs of `warpforge`'s work, and then declare where they should be unpacked on the host, and what happens next!

---


Comparison
----------

Compared to Make:

- similar: it's declarative, and has a DAG of targets.
- similar: you invoke it with target names -- `make install` and `wfx install` would both cause the respective tools to look for a target named "install", evaluate all its dependencies, and then evaluate that target.
- different: `wfx` uses Starlark syntax (aka, a Pythonic syntax).
- different: `wfx` has "phony" targets by default; it only has filesystem related targets if you explicitly say so.  (Removes many strange "bugs" that may occur if you try to use `make` in the ways most humans in post-year-2000 usually use it...)

The Starlark syntax has a lot of implications.  We think these are all significant improvements over Makefile syntax:

- No escaping hell!  Starlark syntax is what you'd expect from anything post-year-2000.  Strings are quoted.  Function calls look like function calls.  There's not a lot of surprises and not a lot of weirdness.
- You can use regular functions for composition!  No wild and crazy macros needed; functions get the job done.
- Sheer familiarity.  Even if you don't use Python, the syntax is rapid to learn and grok.
- Fewer gotchas.  In general, if you write something totally malformed in Starlark, it's a syntax error.  In make syntax, it's often unclear if you've escaped something correctly, or not... and it might not break... right away... even if it will, someday.

---


Show me more!
-------------

**You can see more examples in the [fixtures](./fixtures/) directory!**

All of the fixtures in that directory are working examples
(and not merely documentation, but are part of our automated regression tests).
You can copy them directly, and be sure they'll operate as shown.

---


Implementation and Contributing
-------------------------------

`wfx` is implemented in Golang.

The commands to build and test and generally wrangle the codebase are the usual: `go test ./...`, etc.

The CLI tool is produced by running `go install ./cmd/...` (and typically this will emit the result in `~/go/bin/`, unless you've set up something different).

We also use the [go-serum-analyzer](https://github.com/serum-errors/go-serum-analyzer) tool,
which performs static analysis on our error handling to make sure it's both well-documented and correctly-handled.

Tests are mostly stored in markdown files, in the [testmark](https://github.com/warpfork/go-testmark) format.
(This way, they serve as both living documentation, and as tests!)

If you'd like to contribute to `wfx`, please make sure:

1. Discussion and documentation should be included with any proposed change or additions!
2. The tests should still pass.
3. Any new features should include new tests for the feature!
4. The error analyser should also still give a passing grade.
5. Code should be correctly formatted, etc (e.g. `go fmt`).


License
-------

`wfx` is open source for you to use and share.  We hope it's helpful!

SPDX-License-Identifier: Apache-2.0 OR MIT
