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



