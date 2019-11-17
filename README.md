# trakx

Bittorrent tracker generally focused on speed and efficiency.

![performance](img/performance.png)

HTTP only with extensive metrics enabled. Memory usage was 150MB with this load.

![flame](img/flame.png)

As you can see basically all the CPU usage is overhead from handling so many TCP connections. For example in this flamegraph the databases save function was only 0.3% of the flamegraphs time.

## Install

Go 1.13+ recommended for `sync.Pool` and `sync.RMutex` optimizations.

```sh
git clone github.com/crimist/trakx
./trakx/setup.sh
trakx status # "Trakx is not running"
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

## Notes

If you're going to be serving a lot of clients take a look at the sysctl tuning the resources section. This is especially true if you're using the TCP tracker

## Resources

* [BitTorrent HTTP spec](https://wiki.theory.org/index.php/BitTorrentSpecification)
* [BitTorrent UDP spec](https://www.libtorrent.org/udp_tracker_protocol.html)
* [Sysctl tuning](https://wiki.mikejung.biz/Sysctl_tweaks) primarily for TCP
