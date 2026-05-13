package main

import (
	"log"
	"net"
	"net/http"

	"org-structure-api/internal/app"
	"org-structure-api/internal/config"
)

func main() {
	cfg := config.Load()
	a, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	addr := "0.0.0.0:" + cfg.Port
	log.Printf("starting api on %s", addr)
	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		log.Fatal(err)
	}
	if err := http.Serve(listener, a.Router()); err != nil {
		log.Fatal(err)
	}
}
