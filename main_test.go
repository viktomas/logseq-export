package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	testPage := page{
		filename: "",
		attributes: map[string]string{
			"first":  "1",
			"second": "2",
		},
		text: "page text",
	}
	result := render(testPage)
	require.Equal(t, `---
first: 1
second: 2
---
page text`, result)
}
