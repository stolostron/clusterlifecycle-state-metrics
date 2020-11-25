package templateprocessor

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
)

type YamlFileReader struct {
	rootDirectory string
}

var _ TemplateReader = &YamlFileReader{
	rootDirectory: "",
}

func (r *YamlFileReader) Asset(
	name string,
) ([]byte, error) {
	return ioutil.ReadFile(filepath.Clean(filepath.Join(r.rootDirectory, name)))
}

func (r *YamlFileReader) AssetNames() ([]string, error) {
	keys := make([]string, 0)
	_, err := os.Open(r.rootDirectory)
	if err != nil {
		return keys, err
	}
	err = filepath.Walk(r.rootDirectory, func(path string, info os.FileInfo, err error) error {
		if info != nil {
			if !info.IsDir() {
				newPath, err := filepath.Rel(r.rootDirectory, path)
				if err != nil {
					return err
				}
				keys = append(keys, newPath)
			}
		}
		return nil
	})
	return keys, err
}

func (*YamlFileReader) ToJSON(
	b []byte,
) ([]byte, error) {
	return yaml.YAMLToJSON(b)
}

func NewYamlFileReader(
	rootDirectory string,
) *YamlFileReader {
	return &YamlFileReader{
		rootDirectory: rootDirectory,
	}
}
