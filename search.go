package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// searchResult is one hit returned by the search endpoint.
type searchResult struct {
	Rel     string `json:"rel"`
	URL     string `json:"url"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

// searchCatalog scans every document for query and returns ranked matches.
func searchCatalog(cat *catalog, query string, limit int) []searchResult {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}

	type scored struct {
		result searchResult
		score  int
	}
	var hits []scored

	for _, d := range cat.docs {
		src, err := os.ReadFile(filepath.Join(cat.root, filepath.FromSlash(d.Rel)))
		if err != nil {
			continue
		}
		text := string(src)
		lower := strings.ToLower(text)

		score := 0
		idx := strings.Index(lower, q)
		if idx >= 0 {
			score += 10
		}
		if strings.Contains(strings.ToLower(d.Title), q) {
			score += 5
		}
		score += strings.Count(lower, q)
		if score == 0 {
			continue
		}

		hits = append(hits, scored{
			result: searchResult{
				Rel:     d.Rel,
				URL:     d.URL,
				Title:   d.Title,
				Snippet: snippet(text, idx),
			},
			score: score,
		})
	}

	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		return strings.ToLower(hits[i].result.Title) < strings.ToLower(hits[j].result.Title)
	})

	out := make([]searchResult, 0, len(hits))
	for _, h := range hits {
		out = append(out, h.result)
		if limit > 0 && len(out) == limit {
			break
		}
	}
	return out
}

// snippet returns a short plain-text window around idx.
func snippet(text string, idx int) string {
	const width = 120
	runes := []rune(strings.ReplaceAll(text, "\n", " "))
	if len(runes) <= width {
		return strings.TrimSpace(string(runes))
	}
	start := 0
	if idx > 0 {
		start = len([]rune(text[:idx])) - width/3
	}
	if start < 0 {
		start = 0
	}
	end := start + width
	if end > len(runes) {
		end = len(runes)
		start = end - width
	}
	s := strings.TrimSpace(string(runes[start:end]))
	if start > 0 {
		s = "..." + s
	}
	if end < len(runes) {
		s += "..."
	}
	return s
}
