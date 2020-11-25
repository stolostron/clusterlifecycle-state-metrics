package templateprocessor

import (
	"fmt"

	"github.com/ghodss/yaml"
)

type MapReader struct {
	assets map[string]string
}

var _ TemplateReader = &MapReader{assets: map[string]string{}}

func (r *MapReader) Asset(name string) ([]byte, error) {
	if s, ok := r.assets[name]; ok {
		return []byte(s), nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

func (r *MapReader) AssetNames() ([]string, error) {
	keys := make([]string, 0)
	for k := range r.assets {
		keys = append(keys, k)
	}
	return keys, nil
}

func (r *MapReader) ToJSON(b []byte) ([]byte, error) {
	return yaml.YAMLToJSON(b)
}

func NewTestReader(assets map[string]string) *MapReader {
	return &MapReader{assets}
}
