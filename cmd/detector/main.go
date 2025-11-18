package main

import (
	"FairLAP/internal/app"
	"FairLAP/internal/config"
	"log"
)

const configPath = "config/config.yaml"

func main() {
	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		log.Fatal("read config file error:", err)
	}

	app.Run(cfg)
}
