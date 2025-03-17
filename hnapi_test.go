package hnapi

import (
	"testing"
)

func TestHelloHackerNews(t *testing.T) {
	expected := "Hello, Hacker News API"
	actual := HelloHackerNews()

	if actual != expected {
		t.Errorf("HelloHackerNews() = %q, want %q", actual, expected)
	}
}
