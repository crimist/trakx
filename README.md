# Trakx

Bittorrent tracker written in go.

## Resources

* [Basic spec](https://wiki.theory.org/index.php/BitTorrentSpecification) - Super helpful
* [Gorm](https://github.com/jinzhu/gorm/) - ORM I used for DB
* [MySQL](https://www.mysql.com/) - DB I used

## Todo

* Move `PeerListCompact` and `PeerList` to `func (p *Peer)`
* Add testing
* Support Ipv6
  * http://www.bittorrent.org/beps/bep_0007.html
* Clean up `announce.go`