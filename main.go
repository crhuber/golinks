package main

import (
	"crhuber/golinks/cmd"

	log "github.com/sirupsen/logrus"
)

var version = "0.0.7"

func main() {
	rootCmd := cmd.RootCmd(version)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
