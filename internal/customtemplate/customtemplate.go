package customtemplate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

func Execute(content any, variables map[string]any) (any, error) {
	tmpl, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content before applying template: %w", err)
	}

	const token = "@@RAW@@"
	funcMap := template.FuncMap{
		"raw": func(v any) string {
			return fmt.Sprintf("%s%v%s", token, v, token)
		},
	}

	parsedTmpl, err := template.New("t").Funcs(funcMap).Parse(string(tmpl))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var tmplResult bytes.Buffer
	if err := parsedTmpl.Execute(&tmplResult, variables); err != nil {
		return nil, fmt.Errorf("failed to apply template: %w", err)
	}

	jsonBytes := tmplResult.Bytes()
	jsonBytes = bytes.ReplaceAll(jsonBytes, []byte(`"`+token), []byte{})
	jsonBytes = bytes.ReplaceAll(jsonBytes, []byte(token+`"`), []byte{})

	var result any
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("template produced invalid JSON: %w", err)
	}
	return result, nil
}
