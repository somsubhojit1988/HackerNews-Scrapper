package hnscrapper

import "golang.org/x/net/html"

import "log"

// ParseError indicates error parsing a html.Token
type ParseError struct {
	msg string
}

const (
	tagType      = "tagType"
	tablerowTag  = "tr" // <tr class="athing" id=$postId>
	tabledataTag = "td"
	anchorTag    = "a" // <a class="storylink" href=$storyURL>
	spanTag      = "span"
	textToken    = "textToken"

	class     = "class"
	id        = "id"
	athing    = "athing"
	storylink = "storylink"
	title     = "title"
	from      = "from?"
	sitestr   = "sitestr"
	score     = "score"
	hnuser    = "hnuser"
	href      = "href"
	data      = "data"
)

func (ptr *ParseError) Error() string {
	return ptr.msg
}

func attributes(t *html.Token) map[string]string {
	attrs := make(map[string]string)
	for _, a := range t.Attr {
		attrs[a.Key] = a.Val
	}

	return attrs
}

func parseTableRow(t *html.Token, psm *PostParsingSM) (*PostParsingSM, error) {
	attrs := attributes(t)
	attrs[tagType] = tablerowTag

	err := psm.HandleState(attrs)
	if err != nil {
		return psm, err
	}
	return psm, nil
}

func parseAnchor(t *html.Token, psm *PostParsingSM) (*PostParsingSM, error) {
	attrs := attributes(t)
	// attrs[tagType] = anchorTag

	// for debug
	// if psm.state == stateID {
	// 	log.Printf("[parseAnchorTag] calling handleState(attrs = %v)", attrs)
	// }

	err := psm.HandleState(attrs)
	if err != nil {
		return psm, err
	}
	return psm, nil
}

func parseSpanTag(t *html.Token, psm *PostParsingSM) (*PostParsingSM, error) {
	attrs := attributes(t)
	attrs[tagType] = spanTag

	// for debug
	// if psm.state == stateSiteStr {
	// }
	// log.Printf("parseSpanTag: calling handleState in state: %s with attrs: %v\n", psm.state.String(), attrs)

	err := psm.HandleState(attrs)
	if err != nil {
		return psm, err
	}
	return psm, nil
}

func parseTextTag(t *html.Token, psm *PostParsingSM) (*PostParsingSM, error) {
	tData := t.Data
	attr := make(map[string]string)
	attr[tagType] = textToken

	attr[data] = tData

	// if psm.state == stateScoreIncoming {
	// 	log.Printf("parseTextTag: calling handleState in state: stateScoreIncoming with attrs: %v\n", attr)
	// }

	err := psm.HandleState(attr)
	if err != nil {
		return psm, err
	}
	return psm, nil
}

// ParsePosts parses the posts iteratively
func ParsePosts(z *html.Tokenizer, logger *log.Logger) {
	postSm := NewPostParsingSM()
	var err error
	depth := 0
	for {
		tt := z.Next()
		// log.Printf("Token type= %s depth= %d\n", tt.String(), depth)
		switch tt {
		case html.StartTagToken:
			t := z.Token()
			depth++
			switch t.Data {
			case tablerowTag:
				postSm, err = parseTableRow(&t, postSm)
				if err != nil {
					// logger.Printf("[ERROR] parsing <tr>: %s\n", err.Error())
					continue
				}
			case spanTag:
				postSm, err = parseSpanTag(&t, postSm)
				if err != nil {
					continue
				}
			case anchorTag:
				postSm, err = parseAnchor(&t, postSm)
				if err != nil {
					// logger.Printf("[ERROR] parsing <a>: %s\n", err.Error())
					continue
				}
			}
		case html.EndTagToken:
			depth--

		case html.TextToken:
			t := z.Token()
			postSm, err = parseTextTag(&t, postSm)
			if err != nil {
				// logger.Printf("[ERROR] parsing text-token: %s\n", err.Error())
				continue
			}

		case html.ErrorToken:
			logger.Println("returning..")
			return
		}
	}
}
