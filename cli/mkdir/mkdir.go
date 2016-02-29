package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	paths bool
)

func init() {
	flag.BoolVar(&paths, "p", false,
		"Create intermediate directories as required.")
}

func main() {
	flag.Parse()
	for _, p := range flag.Args() {
		mkdir(p)
	}
}

func mkdir(p string) {
	var err error
	if paths {
		err = os.MkdirAll(p, 0755)
	} else {
		err = os.Mkdir(p, 0755)
	}
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
