package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/som.subhojit1988/hackernews-client/hnscrapper"
	"golang.org/x/net/html"
)

func fetchHnPage(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "hn-cliclient/1 golang cli client for scrapping HN posts")
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	return client.Do(req)
}

func parseHN(logger *log.Logger) {
	hnURL := "https://news.ycombinator.com/news"
	response, err := fetchHnPage(hnURL)
	if err != nil {
		logger.Fatal(err)
	}
	if response.StatusCode != 200 {
		log.Printf("could not fetch %s responseCode=%d", hnURL, response.StatusCode)
	}

	b := response.Body
	z := html.NewTokenizer(b)
	hnscrapper.ParsePosts(z, logger)
}

func main() {
	logger := log.New(os.Stdout, "HN-scrapper", log.Lshortfile)
	parseHN(logger)
}
