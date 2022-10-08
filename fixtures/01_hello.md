hello
-----

Given a `make.fx` file:

[testmark]:# (hello/fs/make.fx)
```python
def foobar(fx):
	baz()

def frobnoz(fx):
	pass

def baz():
	pass
```

You can ask it what targets it has:

[testmark]:# (hello/sequence)
```sh
wfx --listtargets
```

And expect a reasonable answer:

[testmark]:# (hello/output)
```text
foobar
frobnoz
```

Note that only the functions that take an "`fx`" parameter are considered targets within the system.
Other functions are... just regular functions.



printing
--------

Let's do some printing of output, so we can see some basic target evaluation happening.

Here's our `make.fx` file:

[testmark]:# (printing/fs/make.fx)
```python
def foobar(fx):
	print("foobar")
	baz()

def frobnoz(fx):
	print("frob!")

def baz():
	print("baz")
```

We'll explicitly run both targets:

[testmark]:# (printing/sequence)
```sh
wfx foobar frobnoz
```

And lo:

[testmark]:# (printing/output)
```text
foobar
baz
frob!
```



dependencies
------------

Time for something fun: our first dependencies.

Here's our `make.fx` file:

[testmark]:# (dependencies/fs/make.fx)
```python
def foobar(fx, depends_on=["shuff"]):
	print("foobar")
	baz()

def frobnoz(fx):
	print("frob!")

def shuff(fx):
	print("shuff")

def baz(): # still just a regular function
	print("baz")
```

We'll only run one target this time:

[testmark]:# (dependencies/sequence)
```sh
wfx foobar
```

But we'll see the effects of several!

[testmark]:# (dependencies/output)
```text
shuff
foobar
baz
```

Notice that the output "shuff" comes before anything, including foobar.
Dependencies are analysed statically, and are evaluated in dependency-order,
outside of the usual starlark control flow.

The reason dependencies get this special treatment is that we make them
_run-once_, in any evaluation process.
