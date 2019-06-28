package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var MediawikiAPIRoot string = "https://www.wikipedia.org/w/api.php"

type wikiResponse struct {
	Query struct {
		Pages []*WikiArticle `json:"pages"`
	} `json:"query"`
}

type WikiArticle struct {
	ID      int    `json:"pageid"`
	Title   string `json:"title"`
	Extract string `json:"extract"`
	Missing bool   `json:"missing"`
}

// GetWikiArticle returns the text of the article with the given title
func GetWikiArticle(title string) (article *WikiArticle, err error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", MediawikiAPIRoot, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "github.com/wgoodall01/smssh")

	q := req.URL.Query()
	q.Add("action", "query")
	q.Add("format", "json")
	q.Add("prop", "extracts")
	q.Add("explaintext", "true")
	q.Add("exlimit", "1")
	q.Add("formatversion", "2")
	q.Add("titles", title)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)
	var wikiResp wikiResponse
	err = decoder.Decode(&wikiResp)
	if err != nil {
		return nil, err
	}

	pages := wikiResp.Query.Pages
	if len(pages) != 1 {
		return nil, fmt.Errorf("query should return  1 article, got %d", len(pages))
	}

	page := pages[0]
	if page.Missing {
		return nil, fmt.Errorf("page \"%s\" does not exist", title)
	}

	return page, nil
}
