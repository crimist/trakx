# Trakx

Bittorrent tracker written in go.

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