package hnscrapper

import "strconv"

import "log"

import "fmt"

import "strings"

// State pertaining to the parsing state machine
type State int

// PostParsingSM is a container for the state machine of parsing a post
type PostParsingSM struct {
	state State
	post  *Post
}

// SMError indicates error while handling current state
type SMError struct {
	msg string
}

func (p *SMError) Error() string {
	return p.msg
}

const (
	// states
	// Line 1
	stateInit            State = -1
	stateID              State = 0
	stateStoryLink       State = 1
	stateStoryTitle      State = 2
	stateSitestrIncoming State = 3 // a little peculiar
	stateSiteStr         State = 4

	// Line 2
	stateScoreIncoming     State = 5
	stateScore             State = 6
	stateHnuserIncoming    State = 7
	stateHnuser            State = 8
	stateAgeIncoming1      State = 9
	stateAgeIncoming2      State = 10
	stateAge               State = 11
	stateNCommentsIncoming State = 12
	stateNComments         State = 13
)

func (s State) String() string {
	switch s {
	case stateInit:
		return "state-init"
	case stateID:
		return "state-id"
	case stateStoryLink:
		return "state-storylink"
	case stateStoryTitle:
		return "state-storytitle"
	case stateSitestrIncoming:
		return "state-sitestr-incoming"
	case stateSiteStr:
		return "state-sitestr"
	default:
		return fmt.Sprintf("Unknown state=%d", s)
	}
}

// NewPostParsingSM creates a new instance of PostParsingSM
func NewPostParsingSM() *PostParsingSM {
	return &PostParsingSM{state: stateInit, post: nil}
}

func (psm *PostParsingSM) postInit(attrs map[string]string) error {
	// extract id
	idStr, ok := attrs[id]
	if !ok {
		return &SMError{msg: "could not find id"}
	}
	idVal, err := strconv.Atoi(idStr)
	if err != nil {
		return &SMError{msg: err.Error()}
	}
	psm.post = &Post{ID: idVal}
	return nil
}

func (psm *PostParsingSM) postURL(attrs map[string]string) error {
	url, hasURL := attrs[href]
	if hasURL {
		url = strings.TrimSpace(url)
	} else {
		return &SMError{msg: "could not find href in storylink"}
	}

	if hasProto := strings.Index(url, "http") == 0; hasProto {
		psm.post.URL = url
	}
	return nil
}

func (psm *PostParsingSM) postTitle(attrs map[string]string) error {
	title, ok := attrs[data]
	if !ok {
		return &SMError{msg: "could not find title in Data"}
	}
	psm.post.Title = title
	return nil
}

func (psm *PostParsingSM) postSitestr(attrs map[string]string) error {
	sitestr, ok := attrs[data]
	if !ok {
		return &SMError{msg: "could not find sitestr text in Data"}
	}
	psm.post.SiteStr = sitestr
	return nil
}

func (psm *PostParsingSM) postPoints(attrs map[string]string) error {
	pointStr, ok := attrs[data]

	if !ok {
		return &SMError{msg: "could not find points text in Data"}
	}

	// TODO: Please remove this horrible thing and use Regex to maintain sanity
	strs := strings.Split(strings.TrimSpace(pointStr), " ")
	if len(strs) >= 2 { // INT points
		pts, err := strconv.Atoi(strs[0])
		if err != nil {
			return &SMError{msg: fmt.Sprintf("could not parse points text %s", pointStr)}
		}
		psm.post.Points = pts
	}
	return nil
}

func (psm *PostParsingSM) handleTransitState(attrs map[string]string) error {
	var err *SMError = nil
	switch psm.state {

	case stateStoryTitle:
		val, ok := attrs[class]
		if ok && val == sitestr {
			log.Printf("[state-transition] stateStoryTitle -> stateSitestrIncoming [postID: %d]", psm.post.ID)
			psm.state = stateSitestrIncoming
		}
		err = &SMError{msg: "could not find class = sitestr in <span>"}

	case stateSiteStr:
		clsVal, hasClass := attrs[class]
		idVal, hasID := attrs[id]
		if hasClass && hasID && clsVal == score && idVal == fmt.Sprintf("score_%d", psm.post.ID) {
			log.Printf("[state-transition] stateSiteStr -> stateScoreIncoming [postID: %d]", psm.post.ID)
			psm.state = stateScoreIncoming
		}
		// skip the rest
	}
	return err
}

// HandleState handles the Post State machine based on the passed map of attributes
func (psm *PostParsingSM) HandleState(attrs map[string]string) error {

	switch psm.state {

	case stateInit:
		if attrs[tagType] != tablerowTag && attrs[class] != athing {
			return nil
		}

		err := psm.postInit(attrs)
		if err != nil {
			return err
		}
		log.Printf("[state-transition] stateInit -> stateID [postID: %d]", psm.post.ID)
		psm.state = stateID

	case stateID:
		// Looking for storylink <a class = "storylink" href=$postURL
		if clsVal, ok := attrs[class]; ok && clsVal == storylink {
			err := psm.postURL(attrs)
			if err != nil {
				psm.state = stateInit
				return err
			}
			log.Printf("[state-transition] stateID -> stateStoryLink [postID: %d]", psm.post.ID)
			psm.state = stateStoryLink
		}

	case stateStoryLink:
		err := psm.postTitle(attrs)
		if err != nil {
			return err
		}
		log.Printf("[state-transition] stateStoryLink -> stateStoryTitle [postID: %d]", psm.post.ID)
		psm.state = stateStoryTitle

	case stateStoryTitle:
		psm.handleTransitState(attrs)

	case stateSitestrIncoming:
		err := psm.postSitestr(attrs)
		if err != nil {
			psm.state = stateInit
			return err
		}
		log.Printf("[state-transition] stateSitestrIncoming -> stateSiteStr [postID: %d]", psm.post.ID)
		psm.state = stateSiteStr

	case stateSiteStr:
		psm.handleTransitState(attrs)

		// Line 2
	case stateScoreIncoming:
		err := psm.postPoints(attrs)
		if err != nil {
			log.Printf("handlestate(stateScoreIncoming): failed to parse points err= %v", err)
			psm.state = stateInit
			return err
		}
		psm.state = stateScore
		log.Printf("[state-transition] stateScoreIncoming -> stateScore [postID: %d]", psm.post.ID)
		log.Printf("creating Post{\n\tID: %d, \n\tURL: %s, \n\ttitle:%s, \n\tsitestr: %s, \n\tpoints: %d}\n",
			psm.post.ID, psm.post.URL, psm.post.Title, psm.post.SiteStr, psm.post.Points)
		// TODO: handle states

	case stateScore:
		psm.state = stateInit
	case stateHnuserIncoming:
		psm.state = stateInit
	case stateHnuser:
		psm.state = stateInit
	case stateAgeIncoming1:
		psm.state = stateInit
	case stateAgeIncoming2:
		psm.state = stateInit
	case stateAge:
		psm.state = stateInit
	case stateNCommentsIncoming:
		psm.state = stateInit
	case stateNComments:
		psm.state = stateInit

	}

	return nil
}
