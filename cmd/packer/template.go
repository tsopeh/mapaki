package packer

import (
	"html/template"
	"strings"
)

const basePageCSS = `
div {
    display: none
}

img {
    display: block;
    vertical-align: baseline;
    margin: 0;
    padding: 0;
}`

var pageTemplate = template.Must(template.New("page").Parse(`<div>.</div><img src="kindle:embed:{{ . }}?mime=image/jpeg">`))

func templateToString(tpl *template.Template, data interface{}) string {
	buf := new(strings.Builder)
	if err := tpl.Execute(buf, data); err != nil {
		panic(err)
	}

	return buf.String()
}
