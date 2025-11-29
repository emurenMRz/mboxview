package main

import (
	"flag"
	"log"

	"github.com/emurenMRz/mboxview/internal/server"
)

func main() {
	var path string
	flag.StringVar(&path, "path", ".", "path to mbox files")
	var edit bool
	flag.BoolVar(&edit, "edit", false, "enable edit mode")
	flag.Parse()

	server.RegisterHandlers(path)
	server.SetEditMode(edit)

	log.Println("Listening on :8080...")
	if err := server.ListenAndServe(":8080"); err != nil {
		log.Fatal(err)
	}
}
