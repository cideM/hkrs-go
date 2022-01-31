package block

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html/template"

	"github.com/pkg/errors"
	"hochtourenkurs.at/pkg/contentful"
	"hochtourenkurs.at/pkg/contentful/link"
	"hochtourenkurs.at/pkg/contentful/richtext"
)

//go:embed block.tmpl
var templateFile string
var tmpl = template.Must(template.New("block").Parse(templateFile))

type Block struct {
	Content richtext.Node `json:"content"`
	Title   string        `json:"title"`
	Icon    *link.Link    `json:"titleIcon"`
}

func New(s json.RawMessage) (Block, error) {
	var f Block
	err := json.Unmarshal(s, &f)
	return f, err
}

func (b Block) Type() string {
	return "block"
}

func (b Block) Render(rctx contentful.RenderContext) (template.HTML, error) {
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

	data := struct {
		Title         string
		Icon          template.HTML
		Content       richtext.Node
		RenderContext contentful.RenderContext
	}{
		Title:         b.Title,
		Icon:          template.HTML(svg),
		RenderContext: rctx,
		Content:       b.Content,
	}

	var buff bytes.Buffer
	err := tmpl.Execute(&buff, data)
	if err != nil {
		return "", err
	}

	return template.HTML(buff.String()), nil
}
