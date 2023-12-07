package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

var (
	// flags
	configPath = flag.String("config", "config.json", "path to config file")
	config     Config

	logDir = flag.String("log", "out", "path to log directory")

	maxTimeout = flag.Int("maxTimeout", 1000, "max timeout for election")

	run_locally = flag.Bool("local", false, "Run entire system locally")

	multiple_req = flag.Bool("multiple_req", false, "Set to true if flag is present")
	multiple_cli = flag.Bool("multiple_cli", false, "Set to true if flag is present")
	leader_verbo = flag.Bool("leader_verbo", false, "Set to true if flag is present")
	call_watch   = flag.Bool("call_watch", false, "Set to true if flag is present")
)

const (
	version    string = "0.0.2"
	maxRetries int    = 5
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

	if *run_locally {
		// Run locally
		for idx := range config.Servers {
			// Initialise each server's file as empty file
			fileName := fmt.Sprintf("server%d.txt", idx)
			_, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				// Handle the error
				fmt.Println("Error opening file:", err)
				return
			}
			// Start zookeeper server with index idx
			go Run(idx)
		}
	} else {
		// Initialise each server's file as empty file
		id, dock_err := strconv.Atoi(os.Getenv("ID"))
		if dock_err != nil {
			log.Fatalf("failed to get ID from environment: %v", dock_err)
		}

		fmt.Printf("Starting server %d v%s\n", id, version)

		fileName := fmt.Sprintf("server%d.txt", id)
		_, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			// Handle the error
			fmt.Println("Error opening file:", err)
			return
		}
		// Start zookeeper server with index idx
		Run(id)
	}

	var input string
	fmt.Scanln(&input)
}
