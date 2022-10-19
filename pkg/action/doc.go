/*
	The 'action' package contains all the built-in functions and effects wfx supports
	and makes available to users.

	One might say that a "target" is composed of calls to "actions".
	And, optionally, other starlark functions of the user's defining
	(but those in turn surely lead to some actions being called eventually,
	or they won't have much effect!).

	The most common action is probably "cmd" (which hands off work to a shell invocation),
	but there are many others.

	Actions take the form of a `starlark.Callable`, because that's what they need to be
	in order to be usable from starlark code, which is what the users write.
*/
package action
