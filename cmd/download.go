package cmd

import (
	"github.com/lcmps/ExodiaLibrary/app"
	"github.com/spf13/cobra"
)

var dwCmd = &cobra.Command{
	Use: "download",
	RunE: func(cmd *cobra.Command, args []string) error {

		app.DownloadImages()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(dwCmd)
}
