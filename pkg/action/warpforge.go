package action

// Future work for wfx involves having an easy way to trigger `warpforge run`,
// as well as (especially) unpacking the resultant filesystems onto the host.

// For the sake of completeness, let's ask:
// Can we just solve problems without even using wfx, making the warpforge commands good enough at JSON?
//
// Let's examine a couple of possibilities:
//
// `warpforge unpack wire:thinglabel "$(warpforge run foobar.wf)" --path /tmp/foo` -- is this crazy?
//   I mean it passes json as a cli arg, so, yes, but otherwise... it's fine?
//     The unpack command can assume existence-is-updated-enough, or take other flags for that.
//     Having it parse json is... we might need another subcommand to make that intention clear, but, seems fine.
//        On further thought: I just put a "wire:" prefix in front of it and that seems superconsistent and easy to parse now.
//
// `warpforge unpack "$(warpforge run foobar.wf --output-select="thinglabel")" --path /tmp/foo` -- how about this?
//    If you needed more than one output out of an evaluation, it's kinda problematico...
//       But we have memoization.  So: fuck it.  Just eval the whole thing as many times as you need.  It's gonna no-op, right?  No problem.
//          (I'm not actually thrilled with that "fuck it".  It would work, but it almost feels like relying on hidden statefulness in a way; uncomfortable.)
//
// I think in either of these cases, it's still not going to trend towards great overall.
//
// We could support those use modes (and probably should, eventually).
// But they're still going to degenerate to someone needing to write shell scripts, and that's not a great joy.
// Leaving anything to shell script very rapidly results in 10 different people doing 10 different things, no matter how hard we might try to have one well-paved path.
// And if the user story involves shell scripts and involves wrangling multiple steps... you're going to start eventually getting...
//   - a desire to list all the possible steps and effects
//   - a way to kick some of them in the shins if the update detection fails
//   - _a way to tab-complete what to kick in the shins_, synthesizing the above two
//   - a way to do "all the things" and have partial success... without having miles and more miles of shell script
//   - probably a way to do some of it in parallel
// ... and by the time you start heaping on ANY of these goals, much less all of them, the shell script approach is getting painful to the point of nonviability very very rapidly.
//
// So.  Yes.  It seems like teaching wfx some simple built-in actions for direction warpforge will have high utility.
// (We might even implement them as simple exec internally, with minimal API coupling!  Even so: utility.)
