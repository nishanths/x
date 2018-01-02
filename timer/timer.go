package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: timer <duration>\n")
	os.Exit(2)
}

func main() {
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	if runtime.GOOS != "darwin" {
		log.Fatal("timer notification supported on darwin only")
	}

	if flag.NArg() == 0 {
		usage()
	}

	input := flag.Arg(0)
	d, err := time.ParseDuration(input)
	if err != nil {
		log.Fatalf("failed to parse duration %q\nsee https://golang.org/pkg/time/#ParseDuration for accepted format", input)
	}

	start := time.Now()
	<-time.After(d)

	elapsed := fmt.Sprintf("%s elapsed", input)
	started := fmt.Sprintf("started %s", start.Format("Jan 2 15:04:05")) // time.Stamp but without the obnoxious space
	if err := notify("timer done", elapsed, started); err != nil {
		log.Print("failed to notify")
	}

	fmt.Println("done")
}

func notify(title, subtitle, message string) error {
	arg := fmt.Sprintf(`display notification "%s" with title "%s" subtitle "%s"`, message, title, subtitle)
	return exec.Command("osascript", "-e", arg).Run()
}
