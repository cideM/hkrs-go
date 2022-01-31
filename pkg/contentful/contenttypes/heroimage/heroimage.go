package heroimage

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html/template"

	"github.com/pkg/errors"
	"hochtourenkurs.at/pkg/contentful"
	"hochtourenkurs.at/pkg/contentful/link"
)

//go:embed heroimage.tmpl
var templateFile string
var tmpl = template.Must(template.New("heroimage").Parse(templateFile))

type Image struct {
	Title       string
	Description string
	URL         string
	Width       int
	Height      int
}

type HeroImage struct {
	Desktop link.Link `json:"desktop"`
	Mobile  link.Link `json:"mobile"`
}

func New(s json.RawMessage) (HeroImage, error) {
	var img HeroImage
	err := json.Unmarshal(s, &img)
	return img, err
}

func (b HeroImage) Type() string {
	return "heroImage"
}

func (s HeroImage) Render(rctx contentful.RenderContext) (template.HTML, error) {
	desktop, ok := rctx.Assets[s.Desktop.Sys.ID]
	if !ok {
		return "", errors.Errorf("no asset for %s", s.Desktop.Sys.ID)

	}

	mobile, ok := rctx.Assets[s.Mobile.Sys.ID]
	if !ok {
		return "", errors.Errorf("no asset for %s", s.Mobile.Sys.ID)

	}

	pageData := struct {
		Desktop Image
		Mobile  Image
	}{
		Desktop: Image{
			Title:       desktop.Fields.Title,
			Description: desktop.Fields.Description,
			URL:         "https:" + desktop.Fields.File.URL,
			Width:       desktop.Fields.File.Details.Image.Width,
			Height:      desktop.Fields.File.Details.Image.Height,
		},
		Mobile: Image{
			Title:       mobile.Fields.Title,
			Description: mobile.Fields.Description,
			URL:         "https:" + mobile.Fields.File.URL,
			Width:       mobile.Fields.File.Details.Image.Width,
			Height:      mobile.Fields.File.Details.Image.Height,
		},
	}

	var b bytes.Buffer
	err := tmpl.Execute(&b, pageData)
	if err != nil {
		return "", err
	}

	return template.HTML(b.String()), nil
}
