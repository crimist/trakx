# Trakx

Efficient bittorrent tracker written in go.

## Netdata Setup

Open python conf with `/etc/netdata/edit-config python.d.conf` and change `go_expvar` to `yes`

Open go expvar conf with `/etc/netdata/edit-config python.d/go_expvar.conf` and paste in contents of `trakx_expvar.conf` (`cp trakx_expvar.conf /etc/netdata/python.d/go_expvar.conf` if you don't have any other go progs)

Add Trakx alarms with `cp trakx_alarm.conf /etc/netdata/health.d`

Restart netdata with `netdata` to run with the new config

## pprof

```bash
go tool pprof -seconds=180 http://127.0.0.1:1338/debug/pprof/profile
pprof -http=0.0.0.0:7331 /root/pprof/...
```

## Resources

* [HTTP spec](https://wiki.theory.org/index.php/BitTorrentSpecification)
* [UDP spec](https://www.libtorrent.org/udp_tracker_protocol.html)
* [Sysctl tuning](https://wiki.mikejung.biz/Sysctl_tweaks)
