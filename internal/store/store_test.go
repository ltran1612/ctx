package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ── helpers ───────────────────────────────────────────────────────────────

func newStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := Init(dir)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	return s
}

func mustCreate(t *testing.T, s *Store, title string, tags []string, ticket string) *Topic {
	t.Helper()
	topic, err := s.Create(title, tags, ticket)
	if err != nil {
		t.Fatalf("Create(%q): %v", title, err)
	}
	return topic
}

// ── Init ──────────────────────────────────────────────────────────────────

func TestInit_CreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	s, err := Init(dir)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	for _, sub := range []string{"active", "archive"} {
		path := filepath.Join(s.Root, sub)
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			t.Errorf("expected dir %s to exist", path)
		}
	}
}

func TestInit_Idempotent(t *testing.T) {
	dir := t.TempDir()
	if _, err := Init(dir); err != nil {
		t.Fatalf("first Init: %v", err)
	}
	if _, err := Init(dir); err != nil {
		t.Fatalf("second Init: %v", err)
	}
}

func TestInit_RootIsCtxSubdir(t *testing.T) {
	dir := t.TempDir()
	s, err := Init(dir)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(s.Root) != ctxDir {
		t.Errorf("root should be .ctx, got %q", filepath.Base(s.Root))
	}
}

// ── Resolve path-traversal guard ─────────────────────────────────────────

func TestResolve_PathTraversalBlocked(t *testing.T) {
	s := newStore(t)
	mustCreate(t, s, "Legit Topic", nil, "")

	traversals := []string{"../../etc/passwd", "../active", ".."}
	for _, q := range traversals {
		_, _, err := s.Resolve(q)
		if err == nil {
			t.Errorf("Resolve(%q): expected error, got nil", q)
		}
	}
}

// ── Create ────────────────────────────────────────────────────────────────

func TestCreate_BasicFields(t *testing.T) {
	s := newStore(t)
	before := time.Now().UTC().Add(-time.Second)
	topic, err := s.Create("Fix auth bug", []string{"go", "auth"}, "PROJ-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if topic.File.Meta.Title != "Fix auth bug" {
		t.Errorf("Title: got %q", topic.File.Meta.Title)
	}
	if topic.File.Meta.Status != "active" {
		t.Errorf("Status: got %q", topic.File.Meta.Status)
	}
	if topic.File.Meta.ID == "" || len(topic.File.Meta.ID) != 6 {
		t.Errorf("ID: got %q (expected 6 hex chars)", topic.File.Meta.ID)
	}
	if topic.File.Meta.Ticket != "PROJ-1" {
		t.Errorf("Ticket: got %q", topic.File.Meta.Ticket)
	}
	if len(topic.File.Meta.Tags) != 2 {
		t.Errorf("Tags: got %v", topic.File.Meta.Tags)
	}
	if topic.File.Meta.Created.Before(before) {
		t.Error("Created timestamp is in the past")
	}
	if topic.File.Meta.Updated.Before(before) {
		t.Error("Updated timestamp is in the past")
	}
}

func TestCreate_FileExistsOnDisk(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "My topic", nil, "")
	if _, err := os.Stat(topic.Path); err != nil {
		t.Errorf("expected file to exist at %s: %v", topic.Path, err)
	}
}

func TestCreate_SlugDerivedFromTitle(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Fix Auth Bug In Login", nil, "")
	if topic.Slug != "fix-auth-bug-in-login" {
		t.Errorf("Slug: got %q, want %q", topic.Slug, "fix-auth-bug-in-login")
	}
}

func TestCreate_SlugCollision(t *testing.T) {
	s := newStore(t)
	t1 := mustCreate(t, s, "Fix Auth Bug", nil, "")
	t2 := mustCreate(t, s, "Fix Auth Bug", nil, "")
	if t1.Slug == t2.Slug {
		t.Error("expected different slugs for duplicate titles")
	}
	if !strings.HasSuffix(t2.Slug, "-2") {
		t.Errorf("second slug should end with -2, got %q", t2.Slug)
	}
}

func TestCreate_IDIsHex(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Test", nil, "")
	for _, ch := range topic.File.Meta.ID {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			t.Errorf("non-hex char %q in ID %q", ch, topic.File.Meta.ID)
		}
	}
}

func TestCreate_EmptyTags(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Test", nil, "")
	if topic.File.Meta.Tags == nil {
		// nil is ok, but should not panic
	}
}

