package cmd

import "github.com/spf13/cobra"

func RootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "golinks",
		Short:   "golinks",
		Long:    `golinks`,
		Version: version,
	}
	cmd.AddCommand(ServeCmd())
	return cmd
}
