# Trakx

Bittorrent tracker written in go.

## Netdata Setup

Open python conf with `/etc/netdata/edit-config python.d.conf` and change `go_expvar` to `yes`

Open go expvar conf with `/etc/netdata/edit-config python.d/go_expvar.conf` and paste in contents of `trakx_expvar.conf`

Add Trakx alarms with `cp trakx_alarm.conf /etc/netdata/health.d`

Restart netdata with `netdata` to run with the new config

## pprof

```bash
go tool pprof -seconds=180 http://127.0.0.1:1338/debug/pprof/profile
pprof -http=nibba.trade:7331 /root/pprof/...
```

## Resources

* [Basic spec](https://wiki.theory.org/index.php/BitTorrentSpecification) - Protocol
* [Zap](https://godoc.org/go.uber.org/zap) - Logging
* [Sysctl tuning](https://wiki.mikejung.biz/Sysctl_tweaks)

## Todo

* Consider BEPs
  * IPv6 http://www.bittorrent.org/beps/bep_0007.html
  * External Address http://www.bittorrent.org/beps/bep_0024.html
  * Failure retry timer http://www.bittorrent.org/beps/bep_0031.html

