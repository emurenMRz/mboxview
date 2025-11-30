package main

import (
	"flag"
	"log"

	"github.com/emurenMRz/mboxview/internal/server"
)

func main() {
	var mboxDir, staticDir, port string
	flag.StringVar(&mboxDir, "mbox-dir", "mail", "path to mbox files")
	flag.StringVar(&staticDir, "static-dir", "static", "path to static files")
	flag.StringVar(&port, "port", "8080", "port to listen on")
	var edit bool
	flag.BoolVar(&edit, "edit", false, "enable edit mode")
	flag.Parse()

	server.RegisterHandlers(mboxDir, staticDir)
	server.SetEditMode(edit)

	log.Println("Listening on", port)
	if err := server.ListenAndServe(":" + port); err != nil {
		log.Fatal(err)
	}
}
