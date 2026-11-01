package cmd

import (
	"github.com/SrGaabriel/stellation/cfg"
	"github.com/SrGaabriel/stellation/cli"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes your stellation modpack",
	Long: `Creates the configuration files necessary for stellation to run.

	This command should be run in the root directory of your modpack.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg.SaveConfig(cfg.DefaultConfig())
		cli.Success("stellation initialized successfully")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
