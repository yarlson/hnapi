package hnapi

// Item represents a Hacker News item, which can be a story, comment, job, poll, or pollopt.
type Item struct {
	// ID is the unique identifier for this item.
	ID int `json:"id"`

	// Deleted indicates if the item is deleted.
	Deleted bool `json:"deleted,omitempty"`

	// Type is the type of the item: "job", "story", "comment", "poll", or "pollopt".
	Type string `json:"type"`

	// By is the username of the item's author.
	By string `json:"by,omitempty"`

	// Time is when the item was created, in Unix seconds.
	Time int64 `json:"time"`

	// Text is the comment, story, or poll text in HTML.
	Text string `json:"text,omitempty"`

	// Dead indicates if the item is dead.
	Dead bool `json:"dead,omitempty"`

	// Parent is the comment's parent (another comment or story).
	Parent int `json:"parent,omitempty"`

	// Poll is the pollopt's associated poll.
	Poll int `json:"poll,omitempty"`

	// Kids is the IDs of the item's comments, in ranked display order.
	Kids []int `json:"kids,omitempty"`

	// URL is the URL of the story.
	URL string `json:"url,omitempty"`

	// Score is the story's score.
	Score int `json:"score,omitempty"`

	// Title is the title of the story, poll, or job.
	Title string `json:"title,omitempty"`

	// Parts are the related pollopts, in display order.
	Parts []int `json:"parts,omitempty"`

	// Descendants is the total comment count.
	Descendants int `json:"descendants,omitempty"`
}

// User represents a Hacker News user.
type User struct {
	// ID is the user's unique username.
	ID string `json:"id"`

	// Created is when the user was created, in Unix seconds.
	Created int64 `json:"created"`

	// Karma is the user's karma.
	Karma int `json:"karma"`

	// About is the user's self-description in HTML.
	About string `json:"about,omitempty"`

	// Submitted are the IDs of the user's stories, polls, and comments.
	Submitted []int `json:"submitted,omitempty"`
}

// Updates represents the changes from the /v0/updates endpoint.
type Updates struct {
	// Items are the IDs of changed or new items.
	Items []int `json:"items"`

	// Profiles are the IDs of changed user profiles.
	Profiles []string `json:"profiles"`
}
