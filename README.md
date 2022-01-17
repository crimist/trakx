# trakx

Performance focused HTTP & UDP BitTorrent tracker.

## Install

Go 1.16+ required.

```sh
git clone github.com/crimist/trakx
cd trakx

# install to go bin
go install .
trakx start # starting...

# or you can build it
go build .
./trakx start # starting...
```
## Performance

The following performance information was collected on Heroku free tier running an HTTP tracker with the `fast` tag disabled.

<details>
<summary>Click to expand</summary>

### Heroku dashboard

![performance](img/performance.png)

### Database stats:

![performance](img/stats.png)

### Flamegraph:

![flame](img/flame.png)

#### **CPU**

Trakx has been optimized to use a little CPU time as possible. In most cases, almost all CPU time will be spent handing (negotiating/send/recv) connections, especially for TCP.

In this example the databases write (save peer information) function took 0.3% of total cpu time.

#### **Memory**

Again, trakx has been optimized for a minimal memory footprint and is mostly limited by the go GC.

In this example the GC runs every 2 minutes ([the forced GC period](https://github.com/golang/go/blob/895b7c85addfffe19b66d8ca71c31799d6e55990/src/runtime/proc.go#L4481-L4486)) at this level of traffic. The `inuse_space` delta from GC is 7.5% meaning this collection frequency would be sustained at `GOGC=8`.
</details>

## Configuration

### Modifying the config

The config can be found at `~/.config/trakx/trakx.yaml`.
You'll have to run the trakx controller at least once to generate this file.

Config settings can be overwritten with environment variables:

```sh
$ cat trakx.yaml
...
loglevel = error
...

$ TRAKX_LOGLEVEL=DEBUG trakx run
2022-01-16T19:52:25.627-0800    DEBUG   Debug level enabled, debug panics are on
...
```

Trakx attempts to load the config file from the following directories in this order `"./", "~/.config/trakx/", "./embeded/", "/app/embeded/"`.

As such you can override the config file located in `~/.config/trakx/` by placing a `trakx.yaml` in the directory trakx runs in (`./`).

### Modifying HTML pages / default config

You can modify the default config and HTML pages by editing them in the `embeded/` folder.

Make sure to run the following commands when finished to embed the files in trakx.

```sh
$ make setup # first time setup, only needed once
...
$ make embed
```

Building will now include the updated files.

### Build Tags

You can build with different tags with `go build/install -tags <tag> .`

**Tags**
* `fast` will build without IP, seed, and leech metrics which will reduce cpu and memory usage
* `heroku` will build trakx for app engines. This means the controller will not be built and trakx will run immediately when the binary is ran. 

## Netdata graphs install

**Warning:** the `install.sh` script will overwrite `go_expvar.conf`. If you are using other expvar programs with netdata you can manually merge the two files.

* Run `/etc/netdata/edit-config python.d.conf`, change `go_expvar` to `yes`.
* Customize the url in `netdata/expvar.conf` if needed.
* Install netdata plugins with `cd netdata; ./install.sh`.

## Notes

* If you're going to be serving a lot of clients on a non managed service it's recommended to perform [sysctl tuning](https://web.archive.org/web/20200706222821/https://wiki.mikejung.biz/Sysctl_tweaks). This is especially important if you're running a TCP tracker
* Trakx uses `unsafe` to read raw memory when backing up the database. This means that database saves could *technically* break between go versions if struct padding or internal slice structures are changed. This will likely never happen but if you wish to be safe you can change the encoding method to `encodeBinary()` to avoid this issue. This fix comes with a performance penelty. Backing up will take 3x more memory and be 7x slower.
