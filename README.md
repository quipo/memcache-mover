Memcache migration tool
=======================

This is a simple tool to help with copying the contents of a memcache cluster into a new one, helping with migrations.


It has two modes of operations:

1. List the slabs available in each source memcached server, list the keys(*) and start copying the entries to the destination cluster.  This mode is non-destructive, but limited in the number of entries that can be migrated(*)
2. Same as mode 1, but delete the migrate entries from the source cluster, and repeat until the source cluster is empty.  This mode WILL REMOVE ALL THE ENTRIES from the source cluster, so use with care.

Usage
-----

    memcache-mover -conf config.json

where `config.json` is a JSON file with the following properties:

    {
    	"memcache_src": {"addresses":["localhost:11211", "localhost:11212"]},
    	"memcache_dest": {"addresses":["localhost:11213","localhost:11213"]},
    	"move": false
    }

* `memcache_src` is the list of memcache servers in the source cluster (`host:port`)
* `memcache_dest` is the list of memcache servers in the destination cluster (`host:port`)
* `move` is the flag to enable mode #1 (i.e. "copy what you can" when set to `false`) vs mode #2 (i.e. "*move* data from source to destination, removing data from the source when done" when set to `true`)

How to build
------------

You need to have [Go installed](https://golang.org/doc/install).
From the project's directory, run

```shell
go build
```


Disclaimer
----------

This tool comes with no guarantees, and I'm not responsible for damage caused by it.


Author
------

Lorenzo Alberton

* Web: [http://alberton.info](http://alberton.info)
* Twitter: [@lorenzoalberton](https://twitter.com/lorenzoalberton)
* Linkedin: [/in/lorenzoalberton](https://www.linkedin.com/in/lorenzoalberton)



Copyright
---------

See [LICENSE](LICENSE) document