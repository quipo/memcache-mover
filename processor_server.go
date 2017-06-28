package main

import (
	"log"
	"net"
)

type ServerProcessor struct {
	Name                  string
	MemcacheSrcAddress    string
	MemcacheDestAddresses []string

	Move bool

	stats StatsSummary
}

func NewServerProcessor(srcMemcacheAddr string, destMemcacheAddr []string) *ServerProcessor {
	p := &ServerProcessor{
		Name:                  "Server Processor on host " + srcMemcacheAddr,
		MemcacheSrcAddress:    srcMemcacheAddr,
		MemcacheDestAddresses: destMemcacheAddr,
	}

	p.stats = StatsSummary{Title: p.Name}

	return p
}

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
		proc := NewSlabProcessor(p.MemcacheSrcAddress, p.MemcacheDestAddresses, slab, batchSize)
		proc.Move = p.Move
		proc.RunOnce()
		st := proc.GetStats()
		st.Print()
		p.stats.Import(st)
	}
}

func (p *ServerProcessor) Run() {
	if !p.Move {
		p.RunOnce()
		return
	}

	lastRun := uint64(0)
	terminate := false
	for !terminate {
		p.RunOnce()
		if p.GetStatsSummary().Processed == lastRun {
			terminate = true
		}
		lastRun = p.GetStatsSummary().Processed
	}
}

func (p *ServerProcessor) GetStatsSummary() StatsSummary {
	return p.stats
}
