package main

import (
	"fmt"
	"os"

	"github.com/Abhaythakor/dev-tools-installer/internal/config"
	"github.com/Abhaythakor/dev-tools-installer/internal/installer"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("installer.yaml")
	if err != nil {
		fmt.Printf("\033[31mError: %v\033[0m\n", err)
		os.Exit(1)
	}

	// Create and run installer
	inst := installer.New(cfg)
	if err := inst.Run(); err != nil {
		fmt.Printf("\033[31mError: %v\033[0m\n", err)
		os.Exit(1)
	}
}
