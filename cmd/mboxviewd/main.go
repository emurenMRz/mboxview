package main

import (
	"flag"
	"log"

	"github.com/emurenMRz/mboxview/internal/server"
)

func main() {
	var path string
	flag.StringVar(&path, "path", ".", "path to mbox files")
	flag.Parse()

	server.RegisterHandlers(path)

	log.Println("Listening on :8080...")
	if err := server.ListenAndServe(":8080"); err != nil {
		log.Fatal(err)
	}
}
