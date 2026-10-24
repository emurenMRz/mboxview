package main

import (
	"flag"
	"log"
	"os"

	"github.com/emurenMRz/mboxview/internal/server"
)

func main() {
	var mboxDir, staticDir, addr, port, logFile string
	flag.StringVar(&mboxDir, "mbox-dir", "mail", "path to mbox files")
	flag.StringVar(&staticDir, "static-dir", "static", "path to static files")
	flag.StringVar(&addr, "addr", "", "listen address (host:port). If empty, 127.0.0.1:<port> is used")
	flag.StringVar(&port, "port", "8080", "port to listen on (used when -addr is empty)")
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

	listenAddr := addr
	if listenAddr == "" {
		listenAddr = "127.0.0.1:" + port
	} else if port != "8080" {
		log.Printf("note: -addr %q takes precedence over -port %q", addr, port)
	}

	server.RegisterHandlers(mboxDir, staticDir)
	server.SetEditMode(edit)

	log.Println("Listening on", listenAddr)
	if err := server.ListenAndServe(listenAddr); err != nil {
		log.Printf("Server stopped with error: %v", err)
		return
	}
}
