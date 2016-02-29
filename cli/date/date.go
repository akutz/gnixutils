package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"time"
)

var (
	utc  bool
	date int64

	parseAllRx = regexp.MustCompile(`(%[-_0^#]?[a-zA-Z]|[^%])`)
	parseRx    = regexp.MustCompile(`%[-_0^#]?([a-zA-Z])`)

	dtFmtFuncs = map[string]func(t time.Time, b []byte) []byte{
		"a": dtFmt("Mon"),
		"A": dtFmt("Monday"),
		"b": dtFmt("Jan"),
		"B": dtFmt("January"),
		"c": dtFmt("Mon Jan 2 15:04:05 2006"),
		"C": dtFmt("20"),
		"d": dtFmt("02"),
		"D": dtFmt("01/02/06"),
		"e": dtFmt("_2"),
		"F": dtFmt("2006-01-02"),
		"g": dtFmtg,
		"G": dtFmtG,
		"h": dtFmt("Jan"),
		"H": dtFmt("15"),
		"I": dtFmt("03"),
		"j": dtFmtj,
		"k": dtFmt("_15"),
		"l": dtFmt("_3"),
		"m": dtFmt("01"),
		"M": dtFmt("04"),
		"n": dtTxt("\n"),
		"N": dtFmt("000000000"),
		"p": dtFmt("PM"),
		"P": dtFmt("pm"),
		"r": dtFmt("15:04:05 PM"),
		"R": dtFmt("15:04"),
		"s": dtEpoch,
		"S": dtFmt("05"),
		"t": dtTxt("\t"),
		"T": dtFmt("15:04:05"),
		"u": dtFmtu,

		//  week number of year, with Sunday as first day of week (00..53)
		// "U": "",

		"V": dtFmtV,

		// day of week (0..6); 0 is Sunday
		"w": dtFmtw,

		// week number of year, with Monday as first day of week (00..53)
		// "W": "",

		"x": dtFmt("01/02/06"),
		"X": dtFmt("15:04:05"),
		"y": dtFmt("06"),
		"Y": dtFmt("2006"),
		"z": dtFmt("-0700"),
		"Z": dtFmt("MST"),
	}

	dtFmts = map[string]string{
		"a": "Mon",
		"A": "Monday",
		"b": "Jan",
		"B": "January",
		"c": "Mon Jan 2 15:04:05 2006",
		"C": "20",
		"d": "02",
		"D": "01/02/06",
		"e": "_2",
		"F": "2006-01-02",

		// last two digits of year of ISO week number (see %G)
		// "g": "",

		// year of ISO week number (see %V); normally useful only with %V
		// "G": "",

		"h": "Jan",
		"H": "15",
		"I": "03",

		//  day of year (001..366)
		// "j": "",

		"k": "_15",
		"l": "_3",
		"m": "01",
		"M": "04",
		"n": "\n",
		"N": "000000000",
		"p": "PM",
		"P": "pm",
		"r": "15:04:05 PM",
		"R": "15:04",

		// seconds since 1970-01-01 00:00:00 UTC
		// "s": "",

		"S": "05",
		"t": "\t",

		"T": "15:04:05",

		// day of week (1..7); 1 is Monday
		// "u": "",

		//  week number of year, with Sunday as first day of week (00..53)
		// "U": "",

		// ISO week number, with Monday as first day of week (01..53)
		// "V": "",

		// day of week (0..6); 0 is Sunday
		// "w": "",

		// week number of year, with Monday as first day of week (00..53)
		// "W": "",

		"x": "01/02/06",
		"X": "15:04:05",
		"y": "06",
		"Y": "2006",
		"z": "-0700",
		"Z": "MST",
	}
)

func init() {
	flag.BoolVar(&utc, "u", false,
		"Print the date in UTC (Coordinated Universal) time")
	flag.Int64Var(&date, "d", time.Now().Unix(),
		"Display time described, not 'now. Value should be epoch'")
}

func main() {
	flag.Parse()

	if utc {
		fmt.Printf("%d\n", date)
		os.Exit(0)
	}

	t := time.Unix(date, 0)
	var tfmt string

	fargs := flag.Args()
	if len(fargs) == 0 {
		tfmt = "+%a %b %d %H:%M:%S %Z %Y"
	} else {
		tfmt = fargs[0]
		if string(tfmt[0]) != "+" {
			fmt.Printf("date: %s: invalid format\n", tfmt)
			os.Exit(1)
		}
	}
	pfmt := tfmt[1:]

	buf := []byte{}

	fmts := parseAllRx.FindAllString(pfmt, -1)
	for _, f := range fmts {
		mf := parseRx.FindStringSubmatch(f)
		if len(mf) == 0 {
			buf = append(buf, []byte(f)...)
			continue
		}
		tok := mf[1]
		dtf, ok := dtFmtFuncs[tok]
		if !ok {
			fmt.Printf("date: %s: missing format function\n", tfmt)
			os.Exit(1)
		}
		buf = dtf(t, buf)
	}

	os.Stdout.Write(buf)
	fmt.Fprintln(os.Stdout)
}

func dtTxt(s string) func(t time.Time, b []byte) []byte {
	return func(t time.Time, b []byte) []byte {
		return append(b, []byte(s)...)
	}
}

func dtFmt(f string) func(t time.Time, b []byte) []byte {
	return func(t time.Time, b []byte) []byte {
		return t.AppendFormat(b, f)
	}
}

func dtFmtg(t time.Time, b []byte) []byte {
	y, _ := t.ISOWeek()
	szy := fmt.Sprintf("%d", y)
	g := []byte{szy[2], szy[3]}
	return append(b, g...)
}

func dtFmtG(t time.Time, b []byte) []byte {
	y, _ := t.ISOWeek()
	G := fmt.Sprintf("%d", y)
	return append(b, []byte(G)...)
}

func dtFmtj(t time.Time, b []byte) []byte {
	j := fmt.Sprintf("%03d", t.YearDay())
	return append(b, []byte(j)...)
}

func dtFmtu(t time.Time, b []byte) []byte {
	dow := int(t.Weekday())
	if dow == 0 {
		dow = 7
	} else {
		dow--
	}
	u := fmt.Sprintf("%d", dow)
	return append(b, []byte(u)...)
}

func dtFmtw(t time.Time, b []byte) []byte {
	dow := int(t.Weekday())
	w := fmt.Sprintf("%d", dow)
	return append(b, []byte(w)...)
}

func dtFmtV(t time.Time, b []byte) []byte {
	_, w := t.ISOWeek()
	V := fmt.Sprintf("%02d", w)
	return append(b, []byte(V)...)
}

func dtEpoch(t time.Time, b []byte) []byte {
	epoch := fmt.Sprintf("%d", t.Unix())
	return append(b, []byte(epoch)...)
}
