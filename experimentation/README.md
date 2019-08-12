# Cache

## Articles

* https://allegro.tech/2016/03/writing-fast-cache-service-in-go.html
* https://blog.dgraph.io/post/caching-in-go/

## Features

| Package                                          | Key         | Val         | TTL | Load File | Range/Iter    |
|--------------------------------------------------|-------------|-------------|-----|-----------|---------------|
| https://github.com/cornelk/hashmap               | interface{} | interface{} |     |           | X             |
| https://github.com/VictoriaMetrics/fastcache     | []byte      | []byte      |     | X         |               |
| https://github.com/coocood/freecache             | []byte      | []byte      | X   |           | X             |
| https://github.com/allegro/bigcache              | string      | []byte      | X   |           | X             |
| https://github.com/ReneKroon/ttlcache            | string      | interface{} | X   |           |               |
| https://github.com/patrickmn/go-cache            | string      | interface{} | X   | X         | Export to map |
| https://godoc.org/github.com/dgraph-io/ristretto | interface{} | interface{} |     |           |               |
| https://github.com/arriqaaq/zizou                | string      | interface{} | X   |           |               |
| https://github.com/goburrow/cache                | interface{} | interface{} | X   |           |               |

## Benchmarks

TODO
