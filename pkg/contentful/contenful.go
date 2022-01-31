package contentful

import (
	"html/template"

	"hochtourenkurs.at/pkg/contentful/assets"
)

type Renderer interface {
	Render(RenderContext) (template.HTML, error)
	Type() string
}

type RenderContext struct {
	Assets  map[string]assets.Asset
	Entries map[string]Renderer
	SVGs    map[string]string
}
