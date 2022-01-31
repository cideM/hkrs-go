package richtext

import (
	"bytes"
	_ "embed"
	"html/template"

	"github.com/pkg/errors"
	"hochtourenkurs.at/pkg/contentful"
)

type Mark struct {
	MarkType string `json:"type"`
}

type Node struct {
	NodeType string
	Value    *string
	Marks    []Mark
	Data     map[string]string
	Content  []Node
}

//go:embed node.tmpl
var templateFile string
var tmpl = template.Must(template.New("node").Parse(templateFile))

func (n Node) Render(rctx contentful.RenderContext) (template.HTML, error) {
	data := struct {
		Node          Node
		RenderContext contentful.RenderContext
	}{
		Node:          n,
		RenderContext: rctx,
	}
	var b bytes.Buffer
	t := tmpl.Lookup(n.NodeType)
	if t == nil {
		return "", errors.Errorf("unknown node type %s", n.NodeType)
	}

	err := t.Execute(&b, data)
	return template.HTML(b.String()), err
}
