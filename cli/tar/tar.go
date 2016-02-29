package main

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"path"

	"github.com/akutz/gnixutils/lib/os/group"
)

var (
	list      bool
	create    bool
	doGzip    bool
	doBzip    bool
	verbose   bool
	extract   bool
	file      string
	changeDir string

	addToTarBuf = make([]byte, 1024)
)

func init() {
	flag.BoolVar(&list, "t", false,
		"List archive contents to stdout.")
	flag.BoolVar(&extract, "x", false,
		"Extract to disk from the archive.")
	flag.BoolVar(&create, "c", false,
		"Create a new archive containing the specified items.")
	flag.StringVar(&changeDir, "C", "",
		"In c mode, this changes the directory before adding files.")
	flag.StringVar(&file, "f", "",
		"Read the archive from or write the archive to the specified file.")
	flag.BoolVar(&doGzip, "z", false,
		"Compress the resulting archive with gzip.")
	flag.BoolVar(&doBzip, "j", false,
		"Compress the resulting archive with bzip2.")
	flag.BoolVar(&verbose, "v", false,
		"Produce verbose output.")
}

func main() {
	flag.Parse()

	if file == "" {
		flag.Usage()
		os.Exit(1)
	}

	if create {
		createTar()
	} else if extract || list {

		_, err := os.Stat(file)
		if err != nil && os.IsNotExist(err) {
			fmt.Printf("tar: %s: file does not exist\n", file)
			os.Exit(1)
		}

		var r io.Reader

		fr, err := os.Open(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer fr.Close()

		if doGzip {
			r = gzipDecompress(fr)
		} else if doBzip {
			r = bzip2Decompress(fr)
		} else {
			r = fr
		}

		if extract {
			extractTar(r)
		} else if list {
			listTar(r)
		}
	}
}

func createTar() {
	paths := flag.Args()
	if len(paths) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if _, err := os.Stat(path.Dir(file)); os.IsNotExist(err) {
		fmt.Printf("tar: %s: parent path does not exist\n", file)
		os.Exit(1)
	}

	fw, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("tar: %s: error creating file: %v\n", file, err)
		os.Exit(1)
	}
	defer fw.Close()

	cwdOrig, err := os.Getwd()
	if err != nil {
		fmt.Printf("tar: %s: error getting current dir: %v\n", file, err)
		os.Exit(1)
	}

	if changeDir != "" {
		os.Chdir(changeDir)
	}
	defer os.Chdir(cwdOrig)

	var w io.Writer

	if doGzip {
		gw := gzip.NewWriter(fw)
		defer gw.Close()
		w = gw
	} else {
		w = fw
	}

	tw := tar.NewWriter(w)
	defer tw.Close()

	for _, p := range paths {
		addToTar(p, tw)
	}
}

func addToTar(p string, w *tar.Writer) {
	fi, err := os.Stat(p)
	if err != nil {
		fmt.Printf("tar: %s: error stat'ing file: %v\n", p, err)
		os.Exit(1)
	}

	h, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		fmt.Printf("tar: %s: error creating header: %v\n", p, err)
		os.Exit(1)
	}
	h.Name = p

	user, err := user.LookupId(fmt.Sprintf("%d", h.Uid))
	if err != nil {
		fmt.Printf("tar: %s: error getting file's owner: %v\n", p, err)
		os.Exit(1)
	}
	h.Uname = user.Username

	grp, err := group.LookupGroupID(fmt.Sprintf("%d", h.Gid))
	if err != nil {
		fmt.Printf("tar: %s: error getting file's group: %v\n", p, err)
		os.Exit(1)
	}
	h.Gname = grp.Name

	w.WriteHeader(h)

	switch h.Typeflag {
	case tar.TypeDir:
		if verbose {
			fmt.Printf("a %s\n", p)
		}

		dir, err := os.Open(p)
		if err != nil {
			fmt.Printf("tar: %s: error opening dir: %v\n", p, err)
			os.Exit(1)
		}

		objs, err := dir.Readdir(-1)
		if err != nil {
			fmt.Printf("tar: %s: error listing dir contents: %v\n", p, err)
			os.Exit(1)
		}
		for _, o := range objs {
			addToTar(path.Join(p, o.Name()), w)
		}
	case tar.TypeReg, tar.TypeRegA:
		r, err := os.Open(p)
		if err != nil {
			fmt.Printf("tar: %s: error opening file: %v\n", p, err)
			os.Exit(1)
		}
		defer r.Close()
		if _, err := io.CopyBuffer(w, r, addToTarBuf); err != nil {
			fmt.Printf("tar: %s: error writing to tar: %v\n", p, err)
			os.Exit(1)
		}
		if verbose {
			fmt.Printf("a %s\n", p)
		}
	}
}

