package application

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"path/filepath"
)

func ServePresentationCmd(cmd *cobra.Command, args []string) {
	config := ConfigServer{}
	config.Port, _ = cmd.Flags().GetInt("port")
	config.OutputDir, _ = cmd.Flags().GetString("output-dir")
	config.OutputDir, _ = filepath.Abs(config.OutputDir)
	ServePresentation(config.OutputDir, config.Port)
}

func ServePresentation(OutputDir string, Port int) {
	fmt.Printf("Static directory: %s \nserving: http://127.0.0.1:%d \n", OutputDir, Port)

	fs := http.FileServer(http.Dir(OutputDir))
	http.Handle("/", fs)
	err := http.ListenAndServe(fmt.Sprintf(":%d", Port), nil)
	if err != nil {
		fmt.Println(err)
	}
}

type ConfigServer struct {
	OutputDir string
	Port      int
}
