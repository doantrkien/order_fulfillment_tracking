package ai

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChunkMarkdown_SplitsByHeadings(t *testing.T) {
	body := `# Delivery Failure — Severity Rules

Exception Type: DELIVERY_FAILURE

## Severity Classification
### LOW — Temporary Issue
Minor issue that does not block delivery.

### MEDIUM — Customer Issue
Delivery failed due to customer availability.

### HIGH — Operational Disruption
Vehicle breakdown or accident on route.

### CRITICAL — Lost Package
Package is lost, stolen, or cannot be located.
`

	chunks := ChunkMarkdown(body, "Delivery Failure", 10)

	// Should have overview + 4 severity sections (or merged if small)
	require.True(t, len(chunks) >= 2, "expected at least 2 chunks, got %d", len(chunks))

	// First chunk should be overview with the entry title as heading
	assert.Equal(t, "Delivery Failure", chunks[0].Heading)
	assert.Contains(t, chunks[0].Content, "Exception Type: DELIVERY_FAILURE")

	// Later chunks should contain severity headings
	found := false
	for _, c := range chunks {
		if strings.Contains(c.Heading, "Severity") || strings.Contains(c.Heading, "LOW") ||
			strings.Contains(c.Heading, "MEDIUM") || strings.Contains(c.Heading, "HIGH") ||
			strings.Contains(c.Heading, "CRITICAL") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected at least one chunk with severity heading")
}

func TestChunkMarkdown_NoHeadings_SingleChunk(t *testing.T) {
	body := "This is a simple document with no headings. It should be returned as a single chunk."

	chunks := ChunkMarkdown(body, "Test Entry", DefaultMinChunkTokens)

	require.Len(t, chunks, 1)
	assert.Equal(t, "Test Entry", chunks[0].Heading)
	assert.Equal(t, body, chunks[0].Content)
}

func TestChunkMarkdown_MergesSmallChunks(t *testing.T) {
	body := `## Section One
This is section one content with enough words to be a valid chunk on its own.

## Tiny
Hi.

## Section Three
This section three has plenty of content to stand on its own as a separate chunk.
`

	chunks := ChunkMarkdown(body, "Test", 10) // minTokens = 10

	// "Tiny" section has ~1 word → should be merged into Section One
	for _, c := range chunks {
		assert.NotEqual(t, "## Tiny", c.Heading,
			"tiny section should be merged, not standalone")
	}
}

func TestChunkMarkdown_EstimatesTokens(t *testing.T) {
	body := `## Test Section
one two three four five six seven eight nine ten eleven twelve thirteen fourteen fifteen
`

	chunks := ChunkMarkdown(body, "Test", 0)
	require.Len(t, chunks, 1)

	// 17 words * 1.3 ≈ 22 tokens
	assert.True(t, chunks[0].TokenCount >= 15, "token count should be reasonable, got %d", chunks[0].TokenCount)
	assert.True(t, chunks[0].TokenCount <= 30, "token count should be reasonable, got %d", chunks[0].TokenCount)
}

func TestChunkMarkdown_EmptyBody(t *testing.T) {
	chunks := ChunkMarkdown("", "Empty", DefaultMinChunkTokens)
	// Empty body → either 0 or 1 chunk with empty content
	if len(chunks) > 0 {
		assert.Empty(t, strings.TrimSpace(chunks[0].Content))
	}
}

func TestChunkMarkdown_RealKBDocument(t *testing.T) {
	// Simulate a real KB document structure
	body := `# Delivery Failure — Severity Rules

Exception Type: DELIVERY_FAILURE

Applicable When:
Order is in shipped status and a delivery attempt fails due to a driver-reported issue.

## Severity Classification
### LOW — Temporary / Minor Issue
Condition: Minor issue that does not block delivery.
Examples: customer temporarily unreachable, no answer will retry call, short delay at delivery point.
Likely Impact: Delivery can still succeed in the same day.
Internal Next Action: Retry contact customer and reattempt delivery.

### MEDIUM — Customer or Address Issue
Condition: Delivery failed due to customer availability or address-related issues.
Examples: customer not home, no one available, wrong address, address not found.
Likely Impact: Delivery can be completed after rescheduling.
Internal Next Action: Contact customer to reschedule delivery.

### HIGH — Operational / External Disruption
Condition: Delivery cannot be completed due to external or operational obstacles.
Examples: vehicle breakdown, accident on route, bad weather, road blocked.
Likely Impact: Delivery delayed but still recoverable.
Internal Next Action: Escalate to logistics team.

### CRITICAL — Lost or Unrecoverable Package
Condition: Package is lost, stolen, or cannot be located.
Examples: package lost, stolen, cannot find package, missing parcel.
Likely Impact: Order cannot be fulfilled without investigation.
Internal Next Action: Escalate to logistics manager.

## Examples
### Example 1 — LOW
Input: Status shipped, Driver note: khach dang ban
Output: DELIVERY_FAILURE, LOW
`

	chunks := ChunkMarkdown(body, "Delivery Failure", DefaultMinChunkTokens)

	// Should produce multiple meaningful chunks
	require.True(t, len(chunks) >= 3, "real KB doc should produce at least 3 chunks, got %d", len(chunks))

	// Each chunk should have non-empty content
	for i, c := range chunks {
		assert.NotEmpty(t, c.Content, "chunk %d should have content", i)
		assert.NotEmpty(t, c.Heading, "chunk %d should have heading", i)
		assert.True(t, c.TokenCount > 0, "chunk %d should have positive token count", i)
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		minExpect int
		maxExpect int
	}{
		{"empty", "", 0, 0},
		{"single word", "hello", 1, 2},
		{"sentence", "the quick brown fox jumps over the lazy dog", 10, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := estimateTokens(tt.text)
			assert.True(t, tokens >= tt.minExpect, "expected >= %d, got %d", tt.minExpect, tokens)
			assert.True(t, tokens <= tt.maxExpect, "expected <= %d, got %d", tt.maxExpect, tokens)
		})
	}
}
