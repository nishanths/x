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

var (
	notify = flag.Bool("notify", true, "show system notification when done")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: timer [-notify] [duration]\n")
}

func main() {
	log.SetPrefix("timer: ")
	log.SetFlags(0)

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}

	if *notify && runtime.GOOS != "darwin" {
		log.Fatal("system notification supported only on darwin")
	}

	d, err := time.ParseDuration(flag.Arg(0))
	if err != nil {
		log.Fatalf("error parsing duration: %s", err)
	}

	exitCode := 0

	start := time.Now()
	<-time.After(d)
	log.Println("done")

	if *notify {
		if err := systemNotification("timer done", d.String(), fmt.Sprintf("started %s", start.Format("Jan 2 15:04:05"))); err != nil {
			log.Printf("error showing system notification: %s", err)
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}

func systemNotification(title, subtitle, message string) error {
	arg := fmt.Sprintf(`display notification "%s" with title "%s" subtitle "%s"`, message, title, subtitle)
	return exec.Command("osascript", "-e", arg).Run()
}
