package cmd

import (
	"os"

	"github.com/SrGaabriel/stellation/cfg"
	"github.com/SrGaabriel/stellation/cli"
	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := cfg.GetConfig()
		if err != nil {
			cli.Errorf("%v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
}
