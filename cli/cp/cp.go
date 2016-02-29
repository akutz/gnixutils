package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
)

var (
	errSourceIsDir = errors.New("source file is directory")
	errNoForce     = errors.New("destination path exists; force is not enabled")

	rmTrailingSlashRx *regexp.Regexp

	force     bool
	recursive bool
	verbose   bool
)

func init() {
	rmTrailingSlashRx = regexp.MustCompile(`^(.*?)[/\\]?$`)

	flag.BoolVar(&force, "f", false,
		"If the destination exists, remove it and create a new file")
	flag.BoolVar(&recursive, "r", false,
		"Recursively copy directories")
	flag.BoolVar(&recursive, "R", false,
		"Recursively copy directories")
	flag.BoolVar(&verbose, "v", false,
		"Cause cp to be verbose, showing files as they are copied.")
}

func main() {
	flag.Parse()

	fargs := flag.Args()

	// there needs to be at least two arguments, a source and target path
	if len(fargs) < 2 {
		flag.Usage()
		os.Exit(64)
	}

	from := fargs[:len(fargs)-1]
	to := fargs[len(fargs)-1]

	var toIsDir bool
	toFile, err := os.Stat(to)

	if err == nil {
		toIsDir = toFile.IsDir()
	}

	// cannot copy more than one source file to a target path if the target
	// path is not a directory
	if !toIsDir && len(from) > 1 {
		flag.Usage()
		os.Exit(64)
	}

	// remove a trailing slash from the to path if it has one
	mTo := rmTrailingSlashRx.FindStringSubmatch(to)
	if len(mTo) == 2 {
		to = mTo[1]
	}

	errc := make(chan interface{})
	go func() {
		for _, p := range from {
			cp(p, to, errc)
		}
		close(errc)
	}()

	hasErrs := false
	for e := range errc {
		switch te := e.(type) {
		case string:
			fmt.Println(te)
			hasErrs = true
		case error:
			fmt.Println(te.Error())
			hasErrs = true
		}
	}

	if hasErrs {
		os.Exit(1)
	}
}

func cp(from, to string, errc chan interface{}) {

	fromFileInfo, err := os.Stat(from)
	if err != nil && os.IsNotExist(err) {
		errc <- fmt.Sprintf("cp: %s: No such file or directory", from)
		return
	}

	fromFileIsDir := fromFileInfo.IsDir()
	if !recursive && fromFileIsDir {
		errc <- fmt.Sprintf("cp: %s is a directory (not copied)", from)
		return
	}

	toFileInfo, err := os.Stat(to)

	// destination path does not exist
	if err != nil && os.IsNotExist(err) {
		if fromFileIsDir {

			if err := os.Mkdir(to, fromFileInfo.Mode()); err != nil {
				errc <- err
				return
			}

			fromDir, err := os.Open(from)
			if err != nil {
				errc <- err
				return
			}

			fromDirObjs, err := fromDir.Readdir(-1)
			if err != nil {
				errc <- err
				return
			}

			for _, o := range fromDirObjs {

				oName := o.Name()
				oFrom := path.Join(from, oName)
				oTo := path.Join(to, oName)

				cp(oFrom, oTo, errc)
			}

			// copy contents of from dir to the to dir
		} else {
			cpFileToFile(from, to, errc)
		}

		return
	}

	// destination path does exist. if the force flag is enabled nuke the
	// destination path
	if force {
		if err := os.RemoveAll(to); err != nil {
			errc <- err
			return
		}
		cp(from, to, errc)
		return
	}

	// if force is disabled and the destination path is a non-directory, then
	// return an error
	if !toFileInfo.IsDir() {
		errc <- fmt.Sprintf("cp: %s is a non-directory (not copied)", to)
		return
	}

	// at this point we know we're copying something into a directory
	cp(from, path.Join(to, path.Base(from)), errc)
}

func cpFileToFile(from, to string, errc chan interface{}) {
	fromFile, err := os.Open(from)
	if err != nil {
		errc <- err
		return
	}

	defer fromFile.Close()

	toFile, err := os.Create(to)
	if err != nil {
		errc <- err
		return
	}

	defer toFile.Close()

	if _, err := io.Copy(toFile, fromFile); err != nil {
		errc <- err
		return
	}

	if verbose {
		fmt.Printf("%[1]s -> %[2]s\n", from, to)
	}
}
