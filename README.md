# trakx

Efficient bittorrent tracker written in go.

## Install

Requires Go 1.12+

* `go get -v github.com/syc0x00/trakx`

## Setup

Note: If you have other go program with expvar in netdata you'll have to edit go_expvar.conf with `/etc/netdata/edit-config python.d/go_expvar.conf` and paste in the contents of `trakx_expvar.conf` while keeping your other programs config as well. The `install.sh` script will overwrite your other program otherwise.

* cd into Trakx in the gopath
* Open netdata python conf with `/etc/netdata/edit-config python.d.conf` and change `go_expvar` to `yes`
* Install netdata plugins with `sh netdata/install.sh`
* Run `sh setup.sh` to setup root directory and copy the config

## Updating

* Running `sh setup.sh` will update without overwriting config if it already exists

## Recommendations

Some clients will ignore the port that you set on your HTTP tracker and will instead hit 80. To stop this you can add  
`iptables -A INPUT -p tcp --dport 80 -m string ! --string "/announce?info_hash" --algo bm -j REJECT`

If you're going to be serving a lot of clients take a look at the sysctl tuning the resources section. Your sysctl will most likely need to be modified if you're using HTTP.

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
