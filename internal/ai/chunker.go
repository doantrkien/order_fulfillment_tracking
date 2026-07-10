package ai

import (
	"regexp"
	"strings"
)

const DefaultMinChunkTokens = 50

type Chunk struct {
	Heading    string
	Content    string
	TokenCount int
}

var headingRe = regexp.MustCompile(`(?m)^(#{2,3}\s+.*)$`)

func ChunkMarkdown(body string, entryTitle string, minTokens int) []Chunk {
	if minTokens <= 0 {
		minTokens = DefaultMinChunkTokens
	}

	body = strings.ReplaceAll(body, "\r\n", "\n")

	matches := headingRe.FindAllStringIndex(body, -1)

	if len(matches) == 0 {
		return []Chunk{{
			Heading:    entryTitle,
			Content:    strings.TrimSpace(body),
			TokenCount: estimateTokens(body),
		}}
	}

	var rawChunks []Chunk

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

	if len(rawChunks) <= 1 {
		return rawChunks
	}

	merged := []Chunk{rawChunks[0]}
	for i := 1; i < len(rawChunks); i++ {
		if rawChunks[i].TokenCount < minTokens {
			prev := &merged[len(merged)-1]
			prev.Content = prev.Content + "\n\n" + rawChunks[i].Content
			prev.TokenCount = estimateTokens(prev.Content)
		} else {
			merged = append(merged, rawChunks[i])
		}
	}

	return merged
}

func estimateTokens(text string) int {
	words := len(strings.Fields(text))
	estimate := int(float64(words) * 1.3)
	if estimate < 1 && len(strings.TrimSpace(text)) > 0 {
		return 1
	}
	return estimate
}
