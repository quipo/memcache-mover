package main

import (
	"log"
	"net"
)

// ServerProcessor migrates data from a single memcached node,
// by iterating over the slabs
type ServerProcessor struct {
	Name                  string
	MemcacheSrcAddress    string
	MemcacheDestAddresses []string

	Move bool

	stats StatsSummary
}

// NewServerProcessor returns a new processor which can own the operations from a specific memcached node
func NewServerProcessor(srcMemcacheAddr string, destMemcacheAddr []string, move bool) *ServerProcessor {
	p := &ServerProcessor{
		Name:                  "Server Processor on host " + srcMemcacheAddr,
		MemcacheSrcAddress:    srcMemcacheAddr,
		MemcacheDestAddresses: destMemcacheAddr,
		Move: move,
	}

	p.stats = StatsSummary{Title: p.Name}

	return p
}

// RunOnce migrates a single batch of keys from the slabs on this server
func (p *ServerProcessor) RunOnce() {
	conn, err := net.Dial("tcp", p.MemcacheSrcAddress)
	if err != nil {
		log.Fatal(err)
	}

	// get list of slabs first
	slabs := getSlabList(conn)

	// get the keys from each slab (limited to 2MB max response)
	// @see https://github.com/memcached/memcached/issues/153
	// @see https://github.com/memcached/memcached/blob/master/items.c#L460

	for slab, batchSize := range slabs {
		proc := NewSlabProcessor(p.MemcacheSrcAddress, p.MemcacheDestAddresses, slab, batchSize, p.Move)
		proc.RunOnce()
		st := proc.GetStats()
		st.Print()
		p.stats.Import(st)
	}
}

// Run loops through pages of cached data, until all data has been migrated
// from the source server to the destination cluster
func (p *ServerProcessor) Run() {
	if !p.Move {
		// if we only copy data, there's no point looping, as we'd get the same
		// set of keys on each iteration
		p.RunOnce()
		return
	}

	lastRun := uint64(0)
	terminate := false
	for !terminate {
		p.RunOnce()
		// if the number of items migrated hasn't increased in the last iteration, stop looping
		if p.GetStatsSummary().Processed == lastRun {
			terminate = true
		}
		lastRun = p.GetStatsSummary().Processed
	}
}

// GetStatsSummary returns some stats about items copied/moved and errors
func (p *ServerProcessor) GetStatsSummary() StatsSummary {
	return p.stats
}
