# Trakx

Bittorrent tracker written in go.

## How

It uses the go default webserver and MySQL to hold the client list.
It currently uses my own bencode package but I will most likely move to something else eventually.

## Resources

* [Basic spec](https://wiki.theory.org/index.php/BitTorrentSpecification) - super helpful.

## Todo

* Try using https://github.com/go-torrent/bencode
* Docker for easy testing
* Prod vs dev logging
* LastSeen timestamp to remove peers with network issues
  * `go tracker.Clean()` should run every minuit and remove peers who havn't been seen in 1 hour
* Support Ipv6

## Done

* Support and test peers that join the tracker when they're already complete.
  * Wireshark it with debian torrent
* Comply with compact peer list