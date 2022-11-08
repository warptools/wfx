controllers
===========

"controllers" are special functions that do some interesting things to control flow.
("pipe" is one of the commonly seen examples of a control.)


hello controllers
-----------------

Given some wfx script:

[testmark]:# (hello/fs/make.fx)
```python
def foobar(fx):
	pipe(
		cmd("echo hi"),
		cmd("tr h q"),
	)
	panic()
```

We'll run that one target:

[testmark]:# (hello/sequence)
```sh
wfx foobar
```

And it should act pretty much the same as "`echo hi | tr h q`" should in the shell:

[testmark]:# (hello/output)
```text
qi
```

This may not look like much, but it's a pretty wild feature.
Normally, a `cmd()` call will be evaluated immediately!
So what happened here instead?
The `pipe` function is a _controller_, which means it gets to use some special macro-like features.
These features let the `pipe` function tell the `cmd()` calls that they're about to be used by the pipe,
and this causes the `cmd()` calls to behave slightly differently:
they return `CmdPlan` objects instead of evaluating immediately.
The `pipe` function then arranges all their I/O wiring, and takes ownership of their invocation.


