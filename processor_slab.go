package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

// SlabProcessor migrates data from a single slab, by reading a chunk of keys and copying them
// over one by one. If the `Move` flag is set, successfully migrated entries are removed from
// the source node
type SlabProcessor struct {
	Name                  string
	Slab                  int64
	BatchSize             int64
	MemcacheSrcAddress    string
	MemcacheDestAddresses []string

	Move bool

	initialised bool
	stats       Stats
}

// NewSlabProcessor returns a new processor which can own the operations from a specific memcached node/slab
func NewSlabProcessor(srcMemcacheAddr string, destMemcacheAddr []string, slab int64, batchSize int64, move bool) *SlabProcessor {
	p := &SlabProcessor{
		Name:                  fmt.Sprintf("Slab Processor on host %s - slab %d (batch size: %d)", srcMemcacheAddr, slab, batchSize),
		Slab:                  slab,
		BatchSize:             batchSize,
		MemcacheSrcAddress:    srcMemcacheAddr,
		MemcacheDestAddresses: destMemcacheAddr,
		Move: move,
	}

	p.stats = Stats{
		ProcessorName: p.Name,
	}

	return p
}

// RunOnce reads the max amount of keys from the given slab
// (limited to https://github.com/memcached/memcached/issues/153)
// and migrates the corresponding items to the destination cluster.
// When the `Move` flag is set, the successfully migrated items are removed from the source server,
// making it possible to request new batches in a loop until all the data is migrated.
func (p *SlabProcessor) RunOnce() {
	if !p.initialised {
		p.stats.StartTime = time.Now()
	}

	chKeys := make(chan string, 100)
	chCopiedKeys := make(chan string, 100)
	resReaderCh := make(chan Stats)
	resWriterCh := make(chan Stats)
	resEraserCh := make(chan Stats)

	// read keys directly from the source memcache instance
	conn, err := net.Dial("tcp", p.MemcacheSrcAddress)
	if err != nil {
		log.Fatal(err)
	}

	chItems := make(chan *memcache.Item, 100)
	go p.runReader([]string{p.MemcacheSrcAddress}, chKeys, chItems, resReaderCh)
	go p.runWriter(p.MemcacheDestAddresses, chItems, chCopiedKeys, resWriterCh)
	if p.Move {
		go p.runEraser([]string{p.MemcacheSrcAddress}, chCopiedKeys, resEraserCh)
	} else {
		go p.noOp([]string{p.MemcacheSrcAddress}, chCopiedKeys, resEraserCh)
	}

	readKeysFromSlab(conn, p.Slab, p.BatchSize, chKeys)
	close(chKeys)

	stats2 := <-resReaderCh
	p.stats.Processed += stats2.Processed
	p.stats.GetErrors += stats2.GetErrors

	stats2 = <-resWriterCh
	p.stats.SetErrors += stats2.SetErrors

	stats2 = <-resEraserCh
	p.stats.DelErrors += stats2.DelErrors

	//fmt.Println("END RunOnce()")
	p.stats.EndTime = time.Now()
}

// Run will implement similar logic to ServerProcessor.Run(), i.e. will iterate through
// pages of keys and migrate all the data until the source slab is empty
func (p *SlabProcessor) Run() {

}

// GetStats returns stats collected doing a run
func (p *SlabProcessor) GetStats() Stats {
	return p.stats
}

// runWriter writes the items read by runReader
func (p *SlabProcessor) runWriter(memcacheAddresses []string, ch <-chan *memcache.Item, chCopiedKeys chan<- string, resCh chan<- Stats) {
	stats := Stats{}
	client := memcache.New(memcacheAddresses...)
	for item := range ch {
		stats.Processed++
		err := storeItem(client, item)
		if nil != err {
			log.Printf("Error storing '%s' / '%s': %s\n", item.Key, item.Value, err.Error())
			stats.SetErrors++
			continue
		}
		chCopiedKeys <- item.Key
	}
	resCh <- stats
	close(chCopiedKeys)
}

// runReader reads the keys read from a slab, and passes the retrieved items to runWriter
func (p *SlabProcessor) runReader(memcacheAddresses []string, chKeys <-chan string, chItems chan<- *memcache.Item, resCh chan<- Stats) {
	stats := Stats{}
	client := memcache.New(memcacheAddresses...)
	for key := range chKeys {
		stats.Processed++
		item, err := readItem(client, key)
		if nil != err {
			log.Printf("Error reading '%s': %s\n", key, err.Error())
			stats.GetErrors++
			continue
		}
		chItems <- item
	}
	resCh <- stats
	close(chItems)
}

// runEraser optionally deletes the migrated entries from the source node
func (p *SlabProcessor) runEraser(memcacheAddresses []string, chCopiedKeys <-chan string, resCh chan<- Stats) {
	stats := Stats{}
	client := memcache.New(memcacheAddresses...)
	for key := range chCopiedKeys {
		stats.Processed++
		err := deleteItem(client, key)
		if nil != err {
			log.Printf("Error deleting '%s': %s\n", key, err.Error())
			stats.DelErrors++
			continue
		}
	}
	resCh <- stats
}

// noOp empties the input channel of migrated entries. To be used in alternative to runEraser when the `Move` flag is false
func (p *SlabProcessor) noOp(memcacheAddresses []string, chCopiedKeys <-chan string, resCh chan<- Stats) {
	stats := Stats{}
	for _ = range chCopiedKeys {
		// do nothing, just empty the channel
		stats.Processed++
	}
	resCh <- stats
}