func extractTar(r io.Reader) {
	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(h.Name, os.FileMode(h.Mode)); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("x %s\n", h.Name)
		case tar.TypeReg, tar.TypeRegA:
			func() {
				f, err := os.OpenFile(
					h.Name,
					os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
					os.FileMode(h.Mode))

				if err != nil {
					fmt.Printf(
						"tar: %s: error opening file: %v\n",
						h.Name, err)
					os.Exit(1)
				}
				defer f.Close()

				buf := make([]byte, 1024)
				for {
					nr, err := tr.Read(buf)
					if err == io.EOF {
						if verbose {
							fmt.Printf("x %s\n", h.Name)
						}
						break
					}
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					nw, err := f.Write(buf[:nr])
					if err != nil {
						fmt.Printf(
							"tar: %s: error writing file: %v\n",
							h.Name, err)
						os.Exit(1)
					}
					if nw != nr {
						fmt.Printf(
							"tar: %s: %d bytes read; %d bytes written\n",
							h.Name, nr, nw)
					}
				}
			}()
		}
	}
}

func listTar(r io.Reader) {
	tr := tar.NewReader(r)
	hdrs := []*tar.Header{}

	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if verbose {
			hdrs = append(hdrs, h)
		} else {
			fmt.Println(h.Name)
		}
	}

	if !verbose {
		return
	}

	var maxLenUname, maxLenGname, minLenSize int
	minLenSize = 1

	for _, h := range hdrs {
		if lenUname := len(h.Uname); lenUname > maxLenUname {
			maxLenUname = lenUname
		}
		if lenGname := len(h.Gname); lenGname > maxLenGname {
			maxLenGname = lenGname
		}
		if lenSize := len(fmt.Sprintf("%d", h.Size)); lenSize < minLenSize {
			minLenSize = lenSize
		}
	}

	lenUname := maxLenUname + 1
	lenGname := maxLenGname + 1
	lenSize := minLenSize + 5

	fmtStr := fmt.Sprintf(
		"%%s  %%d %%-%ds %%-%ds %%%dd %%s %%s\n",
		lenUname,
		lenGname,
		lenSize)

	//fmt.Print(fmtStr)
	//fmt.Printf("lenUname=%d\n", lenUname)
	//fmt.Printf("lenGname=%d\n", lenGname)
	//fmt.Printf("lenSize=%d\n", lenSize)

	for _, h := range hdrs {
		fmt.Printf(
			fmtStr,
			os.FileMode(h.Mode), 0,
			h.Uname, h.Gname,
			h.Size,
			h.ModTime.Format("Jan 02 15:04"),
			h.Name)
	}
}

func gzipDecompress(r io.Reader) io.Reader {
	rdr, err := gzip.NewReader(r)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rdr.Close()

	buf := make([]byte, 1024)
	pr, pw := io.Pipe()

	go func() {
		for {
			_, err := rdr.Read(buf)
			if err == io.EOF {
				break
			}
			pw.Write(buf)
		}
		pw.Close()
	}()

	return pr
}

func bzip2Decompress(r io.Reader) io.Reader {
	rdr := bzip2.NewReader(r)

	buf := make([]byte, 1024)
	pr, pw := io.Pipe()

	go func() {
		for {
			_, err := rdr.Read(buf)
			if err == io.EOF {
				break
			}
			pw.Write(buf)
		}
		pw.Close()
	}()

	return pr
}
