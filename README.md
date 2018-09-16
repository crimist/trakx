# Trakx

Bittorrent tracker written in go.

## How

It uses the go default webserver and MySQL to hold the client list.
It currently uses my own bencode package but I will most likely move to something else eventually.

## Resources

* [Basic spec](https://wiki.theory.org/index.php/BitTorrentSpecification) - super helpful.

## Todo

* Try using https://github.com/go-torrent/bencode
* Docker for easy testing
* Support Ipv6
  * http://www.bittorrent.org/beps/bep_0007.html

## Done

* Support and test peers that join the tracker when they're already complete.
  * Wireshark it with debian torrent
* Comply with compact peer list
* LastSeen timestamp to remove peers with network issues
  * `go tracker.Clean()` should run every minuit and remove peers who haven't been seen in 1 hour
* Auto delete empty tables
* Logging
  * Using zap
* Support banning hashes
  * Mainly to comply with dmca

## Database layout

It uses the database `bittorrent` and creates a table for each torrent. Tables contain all the peers and is named `Hash_x` where `x = hex encoded info hash of torrent`.

### Descriptions

Torrent table:

```en
+----------+----------------------+------+-----+---------+-------+
| Field    | Type                 | Null | Key | Default | Extra |
+----------+----------------------+------+-----+---------+-------+
| id       | varchar(40)          | YES  |     | NULL    |       |
| peerKey  | varchar(20)          | YES  |     | NULL    |       |
| ip       | varchar(45)          | YES  |     | NULL    |       |
| port     | smallint(5) unsigned | YES  |     | NULL    |       |
| complete | tinyint(1)           | YES  |     | NULL    |       |
| lastSeen | bigint(20) unsigned  | YES  |     | NULL    |       |
+----------+----------------------+------+-----+---------+-------+
```

Banned hash table:

```en
+-------+-------------+------+-----+---------+-------+
| Field | Type        | Null | Key | Default | Extra |
+-------+-------------+------+-----+---------+-------+
| hash  | varchar(45) | YES  | UNI | NULL    |       |
+-------+-------------+------+-----+---------+-------+
```

## Testing

### Code tests

Just use the good old `go test -v`

### Stress

Use vegeta; it's good.

`echo "GET http://localhost:1337/scrape" | vegeta attack -rate 15000/s -duration=20s | vegeta plot > /tmp/plot.html; open plot.html`