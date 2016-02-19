package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	host = flag.String("host", "0.0.0.0", "IP address to bind to")
	port = flag.Int("port", 8000, "TCP port to listen on")
)

var prog = filepath.Base(os.Args[0])

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s [-host=ADDR] [-port=NUM] [DIR]\n", prog)
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(prog + ": ")

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() > 1 {
		usage()
		os.Exit(1)
	}

	path := "."
	if flag.NArg() == 1 {
		path = flag.Arg(0)
	}

	log.Printf("Serving %q at http://%s:%d/", path, *host, *port)
	http.Handle("/", http.FileServer(http.Dir(path)))
	addr := fmt.Sprintf("%s:%d", *host, *port)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
