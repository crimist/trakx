## Notes

* We're moving connection backups into the tracker runner using Marshal / Unmarshal

## Current

* Refactor config
  * Redo the layout in trakx.yaml, then the actual go side
  * Move all config validity checks into Configuration.Validate() (some are in main.Run())

## Future

* Refactor the 'pools' package, see TODOs inside
* Refactor controller / command line interface - KISS!
* Deal w/ every TODO

## Changes

### Controller

* Look at caddy for inspiration
* Add arguments for providing config & pidfile
* Make default paths for config and pid resepect XDG standard

### Config

* Calculate logical defaults by default, this will requirement benchmarks
  * Maybe ~1-1.5x nproc worker threads by default?

### General

* Determine how much memory to pre allocate for maps based on the last N runs maximum size
  * Create a routine that wakes up every minuite to check peak usage, write changes to a file that just has last N runs
