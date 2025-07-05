## Notes

* We're moving connection backups into the tracker runner using Marshal / Unmarshal

## to do

* Refactor the 'pools' package, see TODOs inside
 
* Refactor controller / command line interface - KISS!
  * Look at caddy for inspiration

* Deal w/ every TODO

## Ideas

* Determine how much memory to pre allocate for maps based on the last N runs maximum size
  * Create a routine that wakes up every minuite to check peak usage, write changes to a file that just has last N runs
