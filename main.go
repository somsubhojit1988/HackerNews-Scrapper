package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/html"
)

const (
	tablerowTag  = "tr" // <tr class="athing" id=$postId>
	tabledataTag = "td"
	anchorTag    = "a" // <a class="storylink" href=$storyURL>
	spanTag      = "span"

	athing    = "athing"
	storylink = "storylink"
	title     = "title"
	from      = "from?"
	sitestr   = "sitestr"

	stateUnint   = -1
	stateID      = 0
	stateURL     = 1
	stateTitle   = 2
	stateSiteStr = 3
)

// ParseError indicates error parsing a html.Token
type ParseError struct {
	msg string
}

func (ptr *ParseError) Error() string {
	return ptr.msg
}

// Post represent each post item from the Hacker News wall
type Post struct {
	ID      int
	URL     string
	Title   string
	SiteStr string

	Points        int
	User          string
	Posttime      string
	CommentsCount int
}

func (p *Post) String() string {
	return fmt.Sprintf(
		` -- Post -- 
		id        =   %d
		URL       =   %s
		Title     =   %s
		SiteStr   =   %s
		Points    =   %d
		User      =   %s
		Posttime  =   %s
		nComments =   %d`,
		p.ID,
		p.URL,
		p.Title,
		p.SiteStr,
		p.Points,
		p.User,
		p.Posttime,
		p.CommentsCount)
}

func attributes(t *html.Token) map[string]string {
	attrs := make(map[string]string)
	for _, a := range t.Attr {
		attrs[a.Key] = a.Val
	}
	return attrs
}

func tokenType(t *html.Token) string {
	return t.Data
}

func isAthing(t *html.Token) (bool, map[string]string) {
	attrs := attributes(t)
	cls, ok := attrs["class"]
	return ok && cls == athing, attrs
}

func isStoryLink(t *html.Token) (bool, map[string]string) {
	attrs := attributes(t)
	cls, ok := attrs["class"]
	return ok && cls == storylink, attrs
}

func isClass(t *html.Token, clsType string) (bool, map[string]string) {
	attrs := attributes(t)
	cls, ok := attrs["class"]
	return ok && cls == clsType, attrs
}

func parsePosts(z *html.Tokenizer) ([]*Post, error) {
	var post *Post
	parseErr := &ParseError{}
	posts := []*Post{}

	state := stateUnint

	for {
		tt := z.Next()

		switch tt {
		case html.StartTagToken:
			t := z.Token()
			switch tokenType(&t) {
			case tablerowTag:
				if isathing, attrs := isAthing(&t); isathing {
					// log.Printf("athing: %v\n", attrs)
					// parse Id
					id, err := strconv.Atoi(attrs["id"])
					if err != nil {
						parseErr.msg = "ERROR: parsing <TR CLASS=ATHING ID=$id >\n"
						continue
					}
					post = &Post{ID: id}
					posts = append(posts, post)
					state = stateID
					// log.Printf("creating Post(Id: %d)\n", id)
				}
			case anchorTag:
				if isStryLnk, attrs := isStoryLink(&t); isStryLnk {
					stryURL, ok := attrs["href"]
					if !ok {
						parseErr.msg += "ERROR: <a class=\"storylink\"> missing href"
						continue
					}
					post.URL = stryURL
					state = stateURL
				}
			case spanTag:
				if isSiteStr, _ := isClass(&t, sitestr); isSiteStr {
					state = stateSiteStr
				}
			}
		case html.TextToken:
			// log.Printf("text token: %v\n", string(z.Text()))
			switch state {
			case stateURL:
				post.Title = string(z.Text())
				state = stateTitle
			case stateSiteStr:
				post.SiteStr = string(z.Text())
				state = stateUnint
			}
		case html.ErrorToken:
			log.Printf("Error token")
			if len(posts) >= 0 {
				return posts, nil
			}
			return nil, parseErr
		}
	}
}

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

func parseHN() {
	hnURL := "https://news.ycombinator.com/news"
	response, err := fetchHnPage(hnURL)
	if err != nil {
		err = &ParseError{fmt.Sprintf("ERROR fetching HN page: %s", err.Error())}
	}
	if response.StatusCode != 200 {
		err = &ParseError{fmt.Sprintf("ERROR: HN responded with %s", response.Status)}
	}

	if err != nil {
		log.Fatalln(err.Error())
	}
	b := response.Body
	z := html.NewTokenizer(b)

	posts, err := parsePosts(z)
	if err != nil {
		log.Fatalln(err.Error())
	}

	for _, p := range posts {
		log.Println(p.String())
	}
}

func main() {
	parseHN()

}
