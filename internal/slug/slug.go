package slug

import (
	"fmt"
	"regexp"
	"strings"
)

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// FromTitle derives a slug from a title: lowercase, kebab-case, max 5 words.
func FromTitle(title string) string {
	lower := strings.ToLower(title)
	words := strings.Fields(lower)
	if len(words) > 5 {
		words = words[:5]
	}
	joined := strings.Join(words, "-")
	slug := nonAlphanumeric.ReplaceAllString(joined, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

// Unique returns a slug that doesn't collide with existing slugs.
// It appends -2, -3, etc. until unique.
func Unique(base string, exists func(string) bool) string {
	if !exists(base) {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !exists(candidate) {
			return candidate
		}
	}
}
