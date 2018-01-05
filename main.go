package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

var (
	host = flag.String("host", "", "IP address to bind to")
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

	addr := net.JoinHostPort(*host, strconv.Itoa(*port))
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("cannot listen: %v", err)
	}
	log.Printf("Serving %q at http://%s/", path, l.Addr())
	srv := &http.Server{
		Handler: http.FileServer(http.Dir(path)),
	}
	if err := srv.Serve(l); err != nil {
		log.Fatal(err)
	}
}
