# Trakx

Bittorrent tracker written in go.

## Netdata

`/etc/netdata/edit-config python.d.conf` change `go_expvar` to `yes`

`/etc/netdata/edit-config python.d/go_expvar.conf` add

```
Trakx:
  name : 'Trakx'
  url  : 'http://localhost:1338/debug/vars'
  collect_memstats: false
  extra_charts:
    - id: "tracker_hits_per_sec"
      options:
        name: hits per second
        title: "Number of hits per second"
        units: Hits/s
        family: performance
        context: expvar.tracker.hitspersec
        chart_type: line
      lines:
        - {expvar_key: 'tracker.hitspersec', expvar_type: int, id: hits/s}
    - id: "tracker_hits"
      options:
        name: hits
        title: "Total number of hits"
        units: Hits
        family: performance
        context: expvar.tracker.hits
        chart_type: line
      lines:
        - {expvar_key: 'tracker.hits', expvar_type: int, id: hits}
    - id: "tracker_peers"
      options:
        name: peers
        title: "Number of peers"
        units: Peers
        family: peers
        context: expvar.tracker.peers
        chart_type: line
      lines:
        - {expvar_key: 'tracker.peers', expvar_type: int, id: peers}
     - id: "tracker_complete"
      options:
        name: completed
        title: "Number of seeds / leeches"
        units: Peers
        family: peers
        context: expvar.tracker
        chart_type: line
      lines:
        - {expvar_key: 'tracker.seeds', expvar_type: int, id: seeds}
        - {expvar_key: 'tracker.leeches', expvar_type: float, id: leeches}
    - id: "tracker_ips"
      options:
        name: IPs
        title: "Number of unique IPs"
        units: IPs
        family: IPs
        context: expvar.tracker.ips
        chart_type: line
      lines:
        - {expvar_key: 'tracker.ips', expvar_type: int, id: ips}
    - id: "tracker_hashes"
      options:
        name: hashes
        title: "Number of unique hashes"
        units: Hashes
        family: hashes
        context: expvar.tracker.hashes
        chart_type: line
      lines:
        - {expvar_key: 'tracker.hashes', expvar_type: int, id: hashes}
    - id: "tracker_cleaned"
      options:
        name: cleaned
        title: "Number of peers cleaned since program ran"
        units: Peers
        family: cleaned
        context: expvar.tracker.cleaned
        chart_type: line
      lines:
        - {expvar_key: 'tracker.cleaned', expvar_type: int, id: cleaned}
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
