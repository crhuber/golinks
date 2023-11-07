package main

import (
	"crhuber/golinks/cmd"

	log "github.com/sirupsen/logrus"
)

var version = "0.0.11"

func main() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	rootCmd := cmd.RootCmd(version)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
