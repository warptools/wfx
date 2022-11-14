controllers
===========

"Controllers" are actions that take other actions as parameters, and influence their execution in some way.

For example, "pipe" is one of the commonly seen examples of a controller.


hello controllers
-----------------

Given some wfx script:

[testmark]:# (hello/fs/make.fx)
```python
def foobar(fx):
	pipe(
		cmd("echo hi | tee /tmp/hihi"),
		cmd("tr h q"),
	)
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

_So what happened here instead?_

Because the `cmd()` calls are being used as arguments to another function
(the `pipe()` function -- the controller in this demo),
instead of being executed immediately, they're returning an "action plan" object instead.
The `pipe` function then gets to receive those objects, and tweak them a bit:
in `pipe`'s case, it does some I/O wiring, so the data will feed from one command to the next.
Then, since `pipe` has received "action plan" objects, it's now it's job take ownership of their invocation, too...
so, it does so, and in `pipe`'s case, that means running them in parallel.