func TestCreate_PlacedInActiveDir(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Test", nil, "")
	if !strings.Contains(topic.Path, filepath.Join(s.Root, activeDir)) {
		t.Errorf("topic should be in active/, path: %s", topic.Path)
	}
}

// ── All ───────────────────────────────────────────────────────────────────

func TestAll_ActiveOnly(t *testing.T) {
	s := newStore(t)
	mustCreate(t, s, "A", nil, "")
	mustCreate(t, s, "B", nil, "")
	topics, err := s.All(true, false)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(topics) != 2 {
		t.Errorf("expected 2 active topics, got %d", len(topics))
	}
}

func TestAll_ArchiveOnly(t *testing.T) {
	s := newStore(t)
	t1 := mustCreate(t, s, "A", nil, "")
	mustCreate(t, s, "B", nil, "")
	s.Archive(t1)
	topics, err := s.All(false, true)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(topics) != 1 {
		t.Errorf("expected 1 archived topic, got %d", len(topics))
	}
}

func TestAll_Both(t *testing.T) {
	s := newStore(t)
	t1 := mustCreate(t, s, "A", nil, "")
	mustCreate(t, s, "B", nil, "")
	s.Archive(t1)
	topics, err := s.All(true, true)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(topics) != 2 {
		t.Errorf("expected 2 total topics, got %d", len(topics))
	}
}

func TestAll_BothFalse(t *testing.T) {
	s := newStore(t)
	mustCreate(t, s, "A", nil, "")
	topics, err := s.All(false, false)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(topics) != 0 {
		t.Errorf("expected 0 topics, got %d", len(topics))
	}
}

func TestAll_EmptyStore(t *testing.T) {
	s := newStore(t)
	topics, err := s.All(true, true)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(topics) != 0 {
		t.Errorf("expected 0 topics, got %d", len(topics))
	}
}

// ── Resolve ───────────────────────────────────────────────────────────────

func TestResolve_ByExactSlug(t *testing.T) {
	s := newStore(t)
	created := mustCreate(t, s, "Fix Auth Bug", nil, "")
	found, archived, err := s.Resolve(created.Slug)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if archived {
		t.Error("should not be archived")
	}
	if found.Slug != created.Slug {
		t.Errorf("slug mismatch: got %q", found.Slug)
	}
}

func TestResolve_ByID(t *testing.T) {
	s := newStore(t)
	created := mustCreate(t, s, "Fix Auth Bug", nil, "")
	found, archived, err := s.Resolve(created.File.Meta.ID)
	if err != nil {
		t.Fatalf("Resolve by ID: %v", err)
	}
	if archived {
		t.Error("should not be archived")
	}
	if found.File.Meta.ID != created.File.Meta.ID {
		t.Errorf("ID mismatch")
	}
}

func TestResolve_ByArchivedSlug(t *testing.T) {
	s := newStore(t)
	t1 := mustCreate(t, s, "Fix Auth Bug", nil, "")
	slug := t1.Slug
	s.Archive(t1)
	found, archived, err := s.Resolve(slug)
	if err != nil {
		t.Fatalf("Resolve archived: %v", err)
	}
	if !archived {
		t.Error("expected archived=true")
	}
	_ = found
}

func TestResolve_ByFuzzyTitle(t *testing.T) {
	s := newStore(t)
	mustCreate(t, s, "Fix Auth Bug In Login Flow", nil, "")
	found, _, err := s.Resolve("auth login")
	if err != nil {
		t.Fatalf("Resolve fuzzy: %v", err)
	}
	if !strings.Contains(found.File.Meta.Title, "Auth") {
		t.Errorf("unexpected topic: %q", found.File.Meta.Title)
	}
}

func TestResolve_ActiveBeforeArchive(t *testing.T) {
	s := newStore(t)
	// Create, archive, then create fresh with same title (different slug)
	t1 := mustCreate(t, s, "Auth Topic", nil, "")
	s.Archive(t1)
	t2 := mustCreate(t, s, "Auth Topic Again", nil, "") // different title
	_ = t2
	// query by archived slug should still find it in archive
	_, archived, err := s.Resolve(t1.Slug)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if !archived {
		t.Error("expected archived")
	}
}

func TestResolve_NotFound(t *testing.T) {
	s := newStore(t)
	_, _, err := s.Resolve("nonexistent-topic")
	if err == nil {
		t.Error("expected error for missing topic")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should say 'not found', got: %v", err)
	}
}

