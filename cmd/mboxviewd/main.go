package main

import (
	"flag"
	"log"
	"os"

	"github.com/emurenMRz/mboxview/internal/server"
)

func main() {
	var mboxDir, staticDir, port, logFile string
	flag.StringVar(&mboxDir, "mbox-dir", "mail", "path to mbox files")
	flag.StringVar(&staticDir, "static-dir", "static", "path to static files")
	flag.StringVar(&port, "port", "8080", "port to listen on")
	flag.StringVar(&logFile, "log-file", "", "path to log file (default: stdout)")
	var edit bool
	flag.BoolVar(&edit, "edit", false, "enable edit mode")
	flag.Parse()

	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer file.Close()
		log.SetOutput(file)
	}

	server.RegisterHandlers(mboxDir, staticDir)
	server.SetEditMode(edit)

	log.Println("Listening on", port)
	if err := server.ListenAndServe(":" + port); err != nil {
		log.Printf("Server stopped with error: %v", err)
		return
	}
}
