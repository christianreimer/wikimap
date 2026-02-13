package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/creimer/wikimap/internal/server"
)

func main() {
	tileDir := flag.String("tiles", "data/tiles", "directory containing generated tiles")
	webDir := flag.String("web", "web/dist", "directory containing frontend build")
	port := flag.Int("port", 8080, "port to listen on")
	flag.Parse()

	srv := server.NewServer(*tileDir, *webDir)
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Serving tiles from %s on http://localhost%s", *tileDir, addr)
	log.Fatal(http.ListenAndServe(addr, srv.Handler()))
}
