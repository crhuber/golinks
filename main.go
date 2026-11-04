package main

import (
	"context"
	"crhuber/golinks/cmd"
	"log/slog"
	"os"
)

var Version = "dev"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := cmd.RootCmd(Version).Run(context.Background(), os.Args); err != nil {
		slog.Error("error", slog.Any("error", err))
		os.Exit(1)
	}
}
