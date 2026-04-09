package cmd

import (
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/user/ctx/internal/frontmatter"
	"github.com/user/ctx/internal/store"
)

// ── containsTag ───────────────────────────────────────────────────────────

func TestContainsTag_Found(t *testing.T) {
	if !containsTag([]string{"go", "auth", "p1"}, "auth") {
		t.Error("expected true")
	}
}

func TestContainsTag_NotFound(t *testing.T) {
	if containsTag([]string{"go", "auth"}, "backend") {
		t.Error("expected false")
	}
}

func TestContainsTag_EmptyList(t *testing.T) {
	if containsTag([]string{}, "auth") {
		t.Error("expected false on empty list")
	}
}

func TestContainsTag_EmptyTag(t *testing.T) {
	if containsTag([]string{"a", "b"}, "") {
		t.Error("expected false for empty tag")
	}
}

// ── removeTags ────────────────────────────────────────────────────────────

func TestRemoveTags_RemovesTag(t *testing.T) {
	result := removeTags([]string{"go", "auth", "p1"}, "auth")
	for _, tag := range result {
		if tag == "auth" {
			t.Error("auth should have been removed")
		}
	}
	if len(result) != 2 {
		t.Errorf("expected 2 remaining tags, got %d", len(result))
	}
}

func TestRemoveTags_TagNotPresent(t *testing.T) {
	result := removeTags([]string{"go", "auth"}, "backend")
	if len(result) != 2 {
		t.Errorf("expected unchanged list, got %v", result)
	}
}

func TestRemoveTags_RemovesAllInstances(t *testing.T) {
	result := removeTags([]string{"auth", "go", "auth"}, "auth")
	for _, tag := range result {
		if tag == "auth" {
			t.Error("all instances of auth should be removed")
		}
	}
}

func TestRemoveTags_EmptyList(t *testing.T) {
	result := removeTags([]string{}, "auth")
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

// ── filterByTags ──────────────────────────────────────────────────────────

func makeTopic(title string, tags []string) *store.Topic {
	return &store.Topic{
		Slug: strings.ToLower(strings.ReplaceAll(title, " ", "-")),
		File: &frontmatter.File{
			Meta: frontmatter.Meta{Title: title, Tags: tags},
		},
	}
}

func TestFilterByTags_SingleTagMatch(t *testing.T) {
	topics := []*store.Topic{
		makeTopic("A", []string{"auth", "go"}),
		makeTopic("B", []string{"payment"}),
	}
	result := filterByTags(topics, []string{"auth"})
	if len(result) != 1 || result[0].File.Meta.Title != "A" {
		t.Errorf("expected only A, got %v", result)
	}
}

func TestFilterByTags_MultipleTagsANDLogic(t *testing.T) {
	topics := []*store.Topic{
		makeTopic("A", []string{"auth", "go"}),
		makeTopic("B", []string{"auth"}),
	}
	result := filterByTags(topics, []string{"auth", "go"})
	if len(result) != 1 || result[0].File.Meta.Title != "A" {
		t.Errorf("expected only A (has both tags), got %v", result)
	}
}

func TestFilterByTags_EmptyFilter(t *testing.T) {
	topics := []*store.Topic{
		makeTopic("A", []string{"auth"}),
		makeTopic("B", []string{"go"}),
	}
	result := filterByTags(topics, []string{})
	if len(result) != 2 {
		t.Errorf("empty filter should return all, got %d", len(result))
	}
}

func TestFilterByTags_NoMatch(t *testing.T) {
	topics := []*store.Topic{
		makeTopic("A", []string{"auth"}),
	}
	result := filterByTags(topics, []string{"backend"})
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}

func TestFilterByTags_TopicWithNoTags(t *testing.T) {
	topics := []*store.Topic{
		makeTopic("A", nil),
		makeTopic("B", []string{"auth"}),
	}
	result := filterByTags(topics, []string{"auth"})
	if len(result) != 1 || result[0].File.Meta.Title != "B" {
		t.Errorf("expected only B, got %v", result)
	}
}

// ── hasAllTags ────────────────────────────────────────────────────────────

func TestHasAllTags_AllPresent(t *testing.T) {
	if !hasAllTags([]string{"a", "b", "c"}, []string{"a", "c"}) {
		t.Error("expected true")
	}
}

func TestHasAllTags_Missing(t *testing.T) {
	if hasAllTags([]string{"a", "b"}, []string{"a", "c"}) {
		t.Error("expected false")
	}
}

func TestHasAllTags_EmptyWant(t *testing.T) {
	if !hasAllTags([]string{"a"}, []string{}) {
		t.Error("empty want should return true")
	}
}

func TestHasAllTags_EmptyHave(t *testing.T) {
	if hasAllTags([]string{}, []string{"a"}) {
		t.Error("expected false when have is empty but want is not")
	}
}

func TestHasAllTags_BothEmpty(t *testing.T) {
	if !hasAllTags([]string{}, []string{}) {
		t.Error("both empty should return true")
	}
}

// ── sortTopics ────────────────────────────────────────────────────────────

func makeTopicWithTimes(title string, created, updated time.Time) *store.Topic {
	return &store.Topic{
		Slug: strings.ToLower(strings.ReplaceAll(title, " ", "-")),
		File: &frontmatter.File{
			Meta: frontmatter.Meta{Title: title, Created: created, Updated: updated},
		},
	}
}

func TestSortTopics_ByUpdatedDesc(t *testing.T) {
	t1 := makeTopicWithTimes("A", time.Now().Add(-2*time.Hour), time.Now().Add(-2*time.Hour))
	t2 := makeTopicWithTimes("B", time.Now().Add(-1*time.Hour), time.Now().Add(-1*time.Hour))
	topics := []*store.Topic{t1, t2}
	sortTopics(topics, "updated")
	if topics[0].File.Meta.Title != "B" {
		t.Error("most recently updated should be first")
	}
}

func TestSortTopics_ByCreatedDesc(t *testing.T) {
	t1 := makeTopicWithTimes("A", time.Now().Add(-3*time.Hour), time.Now())
	t2 := makeTopicWithTimes("B", time.Now().Add(-1*time.Hour), time.Now().Add(-2*time.Hour))
	topics := []*store.Topic{t1, t2}
	sortTopics(topics, "created")
	if topics[0].File.Meta.Title != "B" {
		t.Error("most recently created should be first")
	}
}

func TestSortTopics_ByTitleAsc(t *testing.T) {
	t1 := makeTopicWithTimes("Zebra", time.Now(), time.Now())
	t2 := makeTopicWithTimes("Apple", time.Now(), time.Now())
	t3 := makeTopicWithTimes("Mango", time.Now(), time.Now())
	topics := []*store.Topic{t1, t2, t3}
	sortTopics(topics, "title")
	titles := make([]string, len(topics))
	for i, t := range topics {
		titles[i] = t.File.Meta.Title
	}
	if !sort.StringsAreSorted(titles) {
		t.Errorf("expected alphabetical order, got %v", titles)
	}
}

func TestSortTopics_UnknownKeyDefaultsToUpdated(t *testing.T) {
	t1 := makeTopicWithTimes("A", time.Now(), time.Now().Add(-1*time.Hour))
	t2 := makeTopicWithTimes("B", time.Now(), time.Now())
	topics := []*store.Topic{t1, t2}
	sortTopics(topics, "boguskey")
	if topics[0].File.Meta.Title != "B" {
		t.Error("unknown sort key should default to updated desc")
	}
}
