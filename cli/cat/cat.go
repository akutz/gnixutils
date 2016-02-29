package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	args := os.Args

	if len(args) == 1 {
		if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	for _, p := range os.Args[1:] {

		r, err := os.Open(p)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if _, err := io.Copy(os.Stdout, r); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
