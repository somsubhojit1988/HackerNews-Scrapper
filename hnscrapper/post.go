package hnscrapper

import "fmt"

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
