package fuzzy

import (
	"sort"
	"strings"

	gofuzzy "github.com/sahilm/fuzzy"
)

// Match represents a fuzzy search result.
type Match struct {
	Index int
	Score int
}

// Find returns indices and scores for items matching pattern, sorted by score desc.
func Find(pattern string, items []string) []Match {
	results := gofuzzy.Find(pattern, items)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	out := make([]Match, len(results))
	for i, r := range results {
		out[i] = Match{Index: r.Index, Score: r.Score}
	}
	return out
}

// FindFullText searches both title and body content.
// items is a list of "title\nbody" strings; pattern is the query.
func FindFullText(pattern string, titles, bodies []string) []Match {
	combined := make([]string, len(titles))
	for i := range titles {
		combined[i] = titles[i] + " " + strings.ReplaceAll(bodies[i], "\n", " ")
	}
	return Find(pattern, combined)
}
