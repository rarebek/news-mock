package main

import (
	"log"

	"github.com/k0kubun/pp"
	"tarkib.uz/config"
	"tarkib.uz/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}
	pp.Println("serving at http://localhost:8080")

	// Run
	app.Run(cfg)
}
