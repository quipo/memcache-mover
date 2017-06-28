package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

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

func NewSlabProcessor(srcMemcacheAddr string, destMemcacheAddr []string, slab int64, batchSize int64) *SlabProcessor {
	p := &SlabProcessor{
		Name:                  fmt.Sprintf("Slab Processor on host %s - slab %d (batch size: %d)", srcMemcacheAddr, slab, batchSize),
		Slab:                  slab,
		BatchSize:             batchSize,
		MemcacheSrcAddress:    srcMemcacheAddr,
		MemcacheDestAddresses: destMemcacheAddr,
	}

	p.stats = Stats{
		ProcessorName: p.Name,
	}

	return p
}

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

func (p *SlabProcessor) Run() {

}

func (p *SlabProcessor) GetStats() Stats {
	return p.stats
}

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

func (p *SlabProcessor) noOp(memcacheAddresses []string, chCopiedKeys <-chan string, resCh chan<- Stats) {
	stats := Stats{}
	for _ = range chCopiedKeys {
		// do nothing, just empty the channel
		stats.Processed++
	}
	resCh <- stats
}
