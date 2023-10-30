package scrape

import (
	"encoding/json"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"golang.org/x/net/html"
)

type formData struct {
	AvailabilityHTML string            `json:"availability_html"`
	VariationID      int               `json:"variation_id"`
	Attributes       map[string]string `json:"attributes"`
}

var (
	// "mo-20-09-fr-24-09-2021"
	yearRe *regexp.Regexp = regexp.MustCompile(`(?P<FromDay>\w\w)-(?P<FromDayNum>\d+)-(?P<FromMonth>\d+)-(?P<UntilDay>\w\w)-(?P<UntilDayNum>\d+)-(?P<UntilMonth>\d+)-(?P<Year>\d+)`)
)

func getText(node *html.Node) string {
	var text string
	var fn func(node *html.Node)
	fn = func(n *html.Node) {
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if child.Data == "p" {
				text = child.FirstChild.Data
				return
			} else {
				fn(child)
			}
		}
	}

	fn(node)
	return text
}

type AvailabilityDate struct {
	From, To time.Time
	Text     string
}

func GetDates(url string, attrName string) ([]AvailabilityDate, error) {
	c := colly.NewCollector()
	dates := make([]AvailabilityDate, 0)
	c.OnHTML("form[data-product_variations]", func(n *colly.HTMLElement) {
		jsonS := n.Attr("data-product_variations")
		var data []formData
		err := json.Unmarshal([]byte(jsonS), &data)
		if err != nil {
			log.Fatal(err.Error())
		}

		for _, o := range data {
			date, ok := o.Attributes[attrName]
			if !ok {
				log.Fatal("no attribute_pa_termine-hochtouren")
			}

			match := yearRe.FindStringSubmatch(date)
			var fromDay, fromMonth, toDay, toMonth, year int
			fromDay, err = strconv.Atoi(match[yearRe.SubexpIndex("FromDayNum")])
			if err != nil {
				log.Fatal(err.Error())
			}

			year, err = strconv.Atoi(match[yearRe.SubexpIndex("Year")])
			if err != nil {
				log.Fatal(err.Error())
			}

			fromMonth, err = strconv.Atoi(match[yearRe.SubexpIndex("FromMonth")])
			if err != nil {
				log.Fatal(err.Error())
			}

			from := time.Date(year, time.Month(fromMonth), fromDay, 0, 0, 0, 0, time.UTC)

			toMonth, err = strconv.Atoi(match[yearRe.SubexpIndex("UntilMonth")])
			if err != nil {
				log.Fatal(err.Error())
			}

			toDay, err = strconv.Atoi(match[yearRe.SubexpIndex("UntilDayNum")])
			if err != nil {
				log.Fatal(err.Error())
			}

			to := time.Date(year, time.Month(toMonth), toDay, 0, 0, 0, 0, time.UTC)

			doc, err := html.Parse(strings.NewReader(o.AvailabilityHTML))
			if err != nil {
				log.Fatal(err.Error())
			}

			text := getText(doc)
			dates = append(dates, AvailabilityDate{From: from, To: to, Text: text})
		}
	})
	c.Visit(url)
	return dates, nil
}
