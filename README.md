Memcache migration tool
=======================

This is a simple tool to help with copying the contents of a memcache cluster into a new one, helping with migrations.


It has two modes of operations:

1. List the slabs available in each source memcached server, list the keys<b><sup id="a1">[(\*)](#f1)</sup></b> and start copying the entries to the destination cluster.  This mode is non-destructive, but limited in the number of entries that can be migrated<b><sup id="a1">[(\*)](#f1)</sup></b>
2. Same as mode 1, but delete the migrate entries from the source cluster, and repeat until the source cluster is empty.  This mode WILL REMOVE ALL THE ENTRIES from the source cluster, so use with care.

<b id="f1"><sup>(\*)</sup></b> Getting the keys from a memcached slab used to be limited to a response of max 2MB (see [issues/153](https://github.com/memcached/memcached/issues/153) and [items.c](https://github.com/memcached/memcached/blob/1174994a6cb977785fdf38aea915d23c1cfb5a56/items.c#L563)). There are now better ways of dumping all the items from memcache (see [LRU_Crawler](https://github.com/memcached/memcached/blob/master/doc/protocol.txt)) but this tool still provides a decent work-around for older memcached versions.

Usage
-----

    ./memcache-mover -conf config.json

where `config.json` is a JSON file with the following properties:

    {
    	"memcache_src": {"addresses":["localhost:11211", "localhost:11212"]},
    	"memcache_dest": {"addresses":["localhost:11213","localhost:11213"]},
    	"move": false
    }

* `memcache_src` is the list of memcache servers in the source cluster (`host:port`)
* `memcache_dest` is the list of memcache servers in the destination cluster (`host:port`)
* `move`: flag to enable mode #1 or #2:
   * set to `false` for mode #1, i.e. "**copy** what you can" (best effort due to limitation described <span id="a1">[above](#f1)</span>) and leave the source cache as-is (non destructive mode)
   * set to `true` for mode #2, i.e. "**move** data from source to destination, removing data from the source when done"

How to build
------------

You need to have [Go installed](https://golang.org/doc/install).
From the project's directory, run

```shell
go build
```


Disclaimer
----------

This tool comes with no guarantees, and I'm not responsible for any damage caused by it.

TODO
----

* Process different servers/slabs in parallel, with control over concurrency levels to limit how hard memcached is hit.
  I already tested a parallel version, but leaving this code operating in serial mode until I add throttling in, or it can be *too* effective ;-) and cause memcache to crash under load.
* Improve stats
* Implement looping logic in SlabProcessor
* Add functionality to read keys from a file instead of reading them from memcached directly
* Use GetMulti() to reduce network connection requests
* More docs / tests
* Add proper project structure, moving some utility functions to a library package


Author
------

Lorenzo Alberton

* Web: [http://alberton.info](http://alberton.info)
* Twitter: [@lorenzoalberton](https://twitter.com/lorenzoalberton)
* Linkedin: [/in/lorenzoalberton](https://www.linkedin.com/in/lorenzoalberton)



Copyright
---------

See [LICENSE](LICENSE) document