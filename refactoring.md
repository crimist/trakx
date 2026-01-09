## To dos

Before next release.

* Deal w/ every TODO
* refactor project layout - move stuff to internal/ so that code is cleanly seperated from packaging etc.
* Lots of manual testing, maybe write a stressor tool

## Ideas

* Determine how much memory to pre allocate for maps based on the last N runs maximum size
  * Create a routine that wakes up every minuite to check peak usage, write changes to a file that just has last N runs
