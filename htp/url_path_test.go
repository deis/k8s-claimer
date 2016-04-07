package htp

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/arschles/assert"
)

type splitPathTestCase struct {
	r        *http.Request
	expected []string
}

func TestSplitPath(t *testing.T) {
	testCases := []splitPathTestCase{
		splitPathTestCase{r: &http.Request{URL: &url.URL{Path: ""}}, expected: nil},
		splitPathTestCase{r: &http.Request{URL: &url.URL{Path: "/"}}, expected: nil},
		splitPathTestCase{r: &http.Request{URL: &url.URL{Path: "/a"}}, expected: []string{"a"}},
		splitPathTestCase{r: &http.Request{URL: &url.URL{Path: "/a/b/c"}}, expected: []string{"a", "b", "c"}},
	}

	for _, testCase := range testCases {
		spl := SplitPath(testCase.r)
		assert.Equal(t, len(spl), len(testCase.expected), fmt.Sprintf("number of split return values for %s", testCase.r.URL.Path))
		for i, ex := range testCase.expected {
			assert.Equal(t, spl[i], ex, fmt.Sprintf("split value %d", i))
		}
	}
}
