package cta

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html/template"

	"github.com/pkg/errors"

	"hochtourenkurs.at/pkg/contentful"
	"hochtourenkurs.at/pkg/contentful/link"
)

//go:embed cta.tmpl
var templateFile string
var tmpl = template.Must(template.New("cta").Parse(templateFile))

type CTA struct {
	Title string     `json:"title"`
	URL   string     `json:"url"`
	Icon  *link.Link `json:"icon"`
}

func New(s json.RawMessage) (CTA, error) {
	var f CTA
	err := json.Unmarshal(s, &f)
	return f, err
}

func (b CTA) Type() string {
	return "purchaseButton"
}

func (b CTA) Render(rctx contentful.RenderContext) (template.HTML, error) {
	var svg string
	if b.Icon != nil {
		asset, ok := rctx.Assets[b.Icon.Sys.ID]
		if !ok {
			return "", errors.Errorf("no asset for ID %s", b.Icon.Sys.ID)
		}
		svgIcon, ok := rctx.SVGs[asset.Fields.File.Filename]
		if !ok {
			return "", errors.Errorf("no svg for filename %s", asset.Fields.File.Filename)
		}
		svg = svgIcon
	}

	pageData := struct {
		CTA  CTA
		Icon template.HTML
	}{
		CTA:  b,
		Icon: template.HTML(svg),
	}

	var buff bytes.Buffer
	err := tmpl.Execute(&buff, pageData)
	if err != nil {
		return "", err
	}

	return template.HTML(buff.String()), nil
}
