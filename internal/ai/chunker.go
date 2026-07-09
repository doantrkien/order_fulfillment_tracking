package ai

import (
	"regexp"
	"strings"
)

// DefaultMinChunkTokens is the minimum estimated token count for a chunk.
// Chunks smaller than this are merged into the previous chunk to avoid
// excessively short fragments that produce low-quality embeddings.
const DefaultMinChunkTokens = 50

// Chunk represents a single section extracted from a KB markdown document.
type Chunk struct {
	Heading    string // section heading, e.g. "### CRITICAL — Lost Package"
	Content    string // full text of the section (heading + body)
	TokenCount int    // estimated token count
}

// headingRe matches markdown headings at level 2 or 3 (## or ###).
var headingRe = regexp.MustCompile(`(?m)^(#{2,3}\s+.*)$`)

// ChunkMarkdown splits a KB markdown document into sections by ## and ### headings.
// Each chunk contains the heading plus its body text until the next heading.
//
// Chunks smaller than minTokens are merged into the previous chunk to prevent
// fragments that are too short for meaningful embeddings.
//
// The introductory text before the first heading (if any) is emitted as an
// "overview" chunk whose heading is set to entryTitle.
func ChunkMarkdown(body string, entryTitle string, minTokens int) []Chunk {
	if minTokens <= 0 {
		minTokens = DefaultMinChunkTokens
	}

	body = strings.ReplaceAll(body, "\r\n", "\n")

	// Find all heading positions
	matches := headingRe.FindAllStringIndex(body, -1)

	if len(matches) == 0 {
		// No headings found — return the entire body as a single chunk
		return []Chunk{{
			Heading:    entryTitle,
			Content:    strings.TrimSpace(body),
			TokenCount: estimateTokens(body),
		}}
	}

	var rawChunks []Chunk

	// Handle text before the first heading (overview section)
	if matches[0][0] > 0 {
		overview := strings.TrimSpace(body[:matches[0][0]])
		if overview != "" {
			rawChunks = append(rawChunks, Chunk{
				Heading:    entryTitle,
				Content:    overview,
				TokenCount: estimateTokens(overview),
			})
		}
	}

	// Extract each heading + body section
	for i, loc := range matches {
		start := loc[0]
		var end int
		if i+1 < len(matches) {
			end = matches[i+1][0]
		} else {
			end = len(body)
		}

		section := strings.TrimSpace(body[start:end])
		heading := strings.TrimSpace(body[loc[0]:loc[1]])

		rawChunks = append(rawChunks, Chunk{
			Heading:    heading,
			Content:    section,
			TokenCount: estimateTokens(section),
		})
	}

	// Merge small chunks into the previous chunk
	if len(rawChunks) <= 1 {
		return rawChunks
	}

	merged := []Chunk{rawChunks[0]}
	for i := 1; i < len(rawChunks); i++ {
		if rawChunks[i].TokenCount < minTokens {
			// Merge into previous: append content, keep previous heading
			prev := &merged[len(merged)-1]
			prev.Content = prev.Content + "\n\n" + rawChunks[i].Content
			prev.TokenCount = estimateTokens(prev.Content)
		} else {
			merged = append(merged, rawChunks[i])
		}
	}

	return merged
}

// estimateTokens provides a rough token count estimate.
// For mixed English/Vietnamese text, ~1.3 tokens per whitespace-separated word
// is a reasonable approximation.
func estimateTokens(text string) int {
	words := len(strings.Fields(text))
	estimate := int(float64(words) * 1.3)
	if estimate < 1 && len(strings.TrimSpace(text)) > 0 {
		return 1
	}
	return estimate
}
