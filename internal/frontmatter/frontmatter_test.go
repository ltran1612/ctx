package frontmatter

import (
	"strings"
	"testing"
	"time"
)

// ── Parse ──────────────────────────────────────────────────────────────────

func TestParse_ValidFrontmatter(t *testing.T) {
	content := `---
id: abc123
slug: my-topic
title: My Topic
status: active
created: 2026-01-01T00:00:00Z
updated: 2026-01-02T00:00:00Z
tags:
  - go
  - cli
ticket: PROJ-1
---

## Summary

Hello world.
`
	f, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Meta.ID != "abc123" {
		t.Errorf("ID: got %q, want %q", f.Meta.ID, "abc123")
	}
	if f.Meta.Title != "My Topic" {
		t.Errorf("Title: got %q, want %q", f.Meta.Title, "My Topic")
	}
	if f.Meta.Status != "active" {
		t.Errorf("Status: got %q, want %q", f.Meta.Status, "active")
	}
	if f.Meta.Ticket != "PROJ-1" {
		t.Errorf("Ticket: got %q, want %q", f.Meta.Ticket, "PROJ-1")
	}
	if len(f.Meta.Tags) != 2 || f.Meta.Tags[0] != "go" || f.Meta.Tags[1] != "cli" {
		t.Errorf("Tags: got %v, want [go cli]", f.Meta.Tags)
	}
	if !strings.Contains(f.Body, "Hello world.") {
		t.Errorf("Body missing expected content, got: %q", f.Body)
	}
}

func TestParse_NoFrontmatter(t *testing.T) {
	content := "just a plain body\nno frontmatter"
	f, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Body != content {
		t.Errorf("Body: got %q, want %q", f.Body, content)
	}
	if f.Meta.ID != "" {
		t.Errorf("expected empty meta, got ID=%q", f.Meta.ID)
	}
}

func TestParse_StripsBOM(t *testing.T) {
	content := "\xef\xbb\xbf---\nid: bom1\ntitle: BOM Test\nstatus: active\nslug: bom-test\ncreated: 2026-01-01T00:00:00Z\nupdated: 2026-01-01T00:00:00Z\ntags: []\n---\n\nbody"
	f, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Meta.ID != "bom1" {
		t.Errorf("BOM not stripped; ID: got %q", f.Meta.ID)
	}
}

func TestParse_MissingClosingSeparator(t *testing.T) {
	content := "---\ntitle: broken\n"
	_, err := Parse(content)
	if err == nil {
		t.Error("expected error for missing closing ---, got nil")
	}
}

func TestParse_MalformedYAML(t *testing.T) {
	content := "---\ntitle: [unclosed\n---\nbody"
	_, err := Parse(content)
	if err == nil {
		t.Error("expected error for malformed YAML, got nil")
	}
}

func TestParse_EmptyContent(t *testing.T) {
	f, err := Parse("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Body != "" {
		t.Errorf("expected empty body, got %q", f.Body)
	}
}

func TestParse_BodyAfterFrontmatterContainsContent(t *testing.T) {
	content := "---\ntitle: T\nstatus: active\nslug: t\nid: x\ncreated: 2026-01-01T00:00:00Z\nupdated: 2026-01-01T00:00:00Z\ntags: []\n---\n\nbody content"
	f, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(f.Body, "body content") {
		t.Errorf("expected body to contain 'body content', got %q", f.Body)
	}
}

func TestParse_SpecialCharsInTitle(t *testing.T) {
	content := "---\ntitle: \"It's a test: 100%\"\nstatus: active\nslug: its-a-test\nid: x\ncreated: 2026-01-01T00:00:00Z\nupdated: 2026-01-01T00:00:00Z\ntags: []\n---\n"
	f, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Meta.Title != "It's a test: 100%" {
		t.Errorf("Title: got %q", f.Meta.Title)
	}
}

// ── Serialize ─────────────────────────────────────────────────────────────