// ── Save ──────────────────────────────────────────────────────────────────

func TestSave_UpdatesTimestamp(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Test", nil, "")
	original := topic.File.Meta.Updated
	time.Sleep(10 * time.Millisecond)
	if err := s.Save(topic); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if !topic.File.Meta.Updated.After(original) {
		t.Error("Updated should be after original")
	}
}

func TestSave_SetsStatusFromLocation(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Test", nil, "")
	// corrupt the status
	topic.File.Meta.Status = "archived"
	if err := s.Save(topic); err != nil {
		t.Fatalf("Save: %v", err)
	}
	// reload and check
	s.Reload(topic)
	if topic.File.Meta.Status != "active" {
		t.Errorf("status should be corrected to active, got %q", topic.File.Meta.Status)
	}
}

func TestSave_PersistsToDisk(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Test", nil, "")
	topic.File.Meta.Title = "Updated Title"
	if err := s.Save(topic); err != nil {
		t.Fatalf("Save: %v", err)
	}
	// read fresh
	reloaded, _, err := s.Resolve(topic.Slug)
	if err != nil {
		t.Fatalf("Resolve after save: %v", err)
	}
	if reloaded.File.Meta.Title != "Updated Title" {
		t.Errorf("Title not persisted: got %q", reloaded.File.Meta.Title)
	}
}

// ── Reload ────────────────────────────────────────────────────────────────

func TestReload_ReadsFromDisk(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Test", nil, "")
	// modify file on disk directly
	data, _ := os.ReadFile(topic.Path)
	modified := strings.Replace(string(data), "Test", "Modified", 1)
	os.WriteFile(topic.Path, []byte(modified), 0644)
	// reload
	if err := s.Reload(topic); err != nil {
		t.Fatalf("Reload: %v", err)
	}
	if topic.File.Meta.Title != "Modified" {
		t.Errorf("expected 'Modified', got %q", topic.File.Meta.Title)
	}
}

