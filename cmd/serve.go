package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves the output directory for presentation",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("serve called")
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().String("output-dir", "out", "Help message for toggle")
	serveCmd.Flags().Int("port", 8080, "What port to use for serving")
}

type ConfigServer struct {
	OutputDir string
	Port      int
}
