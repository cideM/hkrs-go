package link

import (
	"html/template"

	"github.com/pkg/errors"
	"hochtourenkurs.at/pkg/contentful"
	"hochtourenkurs.at/pkg/contentful/sys"
)

type Link struct {
	Sys sys.Sys `json:"sys"`
}

func (l Link) Render(rctx contentful.RenderContext) (template.HTML, error) {
	item, ok := rctx.Entries[l.Sys.ID]
	if !ok {
		return "", errors.Errorf("no item for %s", l.Sys.ID)
	}

	return item.Render(rctx)
}
