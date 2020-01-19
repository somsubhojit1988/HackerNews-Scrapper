# HackerNews-Scrapper
Scraps hacker news to collect relevant informations about each published story 
(eg: Title, Origin site, Url, Points etc..).

## Overview
The __hnscrapper__ package implements the parsing logic, through a state machine.
Each story on __HN__ are a collection of some Html table-rows `<tr>`.

Informations pertaining to each story are identifiable by standard Html tag types 
associated with specific `class` and/or `id` attributes.

For example the origin site for a story that is shared from an external site is 
contained inside a `span` tag, `<span class="sitestr">example.com</span>`.

The __hnscrapper__ uses this structure and semantics to identify the various informations about each post, and thus aggregates them in a `Post struct`. 

***
## State machine
Each story is parsed by transitioning through a series of states.
### stateInit
The beginning state for parsing each story
### stateID              
After identifying the beggining of a new story and extracting it's Numeric ID
### stateStoryLink
After extracting the article's original URL.
### stateStoryTitle 
After the Title text of the story is parsed.
### stateSitestrIncoming 
States suffixed with _incoming_ are intermediary states that has been used to skip irrelevant tags
### stateSiteStr         
After extracting the source site of the story
### stateScoreIncoming   
An intermerdiary state
### stateScore 
After extracting the points for the story.   
### stateAge
State after extracting the age of the post usually in the form `$n [minutes|hours| seconds] ago`      
### stateNComments
After Extracting the number of comments at that instant for that story, present in the form `$n comments`.

***