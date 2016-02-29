package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	format := os.Args[1]
	formatPatt := regexp.MustCompile(`^"(.*)"$`)
	formatMatch := formatPatt.FindStringSubmatch(format)
	if len(formatMatch) > 0 {
		format = fmt.Sprintf(`\"%s\"`, formatMatch[1])
	}
	format, err := strconv.Unquote(fmt.Sprintf("\"%s\"", format))
	if err != nil {
		fmt.Fprintf(os.Stderr, "printf: error: %v\n", err)
		os.Exit(1)
	}
	if len(os.Args) == 2 {
		fmt.Printf(format)
	} else {
		fmt.Printf(format, os.Args[2])
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "printf: usage printf format [arguments]")
}
