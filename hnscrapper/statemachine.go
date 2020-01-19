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
	stateHnuserIncoming1   State = 7
	stateHnuserIncoming2   State = 8
	stateHnuser            State = 9
	stateAgeIncoming1      State = 10
	stateAgeIncoming2      State = 11
	stateAge               State = 12
	stateNCommentsIncoming State = 13
	stateNComments         State = 14
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
	case stateScoreIncoming:
		return "score-incoming"
	case stateScore:
		return "state-score"
	case stateHnuserIncoming1:
		return "state-hnuser-incoming[1]"
	case stateHnuserIncoming2:
		return "state-hnuser-incoming[2]"
	case stateHnuser:
		return "state-hnuser"
	case stateAgeIncoming1:
		return "state-age-incoming[1]"
	case stateAgeIncoming2:
		return "state-age-incoming[2]"
	case stateAge:
		return "state-age"
	case stateNCommentsIncoming:
		return "state-comments-incoming"
	case stateNComments:
		return "state-comments"
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

func (psm *PostParsingSM) postUser(attrs map[string]string) error {
	user, ok := attrs[data]
	if !ok {
		return &SMError{"could not parse user from TextToken"}
	}
	psm.post.User = user
	return nil
}

func (psm *PostParsingSM) isValidTag(attrs map[string]string) bool {
	nxtState := psm.state + 1
	ret := false
	switch nxtState {
	case stateHnuserIncoming1:
		if txt, ok := attrs[data]; ok {
			ret = strings.TrimSpace(txt) == "by"
		}

	case stateHnuserIncoming2:
		// <a class = "hnuser" href = "user?id=$userid">
		if tType, ok := attrs[tagType]; !ok || tType != anchorTag {
			break
		}
		cls, hasCls := attrs[class]
		url, hasURL := attrs[href]
		ret = hasCls && cls == hnuser && hasURL &&
			strings.Index(strings.TrimSpace(url), "user") == 0
	}

	return ret
}

func (psm *PostParsingSM) handleTransitState(attrs map[string]string) error {
	var err error
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

	case stateScore:
		if psm.isValidTag(attrs) {
			log.Printf("[state-transition] stateScore -> stateHnuserIncoming1 [postID: %d]", psm.post.ID)
			psm.state = stateHnuserIncoming1
		}
	case stateHnuserIncoming1:
		log.Printf("[stateHnuserIncoming1] attributes: %v", attrs)
		if isValid := psm.isValidTag(attrs); isValid {
			log.Printf("[state-transition] stateHnuserIncoming1 -> stateHnuserIncoming2 [postID: %d]", psm.post.ID)
			psm.state = stateHnuserIncoming2
		} else {
			psm.state = stateInit // TODO: Remove
		}
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
			log.Printf("[ERROR state-transition] Falling back stateSitestrIncoming -> stateInit [postID: %d]", psm.post.ID)
			psm.state = stateInit
			return err
		}
		log.Printf("[state-transition] stateSitestrIncoming -> stateSiteStr [postID: %d]", psm.post.ID)
		psm.state = stateSiteStr

	case stateSiteStr:
		return psm.handleTransitState(attrs)

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

	case stateScore:
		return psm.handleTransitState(attrs)

	case stateHnuserIncoming1:
		return psm.handleTransitState(attrs)

	case stateHnuserIncoming2:
		err := psm.postUser(attrs)
		if err != nil {
			log.Fatal("error parsing user error: ", err)
			return err
		}
		psm.state = stateHnuser
		log.Printf("[state-transition] stateHnuser -> stateAgeIncoming1 [postID: %d]", psm.post.ID)
		log.Printf("creating Post{\n\tID: %d, \n\tURL: %s, \n\ttitle:%s, \n\tsitestr: %s, \n\tpoints: %d, \n\tUser:%s }\n",
			psm.post.ID, psm.post.URL, psm.post.Title, psm.post.SiteStr, psm.post.Points, psm.post.User)

		// TODO: handle states
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
