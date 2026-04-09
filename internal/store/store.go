package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/ctx/internal/frontmatter"
	"github.com/user/ctx/internal/fuzzy"
	"github.com/user/ctx/internal/slug"
)

// isSafePathComponent rejects slugs/IDs that could escape the store directory.
func isSafePathComponent(s string) bool {
	if s == "" || s == "." || s == ".." {
		return false
	}
	return !strings.ContainsAny(s, `/\`)
}

const (
	ctxDir      = ".ctx"
	activeDir   = "active"
	archiveDir  = "archive"
	contextFile = "context.md"
)

// Store manages the .ctx directory.
type Store struct {
	Root string // absolute path to .ctx/
}

// Topic is a loaded topic with its parsed content.
type Topic struct {
	Slug string
	Path string // absolute path to context.md
	File *frontmatter.File
}

// Open returns the store rooted at ~/.ctx/, creating it if it does not exist.
func Open() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}
	return Init(home)
}

// Init creates the .ctx directory structure in dir. Idempotent.
// Exported so tests can create isolated stores in temp directories.
func Init(dir string) (*Store, error) {
	root := filepath.Join(dir, ctxDir)
	for _, sub := range []string{activeDir, archiveDir} {
		if err := os.MkdirAll(filepath.Join(root, sub), 0755); err != nil {
			return nil, err
		}
	}
	return &Store{Root: root}, nil
}

func (s *Store) activeRoot() string  { return filepath.Join(s.Root, activeDir) }
func (s *Store) archiveRoot() string { return filepath.Join(s.Root, archiveDir) }

func (s *Store) slugExists(sub, sl string) bool {
	_, err := os.Stat(filepath.Join(s.Root, sub, sl))
	return err == nil
}

// Create writes a new topic file and returns the Topic.
func (s *Store) Create(title string, tags []string, ticket string) (*Topic, error) {
	base := slug.FromTitle(title)
	sl := slug.Unique(base, func(candidate string) bool {
		return s.slugExists(activeDir, candidate)
	})

	id, err := generateID()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	meta := frontmatter.Meta{
		ID:      id,
		Slug:    sl,
		Title:   title,
		Status:  "active",
		Created: now,
		Updated: now,
		Tags:    tags,
		Ticket:  ticket,
	}

	content, err := frontmatter.Template(meta)
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(s.activeRoot(), sl)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, contextFile)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, err
	}

	f, err := frontmatter.Parse(content)
	if err != nil {
		return nil, err
	}
	return &Topic{Slug: sl, Path: path, File: f}, nil
}

// All returns topics from the requested status directories.
func (s *Store) All(includeActive, includeArchive bool) ([]*Topic, error) {
	var topics []*Topic
	if includeActive {
		ts, err := s.listDir(activeDir)
		if err != nil {
			return nil, err
		}
		topics = append(topics, ts...)
	}
	if includeArchive {
		ts, err := s.listDir(archiveDir)
		if err != nil {
			return nil, err
		}
		topics = append(topics, ts...)
	}
	return topics, nil
}

func (s *Store) listDir(sub string) ([]*Topic, error) {
	dir := filepath.Join(s.Root, sub)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var topics []*Topic
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(dir, e.Name(), contextFile)
		t, err := loadTopic(e.Name(), path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not load %s: %v\n", path, err)
			continue
		}
		topics = append(topics, t)
	}
	return topics, nil
}

// Resolve finds a topic by slug, id, or fuzzy title match.
// Returns the matched topic and whether it's from the archive.
func (s *Store) Resolve(query string) (*Topic, bool, error) {
	// 1. exact slug in active
	if t := s.loadBySlug(activeDir, query); t != nil {
		return t, false, nil
	}
	// 2. short id
	if t, archived, err := s.findByID(query); err == nil && t != nil {
		return t, archived, nil
	}
	// 3. exact slug in archive
	if t := s.loadBySlug(archiveDir, query); t != nil {
		return t, true, nil
	}
	// 4. fuzzy title match
	all, err := s.All(true, true)
	if err != nil {
		return nil, false, err
	}
	t, archived := fuzzyResolve(query, all)
	if t != nil {
		return t, archived, nil
	}
	return nil, false, fmt.Errorf("topic %q not found", query)
}

// Search performs fuzzy search over topic titles (and optionally bodies).
type SearchResult struct {
	Topic    *Topic
	Score    int
	Archived bool
}

func (s *Store) Search(query string, fullText bool, limit int) ([]SearchResult, error) {
	all, err := s.All(true, true)
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, nil
	}

	titles := make([]string, len(all))
	bodies := make([]string, len(all))
	for i, t := range all {
		titles[i] = t.File.Meta.Title
		bodies[i] = t.File.Body
	}

	var matches []fuzzy.Match
	if fullText {
		matches = fuzzy.FindFullText(query, titles, bodies)
	} else {
		matches = fuzzy.Find(query, titles)
	}

	var results []SearchResult
	for _, m := range matches {
		t := all[m.Index]
		results = append(results, SearchResult{
			Topic:    t,
			Score:    m.Score,
			Archived: t.File.Meta.Status == "archived",
		})
		if limit > 0 && len(results) >= limit {
			break
		}
	}
	return results, nil
}

func (s *Store) loadBySlug(sub, sl string) *Topic {
	if !isSafePathComponent(sl) {
		return nil
	}
	path := filepath.Join(s.Root, sub, sl, contextFile)
	t, err := loadTopic(sl, path)
	if err != nil {
		return nil
	}
	return t
}

func (s *Store) findByID(id string) (*Topic, bool, error) {
	for _, sub := range []string{activeDir, archiveDir} {
		topics, err := s.listDir(sub)
		if err != nil {
			return nil, false, err
		}
		for _, t := range topics {
			if t.File.Meta.ID == id {
				return t, sub == archiveDir, nil
			}
		}
	}
	return nil, false, nil
}

// Save writes a modified topic back to disk, updating the Updated timestamp.
func (s *Store) Save(t *Topic) error {
	t.File.Meta.Updated = time.Now().UTC()
	if strings.HasPrefix(t.Path, s.archiveRoot()) {
		t.File.Meta.Status = "archived"
	} else {
		t.File.Meta.Status = "active"
	}
	content, err := frontmatter.Serialize(t.File)
	if err != nil {
		return err
	}
	return os.WriteFile(t.Path, []byte(content), 0644)
}

// Archive moves a topic from active/ to archive/.
func (s *Store) Archive(t *Topic) error {
	return s.moveTopic(t, activeDir, archiveDir, "archived")
}

// Restore moves a topic from archive/ to active/.
func (s *Store) Restore(t *Topic) error {
	return s.moveTopic(t, archiveDir, activeDir, "active")
}

func (s *Store) moveTopic(t *Topic, fromSub, toSub, newStatus string) error {
	destSlug := t.Slug
	destDir := filepath.Join(s.Root, toSub, destSlug)
	if _, err := os.Stat(destDir); err == nil {
		destSlug = t.Slug + "-" + time.Now().Format("20060102150405")
		destDir = filepath.Join(s.Root, toSub, destSlug)
	}
	srcDir := filepath.Dir(t.Path)
	if err := os.Rename(srcDir, destDir); err != nil {
		return err
	}
	t.Slug = destSlug
	t.Path = filepath.Join(destDir, contextFile)
	t.File.Meta.Status = newStatus
	t.File.Meta.Slug = destSlug
	return s.Save(t)
}

// Reload re-reads a topic's file from disk into the Topic struct.
func (s *Store) Reload(t *Topic) error {
	data, err := os.ReadFile(t.Path)
	if err != nil {
		return err
	}
	f, err := frontmatter.Parse(string(data))
	if err != nil {
		return err
	}
	t.File = f
	return nil
}

// Delete removes a topic directory permanently.
func (s *Store) Delete(t *Topic) error {
	return os.RemoveAll(filepath.Dir(t.Path))
}

func loadTopic(sl, path string) (*Topic, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f, err := frontmatter.Parse(string(data))
	if err != nil {
		return nil, err
	}
	if f.Meta.Slug == "" {
		f.Meta.Slug = sl
	}
	return &Topic{Slug: sl, Path: path, File: f}, nil
}

func generateID() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func fuzzyResolve(query string, all []*Topic) (*Topic, bool) {
	titles := make([]string, len(all))
	for i, t := range all {
		titles[i] = t.File.Meta.Title
	}
	matches := fuzzy.Find(query, titles)
	if len(matches) == 0 {
		return nil, false
	}
	t := all[matches[0].Index]
	return t, t.File.Meta.Status == "archived"
}
