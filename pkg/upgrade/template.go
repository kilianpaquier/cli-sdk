package upgrade

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	_majorRegexp = regexp.MustCompile("^v[0-9]+$")
	_minorRegexp = regexp.MustCompile(`^v[0-9]+\.[0-9]+$`)
	_wordRegexp  = regexp.MustCompile(`[a-zA-Z]+`)
)

// getTemplateValue takes an input template value and returns the parsed result.
func getTemplateValue(value string, data map[string]any) (string, error) {
	tmpl, err := template.New("file").
		Funcs(funcMap()).
		Parse(value)
	if err != nil {
		return "", fmt.Errorf("template parse: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute: %w", err)
	}
	return buf.String(), nil
}

// funcMap returns the text/template function map for upgrade templating (asset name, target name, etc.).
func funcMap() template.FuncMap {
	return template.FuncMap{
		"lower": strings.ToLower,
		"title": cases.Title(language.Und, cases.NoLower, cases.Compact).String,
		"upper": strings.ToUpper,
	}
}
