package assets

// Asset is a binary file in a Contentful space
type Asset struct {
	Sys    Sys    `json:"sys"`
	Fields Fields `json:"fields"`
}

type Sys struct {
	ID      string `json:"id"`
	SysType string `json:"type"`
}

type Fields struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	File        File   `json:"file"`
}

type File struct {
	URL         string      `json:"url"`
	ContentType string      `json:"contentType"`
	Filename    string      `json:"fileName"`
	Details     FileDetails `json:"details"`
}

// FileDetails further describes the file. Not every file is an image, so Image
// is made optional.
type FileDetails struct {
	Size  int    `json:"size"`
	Image *Image `json:"image"`
}

type Image struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}