func TestReload_FileNotExist(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Test", nil, "")
	os.Remove(topic.Path)
	err := s.Reload(topic)
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// ── Archive ───────────────────────────────────────────────────────────────

func TestArchive_MovesToArchiveDir(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Test", nil, "")
	if err := s.Archive(topic); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	if !strings.Contains(topic.Path, filepath.Join(s.Root, archiveDir)) {
		t.Errorf("expected path in archive/, got %s", topic.Path)
	}
}

func TestArchive_SetsStatusArchived(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Test", nil, "")
	s.Archive(topic)
	if topic.File.Meta.Status != "archived" {
		t.Errorf("expected archived, got %q", topic.File.Meta.Status)
	}
}

func TestArchive_FileExistsInArchive(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Test", nil, "")
	s.Archive(topic)
	if _, err := os.Stat(topic.Path); err != nil {
		t.Errorf("archived file should exist: %v", err)
	}
}

func TestArchive_RemovedFromActive(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Test", nil, "")
	origSlug := topic.Slug
	s.Archive(topic)
	activeDir := filepath.Join(s.Root, "active", origSlug)
	if _, err := os.Stat(activeDir); !os.IsNotExist(err) {
		t.Error("original active dir should no longer exist")
	}
}

func TestArchive_SlugCollision(t *testing.T) {
	s := newStore(t)
	t1 := mustCreate(t, s, "Collision Topic", nil, "")
	originalSlug := t1.Slug
	// Manually create a dir in archive with the same slug to force collision
	conflictDir := filepath.Join(s.Root, archiveDir, originalSlug)
	os.MkdirAll(conflictDir, 0755)
	if err := s.Archive(t1); err != nil {
		t.Fatalf("Archive with collision: %v", err)
	}
	// slug should have been renamed
	if t1.Slug == originalSlug {
		t.Errorf("expected slug to change on collision, still %q", t1.Slug)
	}
}

// ── Restore ───────────────────────────────────────────────────────────────

func TestRestore_MovesToActiveDir(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Test", nil, "")
	s.Archive(topic)
	if err := s.Restore(topic); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if !strings.Contains(topic.Path, filepath.Join(s.Root, activeDir)) {
		t.Errorf("expected path in active/, got %s", topic.Path)
	}
}

func TestRestore_SetsStatusActive(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Test", nil, "")
	s.Archive(topic)
	s.Restore(topic)
	if topic.File.Meta.Status != "active" {
		t.Errorf("expected active, got %q", topic.File.Meta.Status)
	}
}

func TestRestore_FileExistsInActive(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Test", nil, "")
	s.Archive(topic)
	s.Restore(topic)
	if _, err := os.Stat(topic.Path); err != nil {
		t.Errorf("restored file should exist: %v", err)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────

func TestDelete_RemovesDirectory(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Delete Me", nil, "")
	dir := filepath.Dir(topic.Path)
	if err := s.Delete(topic); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("directory should be gone after delete")
	}
}

func TestDelete_TopicNotInListAfterDelete(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Delete Me", nil, "")
	s.Delete(topic)
	topics, _ := s.All(true, false)
	for _, t2 := range topics {
		if t2.File.Meta.ID == topic.File.Meta.ID {
			t.Error("deleted topic should not appear in All()")
		}
	}
}

// ── Search ────────────────────────────────────────────────────────────────

func TestSearch_ByTitle(t *testing.T) {
	s := newStore(t)
	mustCreate(t, s, "Fix Auth Bug", nil, "")
	mustCreate(t, s, "Refactor Payment Service", nil, "")
	results, err := s.Search("auth", false, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one result")
	}
	if !strings.Contains(results[0].Topic.File.Meta.Title, "Auth") {
		t.Errorf("unexpected top result: %q", results[0].Topic.File.Meta.Title)
	}
}

func TestSearch_LimitRespected(t *testing.T) {
	s := newStore(t)
	for i := 0; i < 5; i++ {
		mustCreate(t, s, "auth related topic", nil, "")
	}
	results, err := s.Search("auth", false, 2)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) > 2 {
		t.Errorf("expected at most 2 results, got %d", len(results))
	}
}

func TestSearch_EmptyStore(t *testing.T) {
	s := newStore(t)
	results, err := s.Search("anything", false, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no results on empty store, got %d", len(results))
	}
}

func TestSearch_NoMatch(t *testing.T) {
	s := newStore(t)
	mustCreate(t, s, "Fix Auth Bug", nil, "")
	results, err := s.Search("zzzzz", false, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_FullText(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Unrelated Title", nil, "")
	// Manually write body content to the file
	data, _ := os.ReadFile(topic.Path)
	updated := strings.Replace(string(data), "## Notes\n\n<!-- Scratchpad", "## Notes\n\ncontains uniquekeyword here\n<!-- Scratchpad", 1)
	os.WriteFile(topic.Path, []byte(updated), 0644)
	results, err := s.Search("uniquekeyword", true, 10)
	if err != nil {
		t.Fatalf("Search full-text: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected full-text match in body")
	}
}

func TestSearch_SortedByScore(t *testing.T) {
	s := newStore(t)
	mustCreate(t, s, "auth", nil, "")
	mustCreate(t, s, "auth bug fix in login authentication service", nil, "")
	results, err := s.Search("auth", false, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) < 2 {
		t.Skip("not enough results")
	}
	if results[0].Score < results[1].Score {
		t.Error("results should be sorted by score descending")
	}
}

func TestSearch_ArchivedTopicsIncluded(t *testing.T) {
	s := newStore(t)
	topic := mustCreate(t, s, "Auth Bug Archived", nil, "")
	s.Archive(topic)
	results, err := s.Search("auth", false, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	found := false
	for _, r := range results {
		if r.Topic.File.Meta.ID == topic.File.Meta.ID {
			found = true
			if !r.Archived {
				t.Error("archived flag should be true for archived topic")
			}
		}
	}
	if !found {
		t.Error("archived topic should appear in search results")
	}
}

// ── generateID ────────────────────────────────────────────────────────────

func TestGenerateID_Length(t *testing.T) {
	id, err := generateID()
	if err != nil {
		t.Fatalf("generateID: %v", err)
	}
	if len(id) != 6 {
		t.Errorf("expected 6 chars, got %d: %q", len(id), id)
	}
}

func TestGenerateID_IsHex(t *testing.T) {
	id, err := generateID()
	if err != nil {
		t.Fatalf("generateID: %v", err)
	}
	for _, ch := range id {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			t.Errorf("non-hex char %q in %q", ch, id)
		}
	}
}

func TestGenerateID_Unique(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		id, err := generateID()
		if err != nil {
			t.Fatalf("generateID: %v", err)
		}
		if seen[id] {
			t.Errorf("duplicate ID generated: %q", id)
		}
		seen[id] = true
	}
}
