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
	switch t.Data {
	case tablerowTag:
		attrs[tagType] = tablerowTag
	case spanTag:
		attrs[tagType] = spanTag
	case anchorTag:
		attrs[tagType] = anchorTag
	default:
		if t.Type == html.TextToken {
			attrs[tagType] = textToken
			attrs[data] = t.Data
		}
	}

	return attrs
}

func parseTableRow(t *html.Token, psm *PostParsingSM) (*PostParsingSM, error) {
	attrs := attributes(t)
	err := psm.HandleState(attrs)
	if err != nil {
		return psm, err
	}
	return psm, nil
}

func parseAnchor(t *html.Token, psm *PostParsingSM) (*PostParsingSM, error) {
	attrs := attributes(t)
	err := psm.HandleState(attrs)
	if err != nil {
		return psm, err
	}
	return psm, nil
}

func parseSpanTag(t *html.Token, psm *PostParsingSM) (*PostParsingSM, error) {
	attrs := attributes(t)
	err := psm.HandleState(attrs)
	if err != nil {
		return psm, err
	}
	return psm, nil
}

func parseTextTag(t *html.Token, psm *PostParsingSM) (*PostParsingSM, error) {
	attr := attributes(t)
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
		switch tt {
		case html.StartTagToken:
			t := z.Token()
			depth++
			switch t.Data {
			case tablerowTag:
				postSm, err = parseTableRow(&t, postSm)
				if err != nil {
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
					continue
				}
			}
		case html.EndTagToken:
			depth--

		case html.TextToken:
			t := z.Token()
			postSm, err = parseTextTag(&t, postSm)
			if err != nil {
				continue
			}

		case html.ErrorToken:
			logger.Println("returning..")
			return
		}
	}
}
