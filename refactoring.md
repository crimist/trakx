## Notes

* We're moving connection backups into the tracker runner using Marshal / Unmarshal

## to do

* Deal w/ every TODO

* Refactor the 'pools' package, see TODOs inside

## Ideas

* Determine how much memory to pre allocate for maps based on the last N runs maximum size
  * Create a routine that wakes up every minuite to check peak usage, write changes to a file that just has last N runs
