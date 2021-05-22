package cmd

import (
	"github.com/ergo/revealprez/application"
	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds your presentation",
	Long:  ``,
	Run:   application.BuildFunc,
}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	buildCmd.Flags().String("input-dir", "presentation", "Input directory")
	buildCmd.Flags().String("output-dir", "presentation_out", "Output directory")
	buildCmd.Flags().String("separator", "----SLIDE----", "Separator for slides in presentation")
	buildCmd.Flags().String("filename", "index.md", "Presentation filename")
	buildCmd.Flags().String("assets-dir", "assets", "Directory containing all the assets to include (inside the input dir)")
	buildCmd.Flags().Bool("watcher", false, "Should watch the directory for changes? (default: false)")
	buildCmd.Flags().String("revealjs-version", "4.1.0", "What version of reveal to grab")
}
