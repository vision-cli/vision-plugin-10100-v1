package generate

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/vision-cli/vision-plugin-10100-v1/cmd/model"
)

const TemplateDir = "_template"

//go:embed all:_template
var templateFiles embed.FS

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate a starter 10100 app",
	Long:  "generate a starter 10100 app with the settings from vision.json",
	Args:  cobra.NoArgs,
	RunE:  generateAndCheck,
}

type success struct {
	Success bool `json:"success"`
}

type convertConfig struct {
	PluginConfig model.PluginConfig `json:"config"`
	GoVersion    string
}

type SinglePluginConfig struct {
	Module string
}

type SingleConvertConfig struct {
	PluginConfig SinglePluginConfig
	GoVersion    string
}

// wraps the run function to determine a success or failed response
func generateAndCheck(cmd *cobra.Command, args []string) error {
	jEnc := json.NewEncoder(os.Stdout)

	err := run(cmd, args)
	if err != nil {
		_ = jEnc.Encode(success{Success: false})
		return fmt.Errorf("generating template: %w", err)
	}

	err = jEnc.Encode(success{Success: true})
	if err != nil {
		return fmt.Errorf("encoding JSON response: %w", err)
	}

	return nil
}

func run(cmd *cobra.Command, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	vPath := filepath.Join(wd, "./vision.json")
	vj, err := openVisionJson(vPath)
	if err != nil {
		return fmt.Errorf("opening vision.json: %w", err)
	}

	for _, module := range vj.PluginConfig.Modules {
		d := strings.Split(module, "/")
		outputPath := filepath.Join(model.PluginOutputDir, filepath.Join(d[2:]...))

		// if outputPath dir does not exist, create dir
		_, err = os.Stat(outputPath)
		if os.IsNotExist(err) {
			os.MkdirAll(outputPath, os.ModePerm)
		} else if err != nil {
			return fmt.Errorf("searching for output dir: %w", err)
		}

		var convConf SingleConvertConfig
		convConf.PluginConfig.Module = module
		convConf.GoVersion = vj.GoVersion

		err = walkDirAndClone(wd, outputPath, &convConf)
		if err != nil {
			return fmt.Errorf("walking dir and cloning: %w", err)
		}

		err = execGoModTidy(outputPath)
		if err != nil {
			return fmt.Errorf("executing go mod tidy: %w", err)
		}
	}

	return nil
}

func execGoModTidy(outputPath string) error {
	if outputPath == "." {
		outputPath = ""
	}
	c := exec.Command("go", "mod", "tidy")
	c.Dir = outputPath
	_, err := c.Output()
	if err != nil {
		return fmt.Errorf("running 'go mod tidy': %w", err)
	}
	return nil
}

func walkDirAndClone(wd, outputPath string, vj *SingleConvertConfig) error {
	return fs.WalkDir(templateFiles, TemplateDir, func(path string, d fs.DirEntry, err error) error {
		newPath := filepath.Join(wd, outputPath, strings.TrimPrefix(path, filepath.Join(TemplateDir, "/")))

		switch {
		case path == TemplateDir: // skip the top level template dir
			return nil
		case d.IsDir(): // if it is a dir then create it
			return cloneDir(newPath)
		case filepath.Ext(newPath) == ".tmpl":
			err := cloneExecTmpl(path, newPath, vj)
			if err != nil {
				return fmt.Errorf("cloning template files: %w", err)
			}

			return nil
		default:
			cloneFile(path, newPath)
			if err != nil {
				return fmt.Errorf("cloning files: %w", err)
			}
			return nil
		}
	})
}

func openVisionJson(vPath string) (*convertConfig, error) {
	f, err := os.OpenFile(vPath, os.O_RDWR, 0444)
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading bytes: %w", err)
	}

	var jsonData model.PluginData
	if err = json.Unmarshal(b, &jsonData); err != nil {
		return nil, fmt.Errorf("unmarshalling json: %w", err)
	}

	gv, err := getLatestGoVersion()
	if err != nil {
		return nil, fmt.Errorf("getting latest Go version: %w", err)
	}

	// convert struct to use correct JSON tag
	var convConf convertConfig
	convConf.PluginConfig = jsonData.PluginConfig
	convConf.GoVersion = gv

	return &convConf, nil
}

// if path is a directory, just copy it
func cloneDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// if file isn't template file, just copy it
func cloneFile(src, dst string) error {
	fsrc, err := templateFiles.Open(src)
	if err != nil {
		return fmt.Errorf("opening from templateFiles: %w", err)
	}
	defer fsrc.Close()
	fdst, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if os.IsExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("[clone] opening from clone: %w", err)
	}
	defer fdst.Close()
	_, err = io.Copy(fdst, fsrc)
	return err
}

func cloneExecTmpl(src, dst string, vj *SingleConvertConfig) error {
	// open file and read it
	trimmedNewPath := strings.TrimSuffix(dst, filepath.Ext(dst))
	err := cloneFile(src, trimmedNewPath)
	if err != nil {
		return fmt.Errorf("cloning file: %w", err)
	}
	f, err := os.OpenFile(trimmedNewPath, os.O_RDWR, 0444) // only enable reading mode as we do not need to write anything
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}
	err = f.Truncate(0)
	if err != nil {
		return fmt.Errorf("truncating: %w", err)
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("seeking: %w", err)
	}

	tmplEx, err := template.New("templateFile").Parse(string(b))
	if err != nil {
		return fmt.Errorf("creating template file: %w", err)
	}

	return tmplEx.Execute(f, vj)
}

func getLatestGoVersion() (string, error) {
	r, err := http.Get("https://go.dev/VERSION?m=text")

	if err != nil {
		return "", fmt.Errorf("getting Go version: %w", err)
	}

	body, err := io.ReadAll(r.Body)
	goVersion := string(body[2:8])

	return goVersion, nil
}
