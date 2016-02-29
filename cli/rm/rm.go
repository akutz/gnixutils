package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	force bool
	dirs  bool
)

func init() {
	flag.BoolVar(&force, "f", false,
		"Force removal without prompting for confirmation")
	flag.BoolVar(&dirs, "d", false,
		"Attempt to remove directories")
}

func main() {
	flag.Parse()
	for _, p := range flag.Args() {
		rm(p)
	}
}

func rm(p string) {
	file, err := os.Stat(p)

	if err != nil && os.IsNotExist(err) {
		if force {
			os.Exit(0)
		} else {
			fmt.Printf("rm: %s: No such file or directory\n", p)
			os.Exit(1)
		}
	}

	if file.IsDir() && !dirs {
		if force {
			os.Exit(0)
		} else {
			fmt.Printf("rm: %s: is a directory\n", p)
			os.Exit(1)
		}
	}

	if dirs {
		err = os.RemoveAll(p)
	} else {
		err = os.Remove(p)
	}

	if !force && err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
