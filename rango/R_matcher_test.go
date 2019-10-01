package rango

import (
	"net/http"
	"testing"
)

func newEmptyRequest(m string) *http.Request {
	req, _ := http.NewRequest(m, "loacalhost", nil)
	return req
}

func newURLRequest(u string) *http.Request {
	req, _ := http.NewRequest("GET", u, nil)
	return req
}

type matcherTestCase struct {
	desc    string
	matcher matcher

	req      *http.Request
	expected bool
}

func testCaseOnMatcher(t *testing.T, cases []matcherTestCase) {
	for _, testCase := range cases {
		actual := testCase.matcher.Match(testCase.req)
		if testCase.expected != actual {
			t.Fatalf("%v is fatal, need %v but %v.", testCase.desc, testCase.expected, actual)
		}
	}
}

func TestHeaderMatcher(t *testing.T) {
	header := map[string]string{
		"X-token": "auth signed",
	}
	hm := headerMatcher(header)

	req := newEmptyRequest("GET")
	req2 := newEmptyRequest("GET")
	req2.Header.Set("X-token", "auth signed")

	testCaseOnMatcher(t, []matcherTestCase{
		{"headerMatcher.Match('X-token')", hm, req, false},
		{"headerMatcher.Match('X-token')", hm, req2, true},
	})
}

func TestMethodMatcher(t *testing.T) {
	ms1 := []string{"GET", "POST"}
	matcher1 := methodMatcher(ms1)

	ms2 := []string{"Get"}
	matcher2 := methodMatcher(ms2)

	getReq := newEmptyRequest("GET")
	postReq := newEmptyRequest("POST")
	putReq := newEmptyRequest("PUT")

	testCaseOnMatcher(t, []matcherTestCase{
		{"methodMatcher.Match(GET)", matcher1, getReq, true},
		{"methodMatcher.Match(POST)", matcher1, postReq, true},
		{"methodMatcher.Match(PUT)", matcher1, putReq, false},
		{"methodMatcher.Match(GET)", matcher2, getReq, true},
		{"methodMatcher.Match(POST)", matcher2, postReq, false},
		{"methodMatcher.Match(PUT)", matcher2, putReq, false},
	})
}
func TestPathMappingMatcher(t *testing.T) {
	pm1 := pathMappingMatcher("/")
	pm2 := pathMappingMatcher("/home")

	indexReq := newURLRequest("/")
	homeReq := newURLRequest("/home")
	errorReq := newURLRequest("/error")

	testCaseOnMatcher(t, []matcherTestCase{
		{"pathMapping.Match(/)", pm1, indexReq, true},
		{"pathMapping.Match(/home)", pm1, homeReq, false},
		{"pathMapping.Match(/error)", pm1, errorReq, false},
		{"pathMapping.Match(/)", pm2, indexReq, false},
		{"pathMapping.Match(/home)", pm2, homeReq, true},
		{"pathMapping.Match(/error)", pm2, errorReq, false},
	})
}
func TestPathMatcher(t *testing.T) {
	p1 := newPathMatcher("/api/{arg:\\w+}", false)
	p2 := newPathMatcher("/api/{arg:\\d+}", false)
	p3 := newPathMatcher("/assert", false)

	p4 := newPathMatcher("/home/", true)
	p5 := newPathMatcher("/home/", false)

	indexReq := newURLRequest("/")
	homeReq := newURLRequest("/home")
	homeIndexReq := newURLRequest("/home/")
	apiIndexReq := newURLRequest("/api/")
	wordReq := newURLRequest("/api/word")
	numReq1 := newURLRequest("/api/1")
	numReq2 := newURLRequest("/api/10")

	htmlReq := newURLRequest("/assert/index.html")
	imgReq := newURLRequest("/assert/logo.png")
	jsonReq := newURLRequest("/assert/config.json")

	testCaseOnMatcher(t, []matcherTestCase{
		// /api/{arg:{\\w+}}
		{"Path(/api/{arg:\\w+}).Match(/)", p1, indexReq, false},
		{"Path(/api/{arg:\\w+}).Match(/home)", p1, homeReq, false},
		{"Path(/api/{arg:\\w+}).Match(/api/)", p1, apiIndexReq, false},
		{"Path(/api/{arg:\\w+}).Match(/api/word)", p1, wordReq, true},
		{"Path(/api/{arg:\\w+}).Match(/api/1)", p1, numReq1, true},
		{"Path(/api/{arg:\\w+}).Match(/api/10)", p1, numReq2, true},
		{"Path(/api/{arg:\\w+}).Match(/assert/index.html)", p1, htmlReq, false},
		{"Path(/api/{arg:\\w+}).Match(/assert/logo.png)", p1, imgReq, false},
		{"Path(/api/{arg:\\w+}).Match(/assert/config.json)", p1, jsonReq, false},
		// /api/{arg:{\\d+}}
		{"Path(/api/{arg:\\d+}).Match(/)", p2, indexReq, false},
		{"Path(/api/{arg:\\d+}).Match(/home)", p2, homeReq, false},
		{"Path(/api/{arg:\\d+}).Match(/api/)", p2, apiIndexReq, false},
		{"Path(/api/{arg:\\d+}).Match(/api/word)", p2, wordReq, false},
		{"Path(/api/{arg:\\d+}).Match(/api/1)", p2, numReq1, true},
		{"Path(/api/{arg:\\d+}).Match(/api/10)", p2, numReq2, true},
		{"Path(/api/{arg:\\d+}).Match(/assert/index.html)", p2, htmlReq, false},
		{"Path(/api/{arg:\\d+}).Match(/assert/logo.png)", p2, imgReq, false},
		{"Path(/api/{arg:\\d+}).Match(/assert/config.json)", p2, jsonReq, false},
		// /assert
		{"Path(/assert).Match(/)", p3, indexReq, false},
		{"Path(/assert).Match(/home)", p3, homeReq, false},
		{"Path(/assert).Match(/api/)", p3, apiIndexReq, false},
		{"Path(/assert).Match(/api/word)", p3, wordReq, false},
		{"Path(/assert).Match(/api/1)", p3, numReq1, false},
		{"Path(/assert).Match(/api/10)", p3, numReq2, false},
		{"Path(/assert).Match(/assert/index.html)", p3, htmlReq, true},
		{"Path(/assert).Match(/assert/logo.png)", p3, imgReq, true},
		{"Path(/assert).Match(/assert/config.json)", p3, jsonReq, true},
		// strictSlash test
		{"Path(/home/, true).Match(/home)", p4, homeReq, true},
		{"Path(/home/, true).Match(/home/)", p4, homeIndexReq, true},
		{"Path(/home/, false).Match(/home)", p5, homeReq, false},
		{"Path(/home/, false).Match(/home/)", p5, homeIndexReq, true},
	})
}

func TestBraceIndices(t *testing.T) {
	if _, err := braceIndices("{{}"); err == nil {
		t.Fatalf("braceIndices({{}) is fatal, need ERROR but nil.")
	}

	if _, err := braceIndices("{}}"); err == nil {
		t.Fatalf("braceIndices({}}) is fatal, need ERROR but nil.")
	}
}
