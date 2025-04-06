package utils

import (
	"fmt"
	"html/template"
	"time"
)

func TemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"currentYear": func() int {
			return time.Now().Year()
		},
		"dict": Dict,
	}
}

func Dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("invalid dict call")
	}

	dict := make(map[string]interface{}, len(values)/2)

	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}

	return dict, nil
}
