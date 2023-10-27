package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	// flags
	configPath = flag.String("config", "config.json", "path to config file")
	config     Config

	idx = flag.Int("idx", 0, "server index")

	maxTimeout = flag.Int("maxTimeout", 1000, "max timeout for election")
)

func parseConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("failed to parse config file: %v", err)
	}
}

func main() {
	flag.Parse()
	parseConfig(*configPath)
	// Run(*idx)
	for idx := range config.Servers {
		go Run(idx)
	}

	var input string
	fmt.Scanln(&input)
}
