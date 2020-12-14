package templateprocessor

import (
	"encoding/base64"
	"text/template"

	"github.com/ghodss/yaml"
	"k8s.io/klog"
)

func ApplierFuncMap() template.FuncMap {
	return template.FuncMap(GenericFuncMap())
}

// GenericFuncMap returns a copy of the basic function map as a map[string]interface{}.
func GenericFuncMap() map[string]interface{} {
	gfm := make(map[string]interface{}, len(genericMap))
	for k, v := range genericMap {
		gfm[k] = v
	}
	return gfm
}

var genericMap = map[string]interface{}{
	"toYaml":       toYaml,
	"encodeBase64": encodeBase64,
}

func toYaml(o interface{}) string {
	m, _ := yaml.Marshal(o)
	klog.V(5).Infof(string(m))
	return string(m)
}

func encodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
