package cmd

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/ergo/revealprez/templates"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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
	buildCmd.Flags().String("separator", "----SLIDE----", "Separator for slides in presentation")
	buildCmd.Flags().String("filename", "index.md", "Presentation filename")
	buildCmd.Flags().String("assets-dir", "assets", "Directory containing all the assets to include")
	buildCmd.Flags().Bool("watcher", false, "Should watch the directory for changes? (default: false)")
	buildCmd.Flags().String("revealjs-version", "3.8.0", "What version of reveal to grab")
}

type ConfigBuilder struct {
	InputDir            string
	OutputDir           string
	Separator           string
	Filename            string
	AssetsDir           string
	Watcher             bool
	RevealJSTemplateDir string
	RevealJSVersion     string
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

type HTTPDownloadError struct {
	URL        string
	StatusCode int
}

func (e *HTTPDownloadError) Error() string {
	return fmt.Sprintf("Got %v for %v", e.StatusCode, e.URL)
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

func copyAssets(sourceDir string, destinationDir string) error {
	err := filepath.Walk(sourceDir,
		func(currPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			basePath, err := filepath.Rel(sourceDir, currPath)
			destinationPath := filepath.Join(destinationDir, basePath)
			if info.IsDir() {
				err := os.MkdirAll(destinationPath, info.Mode())
				if err != nil && !os.IsExist(err) {
					return err
				}
			} else if info.Mode().IsRegular() {
				source, err := os.Open(currPath)
				if err != nil {
					return err
				}
				defer source.Close()
				destination, err := os.Create(destinationPath)
				if err != nil {
					return err
				}
				defer destination.Close()
				_, err = io.Copy(destination, source)
			}
			return err
		})
	return err
}

var getRevealJS = func(config ConfigBuilder) error {
	downloadURL := fmt.Sprintf("https://github.com/hakimel/reveal.js/archive/%s.zip", config.RevealJSVersion)
	destinationFile, _ := filepath.Abs(fmt.Sprintf("revealjs.%s.zip", config.RevealJSVersion))

	_, err := os.Stat(destinationFile)
	if err == nil {
		return nil
	}

	fmt.Printf("Downloading %s...\n", downloadURL)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	if resp.StatusCode > 300 {
		return &HTTPDownloadError{downloadURL, resp.StatusCode}
	}

	defer resp.Body.Close()

	destination, err := os.Create(destinationFile)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, resp.Body)
	return err

}

var unpackRevealJS = func(config ConfigBuilder) error {
	files, err := ioutil.ReadDir(config.RevealJSTemplateDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if len(files) > 0 {
		return nil
	}

	err = os.MkdirAll(config.RevealJSTemplateDir, 0777)
	if err != nil && !os.IsExist(err) {
		return err
	}

	fmt.Println("Unpacking to revealjs_template")
	destinationFile, _ := filepath.Abs(fmt.Sprintf("revealjs.%s.zip", config.RevealJSVersion))
	r, err := zip.OpenReader(destinationFile)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		err := func() error {
			newName, err := filepath.Rel(fmt.Sprintf("reveal.js-%s", config.RevealJSVersion), f.Name)
			destinationPath := filepath.Join(config.RevealJSTemplateDir, newName)
			rc, err := f.Open()
			info := f.FileInfo()
			defer rc.Close()
			if err != nil {
				return err
			}
			if info.IsDir() {
				err := os.MkdirAll(destinationPath, info.Mode())
				if err != nil && !os.IsExist(err) {
					return err
				}
			} else if info.Mode().IsRegular() {
				destination, err := os.Create(destinationPath)
				if err != nil {
					return err
				}
				defer destination.Close()
				_, err = io.Copy(destination, rc)
			}
			return err
		}()
		if err != nil {
			return err
		}
	}
	return err
}

func runFunc(cmd *cobra.Command, args []string) {
	config := ConfigBuilder{}
	config.InputDir, _ = cmd.Flags().GetString("input-dir")
	config.OutputDir, _ = cmd.Flags().GetString("output-dir")
	config.Separator, _ = cmd.Flags().GetString("separator")
	config.Filename, _ = cmd.Flags().GetString("filename")
	config.AssetsDir, _ = cmd.Flags().GetString("assets-dir")
	config.Watcher, _ = cmd.Flags().GetBool("watcher")
	config.RevealJSVersion, _ = cmd.Flags().GetString("revealjs-version")
	config.InputDir, _ = filepath.Abs(config.InputDir)
	config.OutputDir, _ = filepath.Abs(config.OutputDir)
	config.RevealJSTemplateDir, _ = filepath.Abs("revealjs_template")
	fmt.Printf("Input directory: %s\n", config.InputDir)
	fmt.Printf("Output directory: %s\n", config.OutputDir)
	assetDir := path.Join(config.InputDir, config.AssetsDir)
	assetOutputDir := path.Join(config.OutputDir, config.AssetsDir)

	_, err := os.Stat(assetDir)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	} else if !os.IsNotExist(err) {
		err = copyAssets(assetDir, assetOutputDir)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("Assets dir not found: %s \n", assetDir)
	}

	err = getRevealJS(config)
	if err != nil {
		log.Fatal(err)
	}
	err = unpackRevealJS(config)
	if err != nil {
		log.Fatal(err)
	}
	err = copyAssets(config.RevealJSTemplateDir, config.OutputDir)
	if err != nil {
		log.Fatal(err)
	}
	slides := loadSlides(config)
	fmt.Printf("Generating presentation...")
	savePresentation(config, slides)
}
