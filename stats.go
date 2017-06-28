package main

import (
	"log"
	"time"
)

type Stats struct {
	ProcessorName string
	StartTime     time.Time
	EndTime       time.Time
	Processed     uint64
	GetErrors     uint64
	SetErrors     uint64
	DelErrors     uint64
}

type StatsSummary struct {
	Title     string
	Duration  time.Duration
	Processed uint64
	GetErrors uint64
	SetErrors uint64
	DelErrors uint64
}

func (s Stats) Print() {
	log.Printf("--[PARTIAL STATS %s]------------\n", s.ProcessorName)
	log.Printf("Elapsed time: \t%s\n", s.EndTime.Sub(s.StartTime).String())
	log.Printf("Processed: \t%d items\n", s.Processed)
	log.Printf("GET errors: \t%d\n", s.GetErrors)
	log.Printf("SET errors: \t%d\n", s.SetErrors)
	log.Printf("DELETE errors: \t%d\n", s.DelErrors)
}

func (sum *StatsSummary) Import(s Stats) {
	sum.Duration += s.EndTime.Sub(s.StartTime)
	sum.Processed += s.Processed
	sum.GetErrors += s.GetErrors
	sum.SetErrors += s.SetErrors
	sum.DelErrors += s.DelErrors
}

func (sum *StatsSummary) Merge(s StatsSummary) {
	sum.Duration += s.Duration
	sum.Processed += s.Processed
	sum.GetErrors += s.GetErrors
	sum.SetErrors += s.SetErrors
	sum.DelErrors += s.DelErrors
}

func (sum *StatsSummary) Print() {
	log.Println()
	log.Printf("--[SUMMARY %s]------------\n", sum.Title)
	log.Printf("Elapsed time: %s\n", sum.Duration.String())
	log.Printf("Processed: \t%d items\n", sum.Processed)
	log.Printf("GET errors: \t%d\n", sum.GetErrors)
	log.Printf("SET errors: \t%d\n", sum.SetErrors)
	log.Printf("DELETE errors: \t%d\n", sum.DelErrors)
}
