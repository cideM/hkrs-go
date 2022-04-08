package dates

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"hochtourenkurs.at/pkg/contentful"
	"hochtourenkurs.at/pkg/contentful/link"
	"hochtourenkurs.at/pkg/scrape"
)

//go:embed dates.tmpl
var templateFile string
var tmpl = template.Must(template.New("dates").Parse(templateFile))

const (
	basicURL     = "https://www.bergpuls.at/angebote/hochtourenkurs-gletscherkurs-fuer-anfaenger-basic/#buchen"
	basicAttr    = "attribute_pa_termine-hochtouren"
	advancedURL  = "https://www.bergpuls.at/angebote/hochtourenkurs-gletscherkurs-aufbaumodul-advanced/#buchen"
	advancedAttr = "attribute_pa_termine-htk-advanced"
)

type Dates struct {
	Dates      []scrape.AvailabilityDate
	Title      string    `json:"title"`
	CourseType string    `json:"course"`
	Icon       link.Link `json:"icon"`
}

func New(s json.RawMessage) (Dates, error) {
	var f Dates
	err := json.Unmarshal(s, &f)
	if err != nil {
		return f, err
	}

	var url, attr string
	switch f.CourseType {
	case "basic":
		url = basicURL
		attr = basicAttr
	case "advanced":
		url = advancedURL
		attr = advancedAttr
	default:
		return f, errors.Errorf("unknown course type %s", f.CourseType)
	}

	dates, err := scrape.GetDates(url, attr)
	if err != nil {
		return f, err
	}
	f.Dates = dates
	return f, err
}

type renderDate struct {
	Available      bool
	From, To, Text string
}

var digitRe *regexp.Regexp = regexp.MustCompile(`\d`)

func (b Dates) Type() string {
	return "basicProduct"
}

func (b Dates) Render(rctx contentful.RenderContext) (template.HTML, error) {
	checkCircle, err := os.ReadFile("check-circle.svg")
	if err != nil {
		return "", err
	}
	r := strings.NewReplacer(
		"Monday", "Mo",
		"Tuesday", "Di",
		"Wednesday", "Mi",
		"Thursday", "Do",
		"Friday", "Fr",
		"Saturday", "Sa",
		"Sunday", "So",
	)

	renderDates := make([]renderDate, len(b.Dates))
	for i, d := range b.Dates {
		rd := renderDate{
			Available: digitRe.Match([]byte(d.Text)),
			From:      r.Replace(fmt.Sprintf("%s", d.From.Format("Monday 02.01.2006"))),
			To:        r.Replace(fmt.Sprintf("%s", d.To.Format("Monday 02.01.2006"))),
			Text:      d.Text,
		}
		renderDates[i] = rd
	}

	icon, ok := rctx.Assets[b.Icon.Sys.ID]
	if !ok {
		return "", errors.Errorf("no asset for ID %s", b.Icon.Sys.ID)
	}
	svg, ok := rctx.SVGs[icon.Fields.File.Filename]
	if !ok {
		return "", errors.Errorf("no svg for filename %s", icon.Fields.File.Filename)
	}

	data := struct {
		Title         string
		Icon          template.HTML
		RenderContext contentful.RenderContext
		CheckIcon     template.HTML
		Dates         []renderDate
	}{
		Title:         b.Title,
		Icon:          template.HTML(svg),
		CheckIcon:     template.HTML(checkCircle),
		RenderContext: rctx,
		Dates:         renderDates,
	}

	var buff bytes.Buffer
	err = tmpl.Execute(&buff, data)
	if err != nil {
		return "", err
	}

	return template.HTML(buff.String()), nil
}
