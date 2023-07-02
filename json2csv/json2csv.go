// Command json2csv converts JSON to CSV.
//
// For example, the following JSON input:
//
//	[
//	  {"key1": a, "key2": b},
//	  {"key1": x, "key2": y}
//	]
//
// is converted to the following CSV:
//
//	key1	key2
//	a	b
//	x	y
//
// The input must be a JSON list of objects, with each object having the
// same set (or subset) of keys. Values such as a, b, x, and y in the
// example are formatted using "%v". The command does not validate that
// the input meets the expected format.
//
// The default field delimiter in the CSV output is "\t".
package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"unicode/utf8"
)

var (
	delim  = flag.String("d", "\t", "CSV field delimiter rune")
	header = flag.Bool("h", true, "include CSV header")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: json2csv [flags] [file]\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("json2csv: ")

	flag.Usage = usage
	flag.Parse()

	var in io.ReadCloser

	switch flag.NArg() {
	case 0:
		in = ioutil.NopCloser(os.Stdin)
	case 1:
		var err error
		in, err = os.Open(flag.Arg(0))
		if err != nil {
			log.Fatalf("error opening file: %s", err)
		}
	default:
		usage()
		os.Exit(2)
	}

	defer in.Close() // ok to ignore error: file is read-only

	delim, _ := utf8.DecodeRuneInString(*delim)
	if delim == utf8.RuneError {
		log.Fatalf("invalid value for -d")
	}

	var l []map[string]interface{}
	if err := json.NewDecoder(in).Decode(&l); err != nil {
		log.Fatalf("error decoding json: %s", err)
	}

	if len(l) == 0 {
		return
	}

	var records [][]string

	if *header {
		var keys []string
		for k := range l[0] {
			keys = append(keys, k)
		}
		records = append(records, keys)
	}

	for _, m := range l {
		var row []string
		for _, v := range m {
			row = append(row, fmt.Sprintf("%v", v))
		}
		records = append(records, row)
	}

	w := csv.NewWriter(os.Stdout)
	w.Comma = delim

	if err := w.WriteAll(records); err != nil {
		log.Fatalf("error writing csv: %s", err)
	}

	if err := w.Error(); err != nil {
		log.Fatalf("%s", err)
	}
}
