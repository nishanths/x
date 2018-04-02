// Command json2csv coverts JSON of the format:
//
//   [
//     {"key": a, "key2": b},
//     {"key": x, "key2": y},
//   ]
//
// to TSV of the format:
//
//   key key2
//   a b
//   x y
//
// The input must be a JSON list of objects, with each object
// having the same set (or subset) of keys. The values a, b, x,
// and y are formatted using '%v'. The program does not validate
// whether the input meets the expected format.
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

var delim = flag.String("d", "\t", "field delimiter rune")

func usage() {
	fmt.Fprint(os.Stderr, "usage: json2csv [flags] [file]\n\n")
	fmt.Fprintf(os.Stderr, "flags\n%s\n", `   -d <delim>  field delimiter rune (default "\t")`)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("json2csv")

	var in io.ReadCloser

	switch flag.NArg() {
	case 0:
		in = ioutil.NopCloser(os.Stdin)
	case 1:
		var err error
		in, err = os.Open(flag.Arg(0))
		if err != nil {
			log.Fatalf("failed to open file: %s", err)
		}
	default:
		usage()
	}

	defer in.Close()

	var l []map[string]interface{}
	if err := json.NewDecoder(in).Decode(&l); err != nil {
		log.Fatalf("failed to read json: %s", err)
	}

	if len(l) == 0 {
		return
	}

	var records [][]string

	var keys []string
	for k := range l[0] {
		keys = append(keys, k)
	}
	records = append(records, keys)

	for _, m := range l {
		var row []string
		for _, v := range m {
			row = append(row, fmt.Sprintf("%v", v))
		}
		records = append(records, row)
	}

	r, _ := utf8.DecodeRuneInString(*delim)
	w := csv.NewWriter(os.Stdout)
	w.Comma = r
	w.WriteAll(records)

	if err := w.Error(); err != nil {
		log.Printf("failed to write CSV: %s", err)
	}
}
