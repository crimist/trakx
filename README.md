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

**Note:** If you have other go program using expvar with netdata you'll have to manually add the trakx config, `install.sh` will overwrite `go_expvar.conf`.

* Open netdata python conf with `/etc/netdata/edit-config python.d.conf` and change `go_expvar` to `yes`
* Customize the url in `expvar.conf` if needed
* Install netdata plugins with `sh netdata/install.sh`

## Updating

* Running `./setup.sh` will update without overwriting config

## Build tags

`-fast` will build without IP, seeds, and leeches metrics which will speed up trakx

```
// 8700K@5.0 DDR4-3000/16-18-18-38
// BenchmarkDrop fasttracks either way

Normal:
BenchmarkSave-12                        18526666                64.7 ns/op
BenchmarkDrop-12                       100000000                10.8 ns/op
BenchmarkSaveDrop-12                     8140426               147   ns/op

Fast:
BenchmarkSave-12                        23938716                50.1 ns/op
BenchmarkDrop-12                       100000000                10.8 ns/op
BenchmarkSaveDrop-12                    11566442               104   ns/op
```

## Notes

If you're going to be serving a lot of clients take a look at the sysctl tuning the resources section. This is especially true if you're using the TCP tracker

## Resources

* [BitTorrent HTTP spec](https://wiki.theory.org/index.php/BitTorrentSpecification)
* [BitTorrent UDP spec](https://www.libtorrent.org/udp_tracker_protocol.html)
* [Sysctl tuning](https://wiki.mikejung.biz/Sysctl_tweaks) primarily for TCP
