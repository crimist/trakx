# Trakx

Bittorrent tracker written in go.

## Netdata

`/etc/netdata/edit-config python.d.conf` change `go_expvar` to `yes`

`/etc/netdata/edit-config python.d/go_expvar.conf` add

```
Trakx:
  name : 'Trakx'
  url  : 'http://localhost:1338/debug/vars'
  collect_memstats: true
  extra_charts:
    - id: "ip_unique"
      options:
        name: IPs
        title: "Number of unique IPs"
        units: IPs
        family: IPs
        context: expvar.ip.unique
        chart_type: line
      lines:
        - {expvar_key: 'ip.unique', expvar_type: int, id: ips}
    - id: "hash_unique"
      options:
        name: hashes
        title: "Number of unique hashes"
        units: Hashes
        family: hashes
        context: expvar.hash.unique
        chart_type: line
      lines:
        - {expvar_key: 'hash.unique', expvar_type: int, id: hashes}
    - id: "peer_unique"
      options:
        name: peers
        title: "Number of unique peers"
        units: Peers
        family: peers
        context: expvar.peer.unique
        chart_type: line
      lines:
        - {expvar_key: 'peer.unique', expvar_type: int, id: peers}
    - id: "cleaned"
      options:
        name: cleaned
        title: "Number of peers cleaned since program ran"
        units: Peers
        family: cleaned
        context: expvar.tracker.cleaned
        chart_type: line
      lines:
        - {expvar_key: 'tracker.cleaned', expvar_type: int, id: cleaned}
    - id: "tracker_stats"
      options:
        name: stats
        title: "Tracker stats"
        units: Peers
        family: stats
        context: expvar.tracker
        chart_type: line
      lines:
        - {expvar_key: 'tracker.seeds', expvar_type: int, id: seeds}
        - {expvar_key: 'tracker.leeches', expvar_type: float, id: leeches}
```

## Resources

* [Basic spec](https://wiki.theory.org/index.php/BitTorrentSpecification) - Super helpful
* [Gorm](https://github.com/jinzhu/gorm/) - ORM I used for DB
* [MySQL](https://www.mysql.com/) - DB I used
* [Zap](https://godoc.org/go.uber.org/zap) - Logger

## Todo

* Use [go-chart](https://github.com/wcharczuk/go-chart) and generate graphs for index.html
* Clean up `announce.go`
* Add testing
* Support Ipv6
  * http://www.bittorrent.org/beps/bep_0007.html
* Proper zap config