func TestSerialize_RoundTrip(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	original := &File{
		Meta: Meta{
			ID:      "abc123",
			Slug:    "my-topic",
			Title:   "My Topic",
			Status:  "active",
			Created: now,
			Updated: now,
			Tags:    []string{"go", "cli"},
			Ticket:  "PROJ-1",
		},
		Body: "## Summary\n\nHello.\n",
	}
	out, err := Serialize(original)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	parsed, err := Parse(out)
	if err != nil {
		t.Fatalf("Parse after Serialize: %v", err)
	}
	if parsed.Meta.ID != original.Meta.ID {
		t.Errorf("ID mismatch: got %q want %q", parsed.Meta.ID, original.Meta.ID)
	}
	if parsed.Meta.Title != original.Meta.Title {
		t.Errorf("Title mismatch: got %q want %q", parsed.Meta.Title, original.Meta.Title)
	}
	if len(parsed.Meta.Tags) != 2 {
		t.Errorf("Tags: got %v", parsed.Meta.Tags)
	}
	if !strings.Contains(parsed.Body, "Hello.") {
		t.Errorf("Body not preserved: %q", parsed.Body)
	}
}

func TestSerialize_EmptyBody(t *testing.T) {
	now := time.Now()
	f := &File{Meta: Meta{ID: "x", Slug: "s", Title: "T", Status: "active", Created: now, Updated: now}, Body: ""}
	out, err := Serialize(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(out, "---\n") {
		t.Error("output should start with ---")
	}
	// No body section after the closing ---
	parts := strings.SplitN(out, "---\n", 3)
	if len(parts) == 3 && strings.TrimSpace(parts[2]) != "" {
		t.Errorf("expected empty body section, got %q", parts[2])
	}
}

func TestSerialize_StartsAndEndsSeparators(t *testing.T) {
	now := time.Now()
	f := &File{Meta: Meta{ID: "x", Slug: "s", Title: "T", Status: "active", Created: now, Updated: now}, Body: "body"}
	out, err := Serialize(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(out, "---\n") {
		t.Error("should start with ---")
	}
	if !strings.Contains(out, "\n---\n") {
		t.Error("should contain closing ---")
	}
}

// ── Template ──────────────────────────────────────────────────────────────

func TestTemplate_ContainsAllSections(t *testing.T) {
	now := time.Now()
	meta := Meta{ID: "x", Slug: "s", Title: "T", Status: "active", Created: now, Updated: now}
	out, err := Template(meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, section := range []string{"## Summary", "## Goal", "## Context", "## Current State", "## Next Steps", "## Notes"} {
		if !strings.Contains(out, section) {
			t.Errorf("template missing section %q", section)
		}
	}
}

func TestTemplate_HasFrontmatter(t *testing.T) {
	now := time.Now()
	meta := Meta{ID: "abc", Slug: "my-slug", Title: "My Title", Status: "active", Created: now, Updated: now}
	out, err := Template(meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "id: abc") {
		t.Error("template missing id in frontmatter")
	}
	if !strings.Contains(out, "My Title") {
		t.Error("template missing title in frontmatter")
	}
}

// ── Section ───────────────────────────────────────────────────────────────

func TestSection_Found(t *testing.T) {
	body := "## Summary\n\nThis is the summary.\n\n## Notes\n\nSome notes.\n"
	content, ok := Section(body, "Summary")
	if !ok {
		t.Fatal("expected section to be found")
	}
	if content != "This is the summary." {
		t.Errorf("got %q", content)
	}
}

func TestSection_NotFound(t *testing.T) {
	body := "## Summary\n\nHello.\n"
	_, ok := Section(body, "Notes")
	if ok {
		t.Error("expected not found")
	}
}

func TestSection_CaseInsensitive(t *testing.T) {
	body := "## Next Steps\n\n- step 1\n"
	content, ok := Section(body, "next steps")
	if !ok {
		t.Fatal("expected section to be found case-insensitively")
	}
	if !strings.Contains(content, "step 1") {
		t.Errorf("got %q", content)
	}
}

func TestSection_LastSection(t *testing.T) {
	body := "## Summary\n\nfirst.\n\n## Notes\n\nlast section content.\n"
	content, ok := Section(body, "Notes")
	if !ok {
		t.Fatal("expected section to be found")
	}
	if !strings.Contains(content, "last section content") {
		t.Errorf("got %q", content)
	}
}

func TestSection_EmptyBody(t *testing.T) {
	_, ok := Section("", "Summary")
	if ok {
		t.Error("expected not found on empty body")
	}
}

func TestSection_EmptySectionContent(t *testing.T) {
	body := "## Summary\n\n## Notes\n\ncontent\n"
	content, ok := Section(body, "Summary")
	if !ok {
		t.Fatal("expected section to exist even if empty")
	}
	if content != "" {
		t.Errorf("expected empty content, got %q", content)
	}
}

func TestSection_H1NotMatched(t *testing.T) {
	body := "# Summary\n\ncontent\n"
	_, ok := Section(body, "Summary")
	if ok {
		t.Error("H1 headings should not be matched as sections")
	}
}

// ── AppendToSection ───────────────────────────────────────────────────────

func TestAppendToSection_SectionExists(t *testing.T) {
	body := "## Notes\n\nexisting note\n"
	result := AppendToSection(body, "Notes", "new note")
	if !strings.Contains(result, "existing note") {
		t.Error("existing content should be preserved")
	}
	if !strings.Contains(result, "new note") {
		t.Error("new text should be appended")
	}
	// new note should come after existing note
	idxExisting := strings.Index(result, "existing note")
	idxNew := strings.Index(result, "new note")
	if idxExisting > idxNew {
		t.Error("new note should appear after existing note")
	}
}

func TestAppendToSection_SectionNotExists(t *testing.T) {
	body := "## Summary\n\ncontent\n"
	result := AppendToSection(body, "Notes", "new note")
	if !strings.Contains(result, "## Notes") {
		t.Error("missing Notes section should be created")
	}
	if !strings.Contains(result, "new note") {
		t.Error("new text should be present")
	}
}

func TestAppendToSection_EmptyBody(t *testing.T) {
	result := AppendToSection("", "Notes", "text")
	if !strings.Contains(result, "## Notes") {
		t.Error("section should be created in empty body")
	}
	if !strings.Contains(result, "text") {
		t.Error("text should be present")
	}
}

func TestAppendToSection_SectionFollowedByAnother(t *testing.T) {
	body := "## Notes\n\nnote1\n\n## Summary\n\nsummary\n"
	result := AppendToSection(body, "Notes", "note2")
	if !strings.Contains(result, "note2") {
		t.Error("note2 should be appended")
	}
	if !strings.Contains(result, "## Summary") {
		t.Error("Summary section should still be present")
	}
	idxNote2 := strings.Index(result, "note2")
	idxSummary := strings.Index(result, "## Summary")
	if idxNote2 > idxSummary {
		t.Error("appended note should appear before the Summary section")
	}
}

// ── PrependToSection ──────────────────────────────────────────────────────

func TestPrependToSection_SectionExists(t *testing.T) {
	body := "## Notes\n\nexisting\n"
	result := PrependToSection(body, "Notes", "new observation")
	if !strings.Contains(result, "new observation") {
		t.Error("new note should be present")
	}
	if !strings.Contains(result, "existing") {
		t.Error("existing content should be preserved")
	}
	idxNew := strings.Index(result, "new observation")
	idxExisting := strings.Index(result, "existing")
	if idxNew > idxExisting {
		t.Error("prepended note should appear before existing content")
	}
}

func TestPrependToSection_TimestampFormat(t *testing.T) {
	body := "## Notes\n\n"
	result := PrependToSection(body, "Notes", "text")
	// should contain a timestamp in the form YYYY-MM-DD HH:MM
	if !strings.Contains(result, time.Now().Format("2006-01-02")) {
		t.Error("prepended note should contain today's date")
	}
	if !strings.Contains(result, "**") {
		t.Error("timestamp should be bolded with **")
	}
}

func TestPrependToSection_SectionNotExists(t *testing.T) {
	body := "## Summary\n\ncontent\n"
	result := PrependToSection(body, "Notes", "obs")
	if !strings.Contains(result, "## Notes") {
		t.Error("Notes section should be created")
	}
	if !strings.Contains(result, "obs") {
		t.Error("text should be present")
	}
}

func TestPrependToSection_EmptyBody(t *testing.T) {
	result := PrependToSection("", "Notes", "note")
	if !strings.Contains(result, "## Notes") {
		t.Error("section should be created")
	}
	if !strings.Contains(result, "note") {
		t.Error("note should be present")
	}
}
