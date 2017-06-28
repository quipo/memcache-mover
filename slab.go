package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
)

var reSlabs *regexp.Regexp

func init() {
	reSlabs = regexp.MustCompile(`^STAT items:(\d*):number (\d*)`)
}

func getSlabList(conn net.Conn) map[int64]int64 {
	slabs := make(map[int64]int64)
	fmt.Fprintf(conn, "stats items\r\n")
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		if scanner.Text() == "END" {
			break
		}
		matches := reSlabs.FindStringSubmatch(scanner.Text())
		if len(matches) < 2 {
			continue
		}

		slab, err1 := strconv.ParseInt(matches[1], 10, 64)
		nEntries, err2 := strconv.ParseInt(matches[2], 10, 64)
		if nil != err1 || nil != err2 {
			log.Fatal(err1, err2)
		}
		slabs[slab] = nEntries
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return slabs
}

func readKeysFromSlab(conn net.Conn, slab int64, n int64, chKeys chan<- string) {
	//log.Printf("Getting %d keys from slab %d\n", n, slab)
	fmt.Fprintf(conn, "stats cachedump %d 0\r\n", slab)
	nItems := 0
	keyscanner := bufio.NewScanner(conn)
	for keyscanner.Scan() {
		line := keyscanner.Text()
		if strings.HasPrefix(line, "END") {
			break
		}
		if strings.HasPrefix(line, "ITEM ") {
			nItems++
			pos := strings.Index(line[5:], " ")
			if -1 != pos {
				chKeys <- line[5 : pos+5]
			}
			//fmt.Println(keyscanner.Text())
		}
		if (nItems % 10000) == 0 {
			log.Printf("[DEBUG] Processed %d items from slab %d\n", nItems, slab)
		}
	}
	if err := keyscanner.Err(); err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("READ %d items from slab %d\n", nItems, slab)
}
