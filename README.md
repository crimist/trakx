# trakx

Efficient bittorrent tracker written in go.

## Install

Requires Go 1.12+

* `go get -v github.com/syc0x00/trakx`

* cd into trakx in the gopath
* Run `sh setup.sh` to install trakx

### Netdata install

Note: If you have other go program using expvar with netdata you'll have to manually add the trakx config, `install.sh` will overwrite `go_expvar.conf`.

* Open netdata python conf with `/etc/netdata/edit-config python.d.conf` and change `go_expvar` to `yes`
* Customize the url in `expvar.conf` if needed
* Install netdata plugins with `sh netdata/install.sh`

## Updating

* Running `sh setup.sh` will update without overwriting

## Build tags

* `-fast` will build without IP, seeds, and leeches metrics which will speed up trakx a bit
* `-pprof` will build with pprof, this is on in `setup.sh`

## Recommendations

If you're going to be serving a lot of clients take a look at the sysctl tuning the resources section. This is especially true if you're using the TCP tracker

## Resources

* [BitTorrent HTTP spec](https://wiki.theory.org/index.php/BitTorrentSpecification)
* [BitTorrent UDP spec](https://www.libtorrent.org/udp_tracker_protocol.html)
* [Sysctl tuning](https://wiki.mikejung.biz/Sysctl_tweaks) primarily for TCP
