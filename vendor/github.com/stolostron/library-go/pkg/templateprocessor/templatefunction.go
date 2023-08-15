// Copyright Contributors to the Open Cluster Management project

// Deprecated: Use github.com/stolostron/applier/pkg/templateprocessor project
package templateprocessor

import (
	"bytes"
	"encoding/base64"
	"text/template"

	"github.com/ghodss/yaml"
	"k8s.io/klog"
)

//ApplierFuncMap adds the function map
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

func toYaml(o interface{}) (string, error) {
	m, err := yaml.Marshal(o)
	if err != nil {
		klog.Error(err)
		return "", err
	}
	klog.V(5).Infof("\n%s", string(m))
	return string(m), nil
}

func encodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

//TemplateFuncMap generates function map for "include"
func TemplateFuncMap(tmpl *template.Template) (funcMap template.FuncMap) {
	funcMap = make(template.FuncMap, 0)
	funcMap["include"] = func(name string, data interface{}) (string, error) {
		buf := bytes.NewBuffer(nil)
		if err := tmpl.ExecuteTemplate(buf, name, data); err != nil {
			return "", err
		}
		return buf.String(), nil
	}
	return funcMap
}
