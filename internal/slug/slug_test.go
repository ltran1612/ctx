package slug

import "testing"

// ── FromTitle ─────────────────────────────────────────────────────────────

func TestFromTitle_Basic(t *testing.T) {
	cases := []struct {
		title string
		want  string
	}{
		{"Fix Auth Bug", "fix-auth-bug"},
		{"fix auth bug", "fix-auth-bug"},
		{"FIX AUTH BUG", "fix-auth-bug"},
		{"  leading and trailing  ", "leading-and-trailing"},
		{"one", "one"},
	}
	for _, c := range cases {
		got := FromTitle(c.title)
		if got != c.want {
			t.Errorf("FromTitle(%q) = %q, want %q", c.title, got, c.want)
		}
	}
}

func TestFromTitle_TruncatesAtFiveWords(t *testing.T) {
	title := "one two three four five six seven"
	got := FromTitle(title)
	want := "one-two-three-four-five"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFromTitle_ExactlyFiveWords(t *testing.T) {
	title := "one two three four five"
	got := FromTitle(title)
	want := "one-two-three-four-five"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFromTitle_SpecialCharactersStripped(t *testing.T) {
	got := FromTitle("Fix: auth! bug@login#flow")
	// special chars become separators, then collapsed
	if got == "" {
		t.Error("expected non-empty slug")
	}
	for _, ch := range got {
		if ch != '-' && !(ch >= 'a' && ch <= 'z') && !(ch >= '0' && ch <= '9') {
			t.Errorf("unexpected character %q in slug %q", ch, got)
		}
	}
}

func TestFromTitle_NumbersPreserved(t *testing.T) {
	got := FromTitle("ticket 42 fix")
	if got != "ticket-42-fix" {
		t.Errorf("got %q, want %q", got, "ticket-42-fix")
	}
}

func TestFromTitle_MultipleSpaces(t *testing.T) {
	got := FromTitle("one   two   three")
	if got != "one-two-three" {
		t.Errorf("got %q, want %q", got, "one-two-three")
	}
}

func TestFromTitle_EmptyString(t *testing.T) {
	got := FromTitle("")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFromTitle_OnlySpecialChars(t *testing.T) {
	got := FromTitle("!!! @@@")
	// should produce empty or only dashes (trimmed)
	for _, ch := range got {
		if ch != '-' && !(ch >= 'a' && ch <= 'z') && !(ch >= '0' && ch <= '9') {
			t.Errorf("unexpected char %q in %q", ch, got)
		}
	}
}

func TestFromTitle_NoLeadingOrTrailingDash(t *testing.T) {
	got := FromTitle("!hello!")
	if len(got) > 0 && (got[0] == '-' || got[len(got)-1] == '-') {
		t.Errorf("slug should not have leading/trailing dashes, got %q", got)
	}
}

// ── Unique ────────────────────────────────────────────────────────────────

func TestUnique_NoCollision(t *testing.T) {
	got := Unique("my-slug", func(string) bool { return false })
	if got != "my-slug" {
		t.Errorf("expected base slug, got %q", got)
	}
}

func TestUnique_OneCollision(t *testing.T) {
	existing := map[string]bool{"my-slug": true}
	got := Unique("my-slug", func(s string) bool { return existing[s] })
	if got != "my-slug-2" {
		t.Errorf("got %q, want %q", got, "my-slug-2")
	}
}

func TestUnique_MultipleCollisions(t *testing.T) {
	existing := map[string]bool{"my-slug": true, "my-slug-2": true, "my-slug-3": true}
	got := Unique("my-slug", func(s string) bool { return existing[s] })
	if got != "my-slug-4" {
		t.Errorf("got %q, want %q", got, "my-slug-4")
	}
}

func TestUnique_EmptyBase(t *testing.T) {
	got := Unique("", func(string) bool { return false })
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestUnique_AlwaysFinds(t *testing.T) {
	// exists only for exact base and -2 through -5
	blocked := map[string]bool{"s": true, "s-2": true, "s-3": true, "s-4": true, "s-5": true}
	got := Unique("s", func(s string) bool { return blocked[s] })
	if got != "s-6" {
		t.Errorf("got %q, want %q", got, "s-6")
	}
}
