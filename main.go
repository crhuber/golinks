package main

import (
	"crhuber/golinks/cmd"
	"log/slog"
	"os"
)

var version = "0.0.13"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	rootCmd := cmd.RootCmd(version)

	if err := rootCmd.Execute(); err != nil {
		slog.Error("error", slog.Any("error", err))
	}
}
