package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/pkg/errors"
	"hochtourenkurs.at/pkg/contentful"
	"hochtourenkurs.at/pkg/contentful/assets"
	"hochtourenkurs.at/pkg/contentful/contenttypes/block"
	"hochtourenkurs.at/pkg/contentful/contenttypes/cta"
	"hochtourenkurs.at/pkg/contentful/contenttypes/dates"
	"hochtourenkurs.at/pkg/contentful/contenttypes/heroimage"
	"hochtourenkurs.at/pkg/contentful/contenttypes/page"
	"hochtourenkurs.at/pkg/contentful/contenttypes/slideshow"
	"hochtourenkurs.at/pkg/contentful/entries"
)

var outDir string = "public"

func writePage(p page.Page, rendered string) error {
	filename := fmt.Sprintf("%s/%s.html", outDir, p.URL)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	_, err = file.Write([]byte(rendered))
	return err
}

func main() {
	entriesAPI, err := entries.NewAPI()
	if err != nil {
		log.Fatal(err)
	}

	noindexS := os.Getenv("NOINDEX")
	var noindex bool
	switch noindexS {
	case "true":
		noindex = true
	case "false":
		noindex = false
	default:
		log.Fatalf("unknown NOINDEX %s", noindexS)
	}

	allEntries, err := entriesAPI.Get()
	if err != nil {
		log.Fatal(err)
	}

	svgs, err := entries.DownloadSVGs(allEntries)
	if err != nil {
		log.Fatal(err)
	}

	navLogo, err := os.ReadFile("nav_logo.svg")
	if err != nil {
		log.Fatal(err)
	}
	svgs["nav_logo"] = string(navLogo)

	assetsByID := make(map[string]assets.Asset)
	entriesByID := make(map[string]contentful.Renderer)

	pages := make([]page.Page, 0)

	for _, i := range allEntries.Items {
		switch i.Sys.ContentType.Sys.ID {
		case "page":
			s, err := page.New(i.Sys.ID, i.Fields, noindex)
			if err != nil {
				log.Fatal(err)
			}
			pages = append(pages, s)
			entriesByID[i.Sys.ID] = s
		case "slideshow":
			s, err := slideshow.New(i.Fields)
			if err != nil {
				log.Fatal(err)
			}
			entriesByID[i.Sys.ID] = s
		case "block":
			bl, err := block.New(i.Fields)
			if err != nil {
				log.Fatal(err)
			}
			entriesByID[i.Sys.ID] = bl
		case "basicProduct":
			v, err := dates.New(i.Fields)
			if err != nil {
				log.Fatal(err)
			}
			entriesByID[i.Sys.ID] = v
		case "heroImage":
			v, err := heroimage.New(i.Fields)
			if err != nil {
				log.Fatal(err)
			}
			entriesByID[i.Sys.ID] = v
		case "purchaseButton":
			v, err := cta.New(i.Fields)
			if err != nil {
				log.Fatal(err)
			}
			entriesByID[i.Sys.ID] = v
		default:
			log.Fatal(errors.Errorf("unknown content type %s", i.Sys.ContentType.Sys.ID))
		}
	}

	for _, v := range allEntries.Includes.Asset {
		assetsByID[v.Sys.ID] = v
	}

	renderContext := contentful.RenderContext{
		Assets:  assetsByID,
		Entries: entriesByID,
		SVGs:    svgs,
	}

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Position < pages[j].Position
	})

	for _, p := range pages {
		s, err := p.Render(renderContext)
		if err != nil {
			log.Fatal(err)
		}
		err = writePage(p, string(s))
		if err != nil {
			log.Fatal(err)
		}
	}
}
