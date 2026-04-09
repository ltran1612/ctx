package fuzzy

import "testing"

// ── Find ──────────────────────────────────────────────────────────────────

func TestFind_BasicMatch(t *testing.T) {
	items := []string{"Fix auth bug", "Refactor payment", "Update docs"}
	matches := Find("auth", items)
	if len(matches) == 0 {
		t.Fatal("expected at least one match")
	}
	if matches[0].Index != 0 {
		t.Errorf("expected index 0, got %d", matches[0].Index)
	}
}

func TestFind_NoMatch(t *testing.T) {
	items := []string{"Fix auth bug", "Refactor payment"}
	matches := Find("zzzzz", items)
	if len(matches) != 0 {
		t.Errorf("expected no matches, got %d", len(matches))
	}
}

func TestFind_EmptyItems(t *testing.T) {
	matches := Find("auth", []string{})
	if len(matches) != 0 {
		t.Errorf("expected no matches on empty list, got %d", len(matches))
	}
}

func TestFind_SortedByScoreDescending(t *testing.T) {
	// "auth" should score higher in "auth bug" than in "authenticate payment flow"
	items := []string{"authenticate payment flow refactor", "auth bug"}
	matches := Find("auth", items)
	if len(matches) < 2 {
		t.Skip("not enough matches to check ordering")
	}
	if matches[0].Score < matches[1].Score {
		t.Errorf("results not sorted descending: scores %d, %d", matches[0].Score, matches[1].Score)
	}
}

func TestFind_MultipleMatches(t *testing.T) {
	items := []string{"auth fix", "auth refactor", "payment bug", "auth token refresh"}
	matches := Find("auth", items)
	if len(matches) < 2 {
		t.Errorf("expected multiple matches, got %d", len(matches))
	}
}

func TestFind_IndexIsCorrect(t *testing.T) {
	items := []string{"apple", "banana", "cherry", "date"}
	matches := Find("cherry", items)
	if len(matches) == 0 {
		t.Fatal("expected a match")
	}
	if matches[0].Index != 2 {
		t.Errorf("wrong index: got %d, want 2", matches[0].Index)
	}
}

func TestFind_EmptyPattern(t *testing.T) {
	items := []string{"a", "b", "c"}
	// sahilm/fuzzy returns all items for empty pattern
	matches := Find("", items)
	// just verify no panic and valid indices
	for _, m := range matches {
		if m.Index < 0 || m.Index >= len(items) {
			t.Errorf("invalid index %d", m.Index)
		}
	}
}

// ── FindFullText ──────────────────────────────────────────────────────────

func TestFindFullText_MatchInTitle(t *testing.T) {
	titles := []string{"Fix auth bug", "Refactor payment"}
	bodies := []string{"some body", "other body"}
	matches := FindFullText("auth", titles, bodies)
	if len(matches) == 0 {
		t.Fatal("expected match in title")
	}
	if matches[0].Index != 0 {
		t.Errorf("expected index 0, got %d", matches[0].Index)
	}
}

func TestFindFullText_MatchInBodyOnly(t *testing.T) {
	titles := []string{"Fix something", "Do another thing"}
	bodies := []string{"contains authentication details here", "unrelated content"}
	matches := FindFullText("authentication", titles, bodies)
	if len(matches) == 0 {
		t.Fatal("expected match in body")
	}
	if matches[0].Index != 0 {
		t.Errorf("expected index 0, got %d", matches[0].Index)
	}
}

func TestFindFullText_BodyNewlinesConvertedToSpaces(t *testing.T) {
	titles := []string{"My Topic"}
	bodies := []string{"line one\nspecial keyword\nline three"}
	// should still find the keyword
	matches := FindFullText("keyword", titles, bodies)
	if len(matches) == 0 {
		t.Fatal("expected match in multi-line body")
	}
}

func TestFindFullText_EmptyLists(t *testing.T) {
	matches := FindFullText("anything", []string{}, []string{})
	if len(matches) != 0 {
		t.Errorf("expected no matches on empty lists, got %d", len(matches))
	}
}

func TestFindFullText_NoMatch(t *testing.T) {
	titles := []string{"Fix auth"}
	bodies := []string{"some content"}
	matches := FindFullText("zzzzz", titles, bodies)
	if len(matches) != 0 {
		t.Errorf("expected no matches, got %d", len(matches))
	}
}
