package packer

import (
	"html/template"
	"strings"
)

const basePageCSS = `
div {
    display: none
}

h1, h2 {
	text-align: center;
}

img {
    display: block;
    vertical-align: baseline;
    margin: 0;
    padding: 0;
}`

var imagePageTemplate = template.Must(template.New("page").Parse(`<div>.</div><img src="kindle:embed:{{ . }}?mime=image/jpeg">`))
var emptyPageTemplate = template.Must(template.New("page").Parse(`<div>.</div><h1>Empty chapter</h1><h2>(╯°□°）╯︵ ┻━┻</h2>`))

func templateToString(tpl *template.Template, data interface{}) string {
	buf := new(strings.Builder)
	if err := tpl.Execute(buf, data); err != nil {
		panic(err)
	}

	return buf.String()
}
