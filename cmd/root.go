package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/josa42/git-sync/sync"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "git-sync",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		noColor, _ := cmd.Flags().GetBool("no-color")
		noPush, _ := cmd.Flags().GetBool("no-push")

		color.NoColor = noColor

		sync.Run(sync.RunOptions{
			Verbose: verbose,
			Push:    !noPush,
		})
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("no-color", "C", false, "")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "")
	rootCmd.Flags().BoolP("no-push", "P", false, "do not push to origin")
}
