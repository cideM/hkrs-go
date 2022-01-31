package slideshow

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html/template"

	"github.com/pkg/errors"
	"hochtourenkurs.at/pkg/contentful"
	"hochtourenkurs.at/pkg/contentful/link"
)

//go:embed slideshow.tmpl
var templateFile string
var tmpl = template.Must(template.New("slideshow").Parse(templateFile))

type Slideshow struct {
	Images []link.Link `json:"images"`
	Title  string      `json:"title"`
}

func New(s json.RawMessage) (Slideshow, error) {
	var slideshow Slideshow
	err := json.Unmarshal(s, &slideshow)
	return slideshow, err
}

func (b Slideshow) Type() string {
	return "slideshow"
}

func (s Slideshow) Render(rctx contentful.RenderContext) (template.HTML, error) {
	pageData := struct {
		Title         string
		Children      []contentful.Renderer
		RenderContext contentful.RenderContext
	}{
		Title:         s.Title,
		Children:      make([]contentful.Renderer, len(s.Images)),
		RenderContext: rctx,
	}

	for i, v := range s.Images {
		item, ok := rctx.Entries[v.Sys.ID]
		if !ok {
			return "", errors.Errorf("no item for %s", v.Sys.ID)

		}
		pageData.Children[i] = item
	}

	var b bytes.Buffer
	err := tmpl.Execute(&b, pageData)
	if err != nil {
		return "", err
	}

	return template.HTML(b.String()), nil
}
