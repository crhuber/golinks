package main

import (
	"crhuber/golinks/cmd"
	"log/slog"
	"os"
)

var Version = "dev"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	rootCmd := cmd.RootCmd(Version)

	if err := rootCmd.Execute(); err != nil {
		slog.Error("error", slog.Any("error", err))
	}
}
