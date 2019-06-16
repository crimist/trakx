# Trakx

Bittorrent tracker written in go.

## Netdata Setup

`/etc/netdata/edit-config python.d.conf` change `go_expvar` to `yes`

`/etc/netdata/edit-config python.d/go_expvar.conf` paste in contents of `netdata_trakx.conf`

## Resources

* [Basic spec](https://wiki.theory.org/index.php/BitTorrentSpecification) - Protocol
* [Zap](https://godoc.org/go.uber.org/zap) - Logging

## Todo

* Clean up `announce.go`
* Support Ipv6
  * http://www.bittorrent.org/beps/bep_0007.html
* Proper zap config

