package hnapi_test

import (
	"testing"

	"github.com/yarlson/hnapi"
)

func TestHelloHackerNews(t *testing.T) {
	expected := "Hello, Hacker News API"
	actual := hnapi.HelloHackerNews()

	if actual != expected {
		t.Errorf("HelloHackerNews() = %q, want %q", actual, expected)
	}
}
