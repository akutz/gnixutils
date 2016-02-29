package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"path"
)

var (
	location   bool
	silent     bool
	showError  bool
	remoteName bool
	verbose    bool
	output     string
)

func init() {
	flag.BoolVar(&location, "L", false,
		"(HTTP/HTTPS) If the server reports that the requested page "+
			"has moved to a different location (indicated with a Location: "+
			"header and a 3XX response code), this option will make curl "+
			"redo the request on the new place.")
	flag.BoolVar(&silent, "s", false,
		"Silent or quiet mode.")
	flag.BoolVar(&showError, "S", false,
		"When used with -s it makes curl show an error message if it fails.")
	flag.BoolVar(&remoteName, "O", false,
		"Write output to a local file named like the remote file we get.")
	flag.BoolVar(&verbose, "v", false,
		"Makes the fetching more verbose/talkative.")
	flag.StringVar(&output, "o", "",
		"Write output to <file> instead of stdout.")
}

func main() {
	flag.Parse()

	fargs := flag.Args()
	if len(fargs) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	curl(fargs[0])
}

func curl(rawURL string) {

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		fmt.Printf("curl: %s: error creating http request: %v\n", rawURL, err)
		os.Exit(1)
	}

	res, err := httpDo(req)

	if err != nil {
		fmt.Printf("curl: %s: error getting http response: %v\n", rawURL, err)
		os.Exit(1)
	}

	if output == "" && remoteName {
		output = path.Base(req.URL.Path)
	}

	var w io.Writer

	if output != "" {
		fw, err := os.OpenFile(output, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("curl: %s: error creating file: %v\n", output, err)
			os.Exit(1)
		}
		defer fw.Close()
		w = fw
	} else {
		w = os.Stdout
	}

	if verbose {
		d, err := httputil.DumpResponse(res, false)
		if err != nil {
			fmt.Printf("curl: %s: error dumping response: %v\n", rawURL, err)
			os.Exit(1)
		}
		if _, err := w.Write(d); err != nil {
			fmt.Printf("curl: %s: error printing response: %v\n", rawURL, err)
			os.Exit(1)
		}
	}

	defer res.Body.Close()
	wbuf := make([]byte, 1024)
	if _, err := io.CopyBuffer(w, res.Body, wbuf); err != nil {
		fmt.Printf("curl: %s: error writing file: %v\n", output, err)
		os.Exit(1)
	}
}

func httpDo(req *http.Request) (*http.Response, error) {
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}
	return client.Do(req)
}
