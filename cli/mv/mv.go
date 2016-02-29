package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	doNotOverwrite bool
)

func init() {
	flag.BoolVar(&doNotOverwrite, "n", false,
		"Do not overwrite an existing file.")
}

func main() {
	flag.Parse()

	from := flag.Arg(0)
	to := flag.Arg(1)

	_, err := os.Stat(to)

	if err == nil && doNotOverwrite {
		os.Exit(0)
	}

	if err := os.Rename(from, to); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
