package cmd

import "github.com/urfave/cli/v3"

func RootCmd(version string) *cli.Command {
	return &cli.Command{
		Name:     "golinks",
		Usage:    "golinks",
		Version:  version,
		Commands: []*cli.Command{ServeCmd()},
	}
}
