package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"regexp/syntax"
)

var (
	invert          bool
	caseInsensitive bool
	ere             bool
	perl            bool
	fixed           bool
)

func init() {
	flag.BoolVar(&invert, "v", false,
		"Selected lines are those not matching any of the specified patterns.")
	flag.BoolVar(&caseInsensitive, "i", false,
		"Perform case insensitive matching.")
	flag.BoolVar(&ere, "E", false,
		"Interpret pattern as an extended regular expression.")
	flag.BoolVar(&perl, "P", false,
		"Interpret pattern as a Perl regular expression.")
	flag.BoolVar(&fixed, "F", false,
		"Interpret pattern as a list of fixed strings.")
}

func main() {
	flag.Parse()

	fargs := flag.Args()
	if len(fargs) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	patt := fargs[0]

	var pattFlags syntax.Flags

	if caseInsensitive {
		pattFlags |= syntax.FoldCase
	}

	if perl || ere {
		pattFlags |= syntax.Perl
	}

	if fixed {
		pattFlags |= syntax.Literal
	}

	rxParsed, err := syntax.Parse(patt, pattFlags)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	rxsz := rxParsed.String()

	var rx *regexp.Regexp
	if pattFlags&syntax.Perl > 0 {
		rx = regexp.MustCompile(rxsz)
	} else {
		rx = regexp.MustCompilePOSIX(rxsz)
	}

	if len(fargs) == 1 {
		if grep(rx, os.Stdin) {
			os.Exit(0)
		}
		os.Exit(1)
	}

	hasMatches := false
	for _, p := range fargs[1:] {
		func() {
			fr, err := os.Open(p)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			defer fr.Close()
			if hm := grep(rx, fr); hm {
				hasMatches = true
			}
		}()
	}
	if hasMatches {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func grep(rx *regexp.Regexp, r io.Reader) bool {
	matched := false
	s := bufio.NewScanner(r)
	for {
		if !s.Scan() {
			break
		}
		b := s.Bytes()

		m := rx.Match(b)
		if invert && !m {
			matched = true
			if _, err := os.Stdout.Write(b); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			if _, err := fmt.Fprintln(os.Stdout); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		} else if !invert && m {
			matched = true
			if _, err := os.Stdout.Write(b); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			if _, err := fmt.Fprintln(os.Stdout); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
	}
	return matched
}
