package main

import (
	"fmt"
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

func saveResponse(res *http.Response, logger *log.Logger) {

	file, err := os.Create("html-response.html")
	if err != nil {
		logger.Printf("error creating file: %v\n", err.Error())
		return
	}

	defer file.Close()
	defer res.Body.Close()

	bufferedWrite := func(res *http.Response, file *os.File, logger *log.Logger) int64 {
		buff := make([]byte, 128)
		var writeOffset int64 = 0
		for {
			_, err := res.Body.Read(buff)
			if err != nil {
				logger.Printf("Error reading response: %s\n", err.Error())
				return writeOffset
			}
			n, err := file.WriteAt(buff, writeOffset)
			if err != nil {
				logger.Printf("error writting to %s= %s\n", file.Name(), err.Error())
				return writeOffset
			}
			writeOffset += int64(n)
		}

	}
	n := bufferedWrite(res, file, logger)
	fmt.Printf("# of bytes written %d -> %s\n", n, file.Name())
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
	// posts, err := parsePosts(z)
	// if err != nil {
	// 	log.Fatalln(err.Error())
	// }

	// for _, p := range posts {
	// 	log.Println(p.String())
	// }
	// parsePosts(z)
}

func main() {
	logger := log.New(os.Stdout, "HN-scrapper", log.Lshortfile)
	parseHN(logger)
}

// switch currentState {
// case stateInit:
// 	// Looking for <tr id=$postId class="athing">
// 	// on success: stateInit -> stateId
// case stateID:
// 	// Looking for storylink <a class = "storylink" href=$postURL
// 	// on success: stateId -> stateStoryLink
// case stateStoryLink:
// 	// Looking for postTitle TextToken**
// 	// on success: stateStoryLink ->stateStoryTitle
// case stateStoryTitle:
// 	// Looking for <span class="sitestr">
// 	// on success: stateStoryTitle ->stateSiteStrIncoming
// case stateSitestrIncoming:
// 	// Looking for postSiteString TextToken**
// 	// onSuccess: stateSiteStrIncoming -> stateSiteStr
// case stateSiteStr:
// 	// Looking for <span id ="score_$postId" class=score>
// 	// onSuccess: stateSiteStr -> stateScoreIncoming
// case stateScoreIncoming:
// 	// Looking $score points TextToken**
// 	// onSuccess: stateScoreIncoming -> stateScore
// case stateScore:
// 	// Looking for <a class="hnuser" href="user?id=$hnuser">
// 	// onSuccess: stateScore -> stateHnuserIncoming
// case stateHnuserIncoming:
// 	// Looking for hnuser TextToken**
// 	// onSuccess: stateHnuserIncoming -> stateHnuser
// case stateHnuser:
// 	// Looking for <span class="age">
// 	// onSuccess: stateHnuser -> stateAgeIncoming1
// case stateAgeIncoming1:
// 	// Looking for <a id="item?id=$postID">
// 	// onSuccess: stateAgeIncoming1 -> stateAgeIncoming2
// case stateAgeIncoming2:
// 	// Looking for $age hours ago TextToken**
// 	// onSuccess: stateAgeIncoming2 -> stateAge
// case stateAge:
// 	// Looking for "hide" TextToken**
// 	// stateAge -> stateNCommentsIncoming
// case stateNCommentsIncoming:
// 	// Looking for $nComments Comments TextToken**
// 	// stateNCommentsIncoming -> stateNComments
// case stateNComments:
// 	// push *post into posts[]*Post
// 	// stateNComments -> stateInit
// }
