package page

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html/template"
	"sort"

	"github.com/pkg/errors"
	"hochtourenkurs.at/pkg/contentful"
	"hochtourenkurs.at/pkg/contentful/link"
)

//go:embed page.tmpl
var templateFile string
var tmpl = template.Must(template.New("page").Parse(templateFile))

type Page struct {
	NoIndex        bool
	ID, URL, Title string
	Content        []link.Link
	Slideshow      *link.Link
	CTA            *link.Link
	Description    string
	TabTitle       string
	TopNav         bool
	Position       int
}

type Fields struct {
	URL         string      `json:"url"`
	Blocks      []link.Link `json:"blocks"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Position    int         `json:"position"`
	TabTitle    string      `json:"tabTitle"`
	CTA         *link.Link  `json:"callToAction"`
	TopNav      bool        `json:"topNavigation"`
	Slideshow   *link.Link  `json:"slideshow"`
}

func New(id string, s json.RawMessage, noindex bool) (Page, error) {
	var f Fields
	err := json.Unmarshal(s, &f)
	if err != nil {
		return Page{}, err
	}
	return Page{
		NoIndex:     noindex,
		Slideshow:   f.Slideshow,
		ID:          id,
		Content:     f.Blocks,
		URL:         f.URL,
		Position:    f.Position,
		TopNav:      f.TopNav,
		CTA:         f.CTA,
		Title:       f.Title,
		TabTitle:    f.TabTitle,
		Description: f.Description,
	}, nil
}

type OtherPageLink struct {
	Title, Path string
	TopNav      bool
	Position    int
}

func (b Page) Type() string {
	return "page"
}

func (p Page) Render(rctx contentful.RenderContext) (template.HTML, error) {
	var s contentful.Renderer
	if p.Slideshow != nil {
		item, ok := rctx.Entries[p.Slideshow.Sys.ID]
		if !ok {
			return "", errors.Errorf("no slideshow with ID %s", p.Slideshow.Sys.ID)
		}
		s = item
	}

	var cta contentful.Renderer
	if p.CTA != nil {
		item, ok := rctx.Entries[p.CTA.Sys.ID]
		if !ok {
			return "", errors.Errorf("no CTA with ID %s", p.CTA.Sys.ID)
		}
		cta = item
	}

	bottomLinks := make([]OtherPageLink, 0)
	topLinks := make([]OtherPageLink, 0)
	for _, v := range rctx.Entries {
		if v.Type() == "page" {
			p, ok := v.(Page)
			if !ok {
				return "", errors.New("is page but can't type cast")
			}

			if p.TopNav {
				topLinks = append(topLinks, OtherPageLink{
					Title:    p.Title,
					Path:     p.URL,
					Position: p.Position,
				})
			} else {
				bottomLinks = append(bottomLinks, OtherPageLink{
					Title:    p.Title,
					Path:     p.URL,
					Position: p.Position,
				})
			}
		}
	}

	sort.Slice(topLinks, func(i, j int) bool {
		return topLinks[i].Position < topLinks[j].Position
	})

	logo, ok := rctx.SVGs["nav_logo"]
	if !ok {
		return "", errors.Errorf("no nav logo")
	}

	pageData := struct {
		NavLogo       template.HTML
		CTA           contentful.Renderer
		BottomLinks   []OtherPageLink
		TopLinks      []OtherPageLink
		Slideshow     contentful.Renderer
		Page          Page
		RenderContext contentful.RenderContext
	}{
		NavLogo:       template.HTML(logo),
		CTA:           cta,
		BottomLinks:   bottomLinks,
		TopLinks:      topLinks,
		Slideshow:     s,
		Page:          p,
		RenderContext: rctx,
	}

	var b bytes.Buffer
	err := tmpl.Execute(&b, pageData)
	if err != nil {
		return "", err
	}

	return template.HTML(b.String()), nil
}
