package hnapi

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestItemUnmarshal(t *testing.T) {
	// Sample JSON for a story
	storyJSON := `{
		"by": "dhouston",
		"descendants": 71,
		"id": 8863,
		"kids": [8952, 9224, 8917],
		"score": 111,
		"time": 1175714200,
		"title": "My YC app: Dropbox - Throw away your USB drive",
		"type": "story",
		"url": "http://www.getdropbox.com/u/2/screencast.html"
	}`

	var story Item
	err := json.Unmarshal([]byte(storyJSON), &story)
	if err != nil {
		t.Fatalf("Failed to unmarshal story JSON: %v", err)
	}

	// Verify fields were correctly unmarshalled
	if story.ID != 8863 {
		t.Errorf("Expected ID to be 8863, got %d", story.ID)
	}
	if story.By != "dhouston" {
		t.Errorf("Expected By to be 'dhouston', got %q", story.By)
	}
	if story.Descendants != 71 {
		t.Errorf("Expected Descendants to be 71, got %d", story.Descendants)
	}
	if len(story.Kids) != 3 || story.Kids[0] != 8952 {
		t.Errorf("Expected Kids to be [8952, 9224, 8917], got %v", story.Kids)
	}
	if story.Score != 111 {
		t.Errorf("Expected Score to be 111, got %d", story.Score)
	}
	if story.Time != 1175714200 {
		t.Errorf("Expected Time to be 1175714200, got %d", story.Time)
	}
	if story.Title != "My YC app: Dropbox - Throw away your USB drive" {
		t.Errorf("Expected Title to be 'My YC app: Dropbox - Throw away your USB drive', got %q", story.Title)
	}
	if story.Type != "story" {
		t.Errorf("Expected Type to be 'story', got %q", story.Type)
	}
	if story.URL != "http://www.getdropbox.com/u/2/screencast.html" {
		t.Errorf("Expected URL to be 'http://www.getdropbox.com/u/2/screencast.html', got %q", story.URL)
	}

	// Sample JSON for a comment
	commentJSON := `{
		"by": "norvig",
		"id": 2921983,
		"kids": [2922097, 2922429, 2924562],
		"parent": 2921506,
		"text": "Aw shucks, guys ... you make me blush with your compliments.<p>Tell you what, Ill make a deal: I'll keep writing if you keep reading. K?",
		"time": 1314211127,
		"type": "comment"
	}`

	var comment Item
	err = json.Unmarshal([]byte(commentJSON), &comment)
	if err != nil {
		t.Fatalf("Failed to unmarshal comment JSON: %v", err)
	}

	// Verify fields were correctly unmarshalled
	if comment.ID != 2921983 {
		t.Errorf("Expected ID to be 2921983, got %d", comment.ID)
	}
	if comment.By != "norvig" {
		t.Errorf("Expected By to be 'norvig', got %q", comment.By)
	}
	if comment.Parent != 2921506 {
		t.Errorf("Expected Parent to be 2921506, got %d", comment.Parent)
	}
	if comment.Text != "Aw shucks, guys ... you make me blush with your compliments.<p>Tell you what, Ill make a deal: I'll keep writing if you keep reading. K?" {
		t.Errorf("Expected Text to be correctly unmarshalled, got %q", comment.Text)
	}
	if comment.Type != "comment" {
		t.Errorf("Expected Type to be 'comment', got %q", comment.Type)
	}
}

func TestUserUnmarshal(t *testing.T) {
	// Sample JSON for a user
	userJSON := `{
		"about": "This is a test",
		"created": 1173923446,
		"id": "jl",
		"karma": 2937,
		"submitted": [8265435, 8168423, 8090946]
	}`

	var user User
	err := json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		t.Fatalf("Failed to unmarshal user JSON: %v", err)
	}

	// Verify fields were correctly unmarshalled
	if user.ID != "jl" {
		t.Errorf("Expected ID to be 'jl', got %q", user.ID)
	}
	if user.About != "This is a test" {
		t.Errorf("Expected About to be 'This is a test', got %q", user.About)
	}
	if user.Created != 1173923446 {
		t.Errorf("Expected Created to be 1173923446, got %d", user.Created)
	}
	if user.Karma != 2937 {
		t.Errorf("Expected Karma to be 2937, got %d", user.Karma)
	}

	expectedSubmitted := []int{8265435, 8168423, 8090946}
	if !reflect.DeepEqual(user.Submitted, expectedSubmitted) {
		t.Errorf("Expected Submitted to be %v, got %v", expectedSubmitted, user.Submitted)
	}
}

func TestUpdatesUnmarshal(t *testing.T) {
	// Sample JSON for updates
	updatesJSON := `{
		"items": [8423305, 8420805, 8423379],
		"profiles": ["thefox", "mdda", "plinkplonk"]
	}`

	var updates Updates
	err := json.Unmarshal([]byte(updatesJSON), &updates)
	if err != nil {
		t.Fatalf("Failed to unmarshal updates JSON: %v", err)
	}

	// Verify fields were correctly unmarshalled
	expectedItems := []int{8423305, 8420805, 8423379}
	if !reflect.DeepEqual(updates.Items, expectedItems) {
		t.Errorf("Expected Items to be %v, got %v", expectedItems, updates.Items)
	}

	expectedProfiles := []string{"thefox", "mdda", "plinkplonk"}
	if !reflect.DeepEqual(updates.Profiles, expectedProfiles) {
		t.Errorf("Expected Profiles to be %v, got %v", expectedProfiles, updates.Profiles)
	}
}
