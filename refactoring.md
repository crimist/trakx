## Notes

* We're moving connection backups into the tracker runner using Marshal / Unmarshal

## to do

* Refactor controller / command line interface - KISS!
  * Look at caddy for inspiration
  * Include the ability to pass database in on startup with stdin (or a similar approach that will help match backup export functionality)

* Deal w/ every TODO

* Refactor the 'pools' package, see TODOs inside

## Ideas

* Determine how much memory to pre allocate for maps based on the last N runs maximum size
  * Create a routine that wakes up every minuite to check peak usage, write changes to a file that just has last N runs
