package richtext

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func strp(s string) *string {
	return &s
}

func Test_Container(t *testing.T) {
	input := `
		{
		  "nodeType": "paragraph",
		  "content": [
			{
			  "nodeType": "text",
			  "value": "foo",
			  "marks": [],
			  "data": {}
			}
		  ],
		  "data": {}
		}
	`

	var n Node
	err := json.Unmarshal([]byte(input), &n)
	assert.NoError(t, err)
	expect := Node{
		NodeType: "paragraph",
		Data:     make(map[string]string),
		Content: []Node{
			{
				NodeType: "text",
				Value:    strp("foo"),
				Marks:    make([]string, 0),
				Data:     make(map[string]string),
			},
		},
	}
	assert.Equal(t, expect, n)
}
