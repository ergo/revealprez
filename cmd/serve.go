package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"path/filepath"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves the output directory for presentation",
	Long:  ``,
	Run:   servePresentation,
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

func servePresentation(cmd *cobra.Command, args []string) {
	config := ConfigServer{}
	config.Port, _ = cmd.Flags().GetInt("port")
	config.OutputDir, _ = cmd.Flags().GetString("output-dir")
	config.OutputDir, _ = filepath.Abs(config.OutputDir)
	fmt.Printf("Static directory: %s \nserving: http://127.0.0.1:%d \n", config.OutputDir, config.Port)

	fs := http.FileServer(http.Dir(config.OutputDir))
	http.Handle("/", fs)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}
