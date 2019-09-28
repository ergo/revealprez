package cmd

import (
	"bytes"
	"fmt"
	"github.com/ergo/revealprez/templates"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds your presentation",
	Long:  ``,
	Run:   runFunc,
}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	buildCmd.Flags().String("input-dir", "", "Input directory")
	buildCmd.Flags().String("output-dir", "out", "Output directory")
	buildCmd.Flags().String("separator", "----SLIDE----", "Help message for toggle")
	buildCmd.Flags().String("filename", "index.md", "Presentation filename")
	buildCmd.Flags().String("assets-dir", "assets", "Directory containing all the assets to include")
	buildCmd.Flags().Bool("watcher", false, "Should watch the directory for changes? (default: false)")
}

type ConfigBuilder struct {
	InputDir  string
	OutputDir string
	Separator string
	Filename  string
	AssetsDir string
	Watcher   bool
}

type Slide struct {
	Markup []byte
	Index  int
}

func (s Slide) String() string {
	return fmt.Sprintf("<Slide %d>", s.Index)
}

func (s Slide) RenderedSlide() string {
	return fmt.Sprintf(`<section data-markdown>
	<textarea data-template>
		%s
	</textarea>
</section>`, s.Markup)
}

func savePresentation(config ConfigBuilder, slides []Slide) {
	var indexTemplate, err = template.New("index").Parse(templates.IndexTemplate)
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(config.OutputDir, 0777)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
	file, err := os.Create(path.Join(config.OutputDir, "index.html"))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = indexTemplate.Execute(file, slides)
	return
}

func loadSlides(config ConfigBuilder) []Slide {
	content, err := ioutil.ReadFile(filepath.Join(config.InputDir, config.Filename))
	if err != nil {
		log.Fatal(err)
	}
	pages := bytes.Split(content, []byte("----SLIDE----"))
	var slides []Slide

	for index, page := range pages {
		slide := Slide{
			Markup: page,
			Index:  index + 1,
		}
		slides = append(slides, slide)
	}
	return slides
}

func copyAssets(sourceDir string, destinationDir string) {
	err := filepath.Walk(sourceDir,
		func(currPath string, info os.FileInfo, err error) error {
			if err != nil {
				log.Fatal(err)
			}
			basePath, err := filepath.Rel(sourceDir, currPath)
			destinationPath := filepath.Join(destinationDir, basePath)
			if info.IsDir() {
				err := os.MkdirAll(destinationPath, info.Mode())
				if err != nil && !os.IsExist(err) {
					return err
				}
			} else if info.Mode().IsRegular() {
				content, err := ioutil.ReadFile(currPath)
				if err != nil {
					return err
				}
				err = ioutil.WriteFile(destinationPath, content, info.Mode())
				if err != nil {
					return err
				}
			}
			return err
		})
	if err != nil {
		log.Fatal(err)
	}
}

func runFunc(cmd *cobra.Command, args []string) {
	config := ConfigBuilder{}
	config.InputDir, _ = cmd.Flags().GetString("input-dir")
	config.OutputDir, _ = cmd.Flags().GetString("output-dir")
	config.Separator, _ = cmd.Flags().GetString("separator")
	config.Filename, _ = cmd.Flags().GetString("filename")
	config.AssetsDir, _ = cmd.Flags().GetString("assets-dir")
	config.Watcher, _ = cmd.Flags().GetBool("watcher")
	config.InputDir, _ = filepath.Abs(config.InputDir)
	config.OutputDir, _ = filepath.Abs(config.OutputDir)
	fmt.Printf("Input directory: %s\n", config.InputDir)
	fmt.Printf("Output directory: %s\n", config.OutputDir)
	assetDir := path.Join(config.InputDir, config.AssetsDir)
	assetOutputDir := path.Join(config.OutputDir, config.AssetsDir)
	_, err := os.Stat(assetDir)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	} else if !os.IsNotExist(err) {
		copyAssets(assetDir, assetOutputDir)
	} else {
		fmt.Printf("Assets dir not found: %s \n", assetDir)
	}
	copyAssets(path.Join(".", "revealjs_template"), config.OutputDir)
	slides := loadSlides(config)
	fmt.Printf("Generating presentation...")
	savePresentation(config, slides)
}
