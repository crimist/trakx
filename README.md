# Trakx

Efficient bittorrent tracker written in go.

## Netdata Setup

* Open python conf with `/etc/netdata/edit-config python.d.conf` and change `go_expvar` to `yes`
* Go inside `netdata/` and run `install.sh`

If you have other go program with expvar in netdata you'll have to edit go_expvar.conf with `/etc/netdata/edit-config python.d/go_expvar.conf` and paste in the contents of `trakx_expvar.conf` while keeping your other programs config as well. The `install.sh` script will overwrite your other program otherwise.

## pprof

To get a pprof profile and view it

```bash
go tool pprof -seconds=180 http://127.0.0.1:1338/debug/pprof/profile
go tool pprof -http=0.0.0.0:7331 /root/pprof/... # filename
```

Go 1.11+ recommended for flamegraph support

## Resources

* [BitTorrent HTTP spec](https://wiki.theory.org/index.php/BitTorrentSpecification)
* [BitTorrent UDP spec](https://www.libtorrent.org/udp_tracker_protocol.html)
* [Sysctl tuning](https://wiki.mikejung.biz/Sysctl_tweaks) primarily for TCP
