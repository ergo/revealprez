package application

import (
	"archive/zip"
	_ "embed"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

//go:embed templates/index.go.tmpl
var indexPageTemplate string

// BuildFunc is main function governing building or watching sources
func BuildFunc(cmd *cobra.Command, args []string) {
	config := ConfigBuilder{}
	config.InputDir, _ = cmd.Flags().GetString("input-dir")
	config.OutputDir, _ = cmd.Flags().GetString("output-dir")
	config.Separator, _ = cmd.Flags().GetString("separator")
	config.EmbedSeparator, _ = cmd.Flags().GetString("embed-separator")
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

	err := getRevealJS(config)
	if err != nil {
		log.Fatal(err)
	}
	err = unpackRevealJS(config)
	if err != nil {
		log.Fatal(err)
	}

	// copy revealJS assets into destination
	err = copyAssets(config.RevealJSTemplateDir, config.OutputDir)
	if err != nil {
		log.Fatal(err)
	}
	if config.Watcher {
		watchInputDir(assetDir, assetOutputDir, config)
	} else {
		buildPresentation(assetDir, assetOutputDir, config)
	}
}

// buildPresentation copies the presentation assets to output directory
// and builds revealjs file out of slides file
func buildPresentation(assetDir string, assetOutputDir string, config ConfigBuilder) {
	// copy presentation assets
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

	fmt.Printf("Generating presentation...\n")
	slides := loadSlides(config)
	savePresentation(config, slides)
	fmt.Printf("Done\n")
}

func watchInputDir(assetDir string, assetOutputDir string, config ConfigBuilder) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	buildPresentation(assetDir, assetOutputDir, config)

	fmt.Printf("Started listening for changes in %v, press CTRL+C to stop\n", config.InputDir)

	done := make(chan bool)

	go func() {
		lastRun := time.Now()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				_ = event
				if time.Since(lastRun) > time.Second*2 {
					lastRun = time.Now()
					buildPresentation(assetDir, assetOutputDir, config)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(config.InputDir)
	if err != nil {
		log.Fatal(err)
	}
	<-done
	fmt.Println("DONE")
}

type ConfigBuilder struct {
	InputDir            string
	OutputDir           string
	Separator           string
	EmbedSeparator      string
	Filename            string
	AssetsDir           string
	Watcher             bool
	RevealJSTemplateDir string
	RevealJSVersion     string
}

type Slide struct {
	Markup string
	Index  int
}

func (s Slide) String() string {
	return fmt.Sprintf("<Slide %d>", s.Index)
}

func (s Slide) RenderedSlide() string {
	return fmt.Sprintf(`<section data-markdown><textarea data-template> %s </textarea></section>`, s.Markup)
}

type HTTPDownloadError struct {
	URL        string
	StatusCode int
}

func (e *HTTPDownloadError) Error() string {
	return fmt.Sprintf("Got %v for %v", e.StatusCode, e.URL)
}

// templates out the slides into RevealJS presentation
func savePresentation(config ConfigBuilder, slides []Slide) {
	var indexTemplate, err = template.New("index").Parse(indexPageTemplate)
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

// parses the slides from input file and embeds subslides
func loadSlides(config ConfigBuilder) []Slide {
	contentBytes, err := os.ReadFile(filepath.Join(config.InputDir, config.Filename))
	if err != nil {
		log.Fatal(err)
	}
	content := string(contentBytes)
	pages := strings.Split(content, config.Separator)
	var slides []Slide

	for index, page := range pages {
		re := regexp.MustCompile(config.EmbedSeparator)

		replaceEmbed := func(matched string) string {
			matchedFilename := strings.TrimSpace(re.FindStringSubmatch(matched)[1])
			filePath := filepath.Join(config.InputDir, strings.TrimSpace(matchedFilename))
			subcontentBytes, err := os.ReadFile(filePath)
			if err != nil {
				panic(err)
			}
			return string(subcontentBytes)
		}

		page := re.ReplaceAllStringFunc(page, replaceEmbed)
		slide := Slide{
			Markup: page,
			Index:  index + 1,
		}
		slides = append(slides, slide)
	}
	return slides
}

// generic copy function that mirrors directory structures between in/out
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

// grabs the zip file from github and places it in the directory along the binary
func getRevealJS(config ConfigBuilder) error {
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

// extracts the zip file into template dir for later use
func unpackRevealJS(config ConfigBuilder) error {
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

	fmt.Println("Unpacking RevealJS to revealjs_template")
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
