# Trakx

Bittorrent tracker written in go.

## Netdata Setup

Open python conf with `/etc/netdata/edit-config python.d.conf` and change `go_expvar` to `yes`

Open go expvar conf with `/etc/netdata/edit-config python.d/go_expvar.conf` and paste in contents of `netdata_trakx.conf`

Add Trakx alarms with `cp trakx_alarm.conf /etc/netdata/health.d`

Restart netdata with `netdata` to run with the new config

## Resources

* [Basic spec](https://wiki.theory.org/index.php/BitTorrentSpecification) - Protocol
* [Zap](https://godoc.org/go.uber.org/zap) - Logging

## Todo

* Clean up `announce.go`
* Support Ipv6
  * http://www.bittorrent.org/beps/bep_0007.html
* Proper zap config

