package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

var (
	emptyBuf    = []byte{}
	doNotCreate bool
	now         time.Time
)

func init() {
	now = time.Now()
	flag.BoolVar(&doNotCreate, "c", false,
		"Do not create the file if it does not exist.")
}

func main() {
	flag.Parse()
	for _, p := range flag.Args() {
		touch(p)
	}
}

func touch(p string) {
	_, err := os.Stat(p)

	if err != nil && os.IsNotExist(err) {
		if doNotCreate {
			os.Exit(0)
		} else if err := ioutil.WriteFile(p, emptyBuf, 0644); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	} else {
		os.Chtimes(p, now, now)
	}
}
