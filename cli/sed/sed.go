package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"regexp/syntax"
	"strconv"
	"strings"
)

var (
	ere bool
	erE bool

	replIndicesLenRx = regexp.MustCompile(`\$(\d+)`)

	sedFlagG bool
)

func init() {
	flag.BoolVar(&erE, "E", false,
		"Interpret pattern as an extended regular expression.")
	flag.BoolVar(&ere, "e", false,
		"Interpret pattern as an extended regular expression.")
}

func main() {
	flag.Parse()

	fargs := flag.Args()
	if len(fargs) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	patt := fargs[0]
	delim := string(patt[1])
	parts := strings.Split(patt, delim)

	sedFrom := parts[1]
	sedTo := parts[2]
	sedFlags := parts[3]

	var pattFlags syntax.Flags

	if ere || erE {
		pattFlags |= syntax.Perl
	}
	if strings.Contains(sedFlags, "i") {
		pattFlags |= syntax.FoldCase
	}

	sedFlagG = strings.Contains(sedFlags, "g")

	rxParsed, err := syntax.Parse(sedFrom, pattFlags)
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

	var maxReplIndex int
	if m := replIndicesLenRx.FindAllStringSubmatch(sedTo, -1); len(m) > 0 {
		for _, sm := range m {
			i, err := strconv.Atoi(sm[1])
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			if i > maxReplIndex {
				maxReplIndex = i
			}
		}
	}

	if maxReplIndex > rx.NumSubexp() {
		fmt.Println("sed: replacement indices exceed capture groups")
		os.Exit(1)
	}

	switch len(fargs) {
	case 1:
		if sed(rx, sedTo, os.Stdin) {
			os.Exit(0)
		}
		os.Exit(1)
	case 2:
		fr, err := os.Open(fargs[1])
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		defer fr.Close()
		if sed(rx, sedTo, fr) {
			os.Exit(0)
		}
		os.Exit(1)
	}
}

func sed(rx *regexp.Regexp, repl string, r io.Reader) bool {
	matched := false
	s := bufio.NewScanner(r)
	for {
		if !s.Scan() {
			break
		}
		t := s.Text()

		if rx.MatchString(t) {
			matched = true
		} else {
			continue
		}

		rs := rx.ReplaceAllString(t, repl)
		fmt.Println(rs)
	}
	return matched
}
