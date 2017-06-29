package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// MemcacheConfig holds the details of the memcached instance
type MemcacheConfig struct {
	Addresses []string `json:"addresses"`
	//MaxRetriesOnError int      `json:"max-retries"`
}

// Config holds the configuration of the load testing tool
type Config struct {
	MemcacheSrc  MemcacheConfig `json:"memcache_src"`
	MemcacheDest MemcacheConfig `json:"memcache_dest"`
	NThreads     int            `json:"n_threads"`
	Move         bool           `json:"move"`
}

func main() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	configuration := flags.String("conf", "", "path to configuration file")

	flags.Parse(os.Args[1:])

	if "" == *configuration {
		log.Println("Usage: ", os.Args[0], "-conf=<file.conf>")
		flag.PrintDefaults()
		os.Exit(2)
	}

	conf := readConfig(*configuration)

	timestart := time.Now()
	log.Println("START:", timestart.String())

	summary := StatsSummary{Title: "GLOBAL"}

	for _, srcAddr := range conf.MemcacheSrc.Addresses {
		// read keys directly from the source memcache instance
		proc := NewServerProcessor(srcAddr, conf.MemcacheDest.Addresses, conf.Move)
		proc.Run()
		st := proc.GetStatsSummary()
		st.Print()
		summary.Merge(st)
	}

	summary.Print()
}

// readConfig reads the configuration from file into an Config structure
func readConfig(filename string) Config {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("Cannot read configuration file", filename)
		os.Exit(1)
	}
	var conf Config
	err = json.Unmarshal(b, &conf)
	if err != nil {
		log.Println("Cannot parse configuration file:", err)
		os.Exit(1)
	}
	return conf
}
