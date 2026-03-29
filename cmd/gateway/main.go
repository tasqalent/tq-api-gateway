package main

import (
	"log"
	"net/http"

	"github.com/tasqalent/tq-api-gateway/internal/config"
	"github.com/tasqalent/tq-api-gateway/internal/server"
)

func main() {
	cfg := config.Load()
	h := server.New(cfg)

	log.Printf("listening on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, h); err != nil {
		log.Fatal(err)
	}
	
}