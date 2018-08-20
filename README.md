# Trakx

Bittorrent tracker written in go.

## How

It uses the go default webserver and MySQL to hold the client list.
It currently uses my own bencode package but I will most likely move to something else eventually.

## Resources

[Basic spec](https://wiki.theory.org/index.php/BitTorrentSpecification) super helpful.

## Todo

Try using https://github.com/go-torrent/bencode
Test if this thing even works
Setup docker for ez testing
