package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/gato/aof"
	"github.com/gato/aofgrep/filter"
	"io"
	"os"
	"regexp"
	"strings"
)

func processInput(input *bufio.Reader, ftr filter.Filter, invert bool) (matched, processed int, err error) {
	processed = 0
	matched = 0
	for {
		processed++
		op, e := aof.ReadOperation(input)
		if e != nil {
			if e == io.EOF {
				return
			}
			err = fmt.Errorf("Error processing command %d Error:%s\n", processed, e.Error())
			return
		}
		if filter.Match(op, ftr, invert) {
			e = op.ToAof(os.Stdout)
			matched++
			if e != nil {
				err = fmt.Errorf("Error writing command %d Error:%s\n", processed, e.Error())
				return
			}
		}
	}
}

func processFiles(opt Options) (matched, processed int, err error) {
	processed = 0
	matched = 0
	for _, file := range opt.Files {
		if opt.Debug {
			os.Stderr.WriteString(fmt.Sprintf("Parsing file %s\n", file))
		}
		f, e := os.Open(file)
		if e != nil {
			err = fmt.Errorf("Can't open file:%s Error:%s\n", file, e.Error())
			return
		}
		defer f.Close()
		var m, p int
		m, p, err = processInput(bufio.NewReader(f), opt.Filter, opt.Invert)
		if err != nil {
			return
		}
		matched += m
		processed += p
	}
	return
}

type Options struct {
	Filter filter.Filter
	Debug  bool
	Invert bool
	Files  []string
}

func parseCmdLine() (opt Options, err error) {

	filterCommand := flag.String("command", "", "a regexp for filtering by command")
	filterSubop := flag.String("subop", "", "a regexp for filtering by sub operation keys")
	filterKey := flag.String("key", "", "a regexp for filtering by key")
	filterParameter := flag.String("param", "", "a regexp for filtering by parameter")
	debug := flag.Bool("d", false, "output debug information (to STDERR)")
	invert := flag.Bool("v", false, "output command if does not match")

	flag.Parse()
	var e error
	if *filterCommand != "" {
		opt.Filter.Command, e = regexp.Compile(strings.ToUpper(*filterCommand))
		if e != nil {
			err = fmt.Errorf("Can't compile command regexp:%s Error:%s\n", *filterCommand, e.Error())
			return
		}
	}

	if *filterSubop != "" {
		opt.Filter.SubOp, e = regexp.Compile(*filterSubop)
		if e != nil {
			err = fmt.Errorf("Can't compile subop regexp:%s Error:%s\n", *filterSubop, e.Error())
			return
		}
	}
	if *filterKey != "" {
		opt.Filter.Key, e = regexp.Compile(*filterKey)
		if e != nil {
			err = fmt.Errorf("Can't compile key regexp:%s Error:%s\n", *filterKey, e.Error())
			return
		}
	}
	if *filterParameter != "" {
		opt.Filter.Parameter, e = regexp.Compile(*filterParameter)
		if e != nil {
			err = fmt.Errorf("Can't compile parameter regexp:%s Error:%s\n", *filterParameter, e.Error())
			return
		}
	}
	opt.Files = flag.Args()
	opt.Debug = *debug
	opt.Invert = *invert
	return
}

func main() {

	var matched, processed int

	options, err := parseCmdLine()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	if len(options.Files) > 0 {
		matched, processed, err = processFiles(options)
	} else {
		// process stdin
		matched, processed, err = processInput(bufio.NewReader(os.Stdin), options.Filter, options.Invert)
	}
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(2)
	}
	if options.Debug {
		os.Stderr.WriteString(fmt.Sprintf("%d matches found %d commands processed\n", matched, processed))
	}

}
