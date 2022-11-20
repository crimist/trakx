# Trakx

Performance focused BitTorrent tracker supporting HTTP, UDP, IPv4 and IPv6.

- [Trakx](#trakx)
  - [❤️‍🔥 Instances](#️-instances)
  - [🚀 Install](#-install)
  - [🔧 Configuration](#-configuration)
    - [Configuration file](#configuration-file)
    - [Default configuration & webserver files](#default-configuration--webserver-files)
    - [Binding to privileged ports](#binding-to-privileged-ports)
    - [Netdata setup](#netdata-setup)
    - [Build Tags](#build-tags)
  - [📈 Performance](#-performance)

## ❤️‍🔥 Instances

Try Trakx for yourself! These instances are hosted on Oracles always free tier.

| Status       | Protocol  | Address                             |
|--------------|-----------|-------------------------------------|
| ✅Ok         | IPv4 UDP  | `udp://u4.trakx.crim.ist:1337`      |
| ✅Ok         | IPv6 UDP  | `udp://u6.trakx.crim.ist:1337`      |
| ❌Deprecated | IPv4 HTTP | `http://h4.trakx.crim.ist/announce` |
| ❌Deprecated | IPv6 HTTP | `http://h6.trakx.crim.ist/announce` |

## 🚀 Install

Go 1.19+ required.

```sh
git clone https://github.com/crimist/trakx && cd trakx

# install to go bin
go install .
trakx status # generates configuration

# or build
go build .
./trakx status # generates configuration
```

See [configuration](#configuration) and [netdata setup](#netdata-setup).

## 🔧 Configuration

### Configuration file

The configuration file can be found at `~/.config/trakx/trakx.yaml`.
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

Trakx attempts to load the config file from the following directories in order:

* `.`
* `~/.config/trakx/`

### Default configuration & webserver files

You can modify the default configuration and files served by the webserver in the `tracker/config/embeded/` folder.

**NOTE:** Trakx webserver will only serve files at their full path. `dmca` will 404, `dmca.html` will 200.

### Binding to privileged ports

To bind to privileged ports I recommend using `CAP_NET_BIND_SERVICE`. More information can be found [here](https://stackoverflow.com/a/414258/6389542).

```sh
$ sudo setcap 'cap_net_bind_service=+ep' ./trakx
$ TRAKX_TRACKER_HTTP_PORT=80 ./trakx run
2022-04-05T16:18:05.847-0700    INFO    HTTP tracker enabled    {"port": 80}
```

### Netdata setup

**Warning:** `install.sh` will overwrite `go_expvar.conf`. If you are using other expvar programs with netdata manually merge the two files.

* Run `/etc/netdata/edit-config python.d.conf`, change `go_expvar` to `yes`.
* Customize the url in `netdata/expvar.conf` if needed.
* Install netdata plugins with `cd netdata; ./install.sh`.

### Build Tags

You can build with different tags with `go build/install -tags <tag> .`

**Tags**
* `fast` will build without IP, seed, and leech metrics which will reduce cpu and memory usage
* `heroku` will build trakx for app engines. This means the controller will not be built and trakx will run immediately when the binary is executed. 

## 📈 Performance

The following metrics were collected on Heroku free tier running an HTTP tracker with the `fast` tag disabled.

Heroku dashboard:

![performance](img/performance.png)

Database stats:

![performance](img/stats.png)

Flamegraph:

![flame](img/flame.png)

Trakx has been optimized to use a little CPU time as possible. In most cases, almost all CPU time will be spent handing (negotiating/send/recv) connections, especially for TCP (HTTP).

Trakx has also been optimized to use minimal memory and is mostly limited by the go GC. In this example the GC runs every 2 minutes ([the forced GC period](https://github.com/golang/go/blob/895b7c85addfffe19b66d8ca71c31799d6e55990/src/runtime/proc.go#L4481-L4486)) at this level of traffic. The `inuse_space` delta from GC is 7.5% meaning this collection frequency would be sustained at `GOGC=8`.
