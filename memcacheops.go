package main

import (
	"net"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

// SleepOnTempError how long to wait before a retry in case of a network glitch
var SleepOnTempError = 20 * time.Millisecond

// MaxRetriesOnError how many times to retry in case of temporary connection issues
var MaxRetriesOnError = 3

func readItem(client *memcache.Client, key string) (*memcache.Item, error) {
	item, err := client.Get(key)
	// retry in case of a temporary error
	for i := 0; i < MaxRetriesOnError && isTempError(err); i++ {
		item, err = client.Get(key)
	}
	return item, err
}

func storeItem(client *memcache.Client, item *memcache.Item) error {
	err := client.Set(item)
	for i := 0; i < MaxRetriesOnError && isTempError(err); i++ {
		err = client.Set(item)
	}
	return err
}

func deleteItem(client *memcache.Client, key string) error {
	err := client.Delete(key)
	// retry in case of a temporary error
	for i := 0; i < MaxRetriesOnError && isTempError(err); i++ {
		err = client.Delete(key)
	}
	return err
}

func readBatch(client *memcache.Client, keys []string) (map[string]*memcache.Item, error) {
	results, err := client.GetMulti(keys)
	// retry in case of a temporary error
	for i := 0; i < MaxRetriesOnError && isTempError(err); i++ {
		results, err = client.GetMulti(keys)
	}
	return results, err
	// ErrCacheMiss
}

// check if it's a temporary error like a memcache network glitch
func isTempError(err error) bool {
	if nil == err {
		return false
	}
	if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
		//log.Println("cache error memcache network:", err.Error())
		time.Sleep(SleepOnTempError)
		return true
	}
	if strings.HasPrefix(err.Error(), "memcache: ") && "memcache: cache miss" != err.Error() {
		//log.Println("cache error memcache generic:", err.Error())
		time.Sleep(SleepOnTempError)
		return true
	}
	return false
}
