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
	flag.Usage = usage
	flag.Parse()

	if runtime.GOOS != "darwin" {
		log.Fatal("timer notification supported on darwin only")
	}

	if flag.NArg() == 0 {
		usage()
	}

	dur := flag.Arg(0)
	d, err := time.ParseDuration(dur)
	if err != nil {
		log.Fatalf("failed to parse duration %q\nsee https://golang.org/pkg/time/#ParseDuration for accepted format", dur)
	}

	<-time.After(d)
	elapsed := fmt.Sprintf("%s elapsed", dur)
	if err := notify("timer done", elapsed); err != nil {
		fmt.Fprintf(os.Stderr, "failed to notify\n")
	}
	fmt.Fprintf(os.Stdout, "done (%s)\n", elapsed)
}

func notify(title, message string) error {
	arg := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
	return exec.Command("osascript", "-e", arg).Run()
}
