package entries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"hochtourenkurs.at/pkg/contentful/assets"
	"hochtourenkurs.at/pkg/contentful/sys"
)

// Item is anything defined as a Content Type in a space. Assets are built-in
// and therefore not defined as a Content Type. The fields value depends
// entirely on the kind of Item and is therefore not specified any further.
type Item struct {
	Sys    ItemSys         `json:"sys"`
	Fields json.RawMessage `json:"fields"`
}

// ItemSys is added by Contentful and includes, among other things, the Content
// Type
type ItemSys struct {
	ContentType ContentType `json:"contentType"`
	ID          string      `json:"id"`
}

type ContentType struct {
	Sys sys.Sys `json:"sys"`
}

// Includes should always be present but I think all its contents are optional.
// If no assets are linked to any of the entries in Response, then there won't
// be an .Asset field.
type Includes struct {
	Asset []assets.Asset `json:"Asset"`
}

// Response we get from /entries
type Response struct {
	Total    int      `json:"total"`
	Skip     int      `json:"skip"`
	Limit    int      `json:"limit"`
	Includes Includes `json:"includes"`
	Items    []Item   `json:"items"`
}

type API struct {
	host, accessToken, SpaceID string
	client                     *http.Client
}

func NewAPI() (API, error) {
	httpClient := &http.Client{}
	spaceID := os.Getenv("CONTENTFUL_SPACE_ID")
	if spaceID == "" {
		log.Fatal("CONTENTFUL_SPACE_ID missing")
	}

	token := os.Getenv("CONTENTFUL_API_TOKEN")
	if token == "" {
		log.Fatal("CONTENTFUL_API_TOKEN missing")
	}

	url := os.Getenv("CONTENTFUL_ENDPOINT")
	if url == "" {
		log.Fatal("CONTENTFUL_ENDPOINT missing")
	}

	return API{
		host:        url,
		accessToken: token,
		SpaceID:     spaceID,
		client:      httpClient,
	}, nil
}

func (c API) Get() (Response, error) {
	url := fmt.Sprintf("https://%s/spaces/%s/environments/master/entries", c.host, c.SpaceID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Response{}, err
	}
	req.Header.Add("Authorization", "Bearer "+c.accessToken)
	resp, err := c.client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}
	var entries Response
	err = json.Unmarshal(body, &entries)
	if err != nil {
		return entries, err
	}
	return entries, nil
}

type svg struct {
	name, content string
}

type svgDownload struct {
	name, url string
}

func DownloadSVGs(r Response) (map[string]string, error) {
	downloads := make([]svgDownload, 0)
	for _, v := range r.Includes.Asset {
		if v.Fields.File.ContentType == "image/svg+xml" {
			downloads = append(downloads, svgDownload{
				name: v.Fields.File.Filename, url: "https:" + v.Fields.File.URL,
			})
		}
	}

	sem := semaphore.NewWeighted(100)
	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)

	results := make([]svg, len(downloads))

	for i, v := range downloads {
		if err := sem.Acquire(ctx, 1); err != nil {
			return nil, err
		}
		i := i
		v := v
		g.Go(func() error {
			defer sem.Release(1)
			resp, err := http.Get(v.url)
			if err != nil {
				return err
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			results[i] = svg{name: v.name, content: string(body)}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	if err := sem.Acquire(ctx, 100); err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for _, v := range results {
		m[v.name] = v.content
	}
	return m, nil
}
