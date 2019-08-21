# trakx

Efficient bittorrent tracker written in go.

## Install

Requires Go 1.12+

```sh
git clone github.com/syc0x00/trakx
cd trakx
./setup.sh
```

### Netdata install

Note: If you have other go program using expvar with netdata you'll have to manually add the trakx config, `install.sh` will overwrite `go_expvar.conf`.

* Open netdata python conf with `/etc/netdata/edit-config python.d.conf` and change `go_expvar` to `yes`
* Customize the url in `expvar.conf` if needed
* Install netdata plugins with `sh netdata/install.sh`

## Updating

* Running `./setup.sh` will update without overwriting config

## Build tags

`-fast` will build without IP, seeds, and leeches metrics which will speed up trakx

```
8700K@5.0 DDR4-3000/16-18-18-38

Normal:
BenchmarkDrop-12                        30000000                57.9 ns/op
BenchmarkSave-12                        20000000                91.9 ns/op
BenchmarkSaveDrop-12                    10000000                210 ns/op
Fast:
BenchmarkDrop-12                        50000000                24.5 ns/op
BenchmarkSave-12                        20000000                62.7 ns/op
BenchmarkSaveDrop-12                    20000000                116 ns/op
```

## Notes

If you're going to be serving a lot of clients take a look at the sysctl tuning the resources section. This is especially true if you're using the TCP tracker

## Resources

* [BitTorrent HTTP spec](https://wiki.theory.org/index.php/BitTorrentSpecification)
* [BitTorrent UDP spec](https://www.libtorrent.org/udp_tracker_protocol.html)
* [Sysctl tuning](https://wiki.mikejung.biz/Sysctl_tweaks) primarily for TCP
