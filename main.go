package main

import (
	"flag"

	"github.com/nzlov/tg/generate"
	"github.com/sirupsen/logrus"
)

var (
	trimprefix  = flag.String("trimprefix", "", "trim the `prefix` from the generated constant names")
	output      = flag.String("output", ".", "output path")
	linecomment = flag.Bool("linecomment", false, "use line comment text as printed text when present")
	verbose     = flag.Bool("verbose", false, "verbose")
	gonum       = flag.Int("gonum", 5, "go num")
	debug       = flag.Bool("debug", false, "debug log")
)

func main() {
	flag.Parse()
	logrus.SetLevel(logrus.InfoLevel)
	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// We accept either one directory or a list of files. Which do we have?
	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	}

	g := generate.NewGenerator(*gonum, *trimprefix, *output, *linecomment, *debug)
	g.ParsePackage(args, nil)

	g.Generate()

}
