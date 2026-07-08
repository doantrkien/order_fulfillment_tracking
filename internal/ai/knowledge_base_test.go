package ai

import (
	"context"
	"errors"
	"main/internal/models"
	"testing"

	"github.com/pgvector/pgvector-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Mock KnowledgeRepository ──────────────────────────────────────────────────

type mockKnowledgeRepository struct {
	entries         []models.KnowledgeEntry
	err             error
	saved           []*models.KnowledgeEntry
	similarEntries  []models.KnowledgeEntry
	similarErr      error
	updatedID       int64
	updatedEmbedding []float32
}

func (m *mockKnowledgeRepository) GetAllActive(_ context.Context) ([]models.KnowledgeEntry, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.entries, nil
}

func (m *mockKnowledgeRepository) GetBySlug(_ context.Context, slug string) (*models.KnowledgeEntry, error) {
	for _, e := range m.entries {
		if e.Slug == slug && e.IsActive {
			return &e, nil
		}
	}
	return nil, nil
}

func (m *mockKnowledgeRepository) Save(_ context.Context, entry *models.KnowledgeEntry) error {
	m.saved = append(m.saved, entry)
	return nil
}

func (m *mockKnowledgeRepository) FindSimilar(_ context.Context, _ []float32, _ int, _ float64) ([]models.KnowledgeEntry, error) {
	return m.similarEntries, m.similarErr
}

func (m *mockKnowledgeRepository) UpdateEmbedding(_ context.Context, id int64, vec []float32) error {
	m.updatedID = id
	m.updatedEmbedding = vec
	return nil
}

// ── Mock EmbeddingClient ──────────────────────────────────────────────────────

type mockEmbeddingClient struct {
	vec []float32
	err error
}

func (m *mockEmbeddingClient) Embed(_ context.Context, _ string) ([]float32, error) {
	return m.vec, m.err
}

// ── Tests: Initialize ─────────────────────────────────────────────────────────

func TestKnowledgeStore_Initialize_EmptyDB_SeedsData(t *testing.T) {
	repo := &mockKnowledgeRepository{}
	store := NewKnowledgeStore(repo, nil) // no embedding client

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	// Should seed 8 default entries
	assert.Len(t, repo.saved, 8)
	assert.Len(t, store.cache, 8)

	// Check key entries
	assert.Contains(t, store.cache, "state_machine")
	assert.Contains(t, store.cache, "delivery_failure")
}

func TestKnowledgeStore_Initialize_NonEmptyDB_DoesNotSeed(t *testing.T) {
	existing := []models.KnowledgeEntry{
		{Slug: "state_machine", Title: "Custom State Machine", Body: "Custom body", IsActive: true},
		{Slug: "delivery_failure", Title: "Custom Delivery Failure", Body: "Custom failure body", IsActive: true},
	}
	repo := &mockKnowledgeRepository{entries: existing}
	store := NewKnowledgeStore(repo, nil)

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	// Should NOT seed any data
	assert.Empty(t, repo.saved)
	assert.Len(t, store.cache, 2)
	assert.Equal(t, "Custom State Machine", store.cache["state_machine"].Title)
}

func TestKnowledgeStore_Initialize_GeneratesEmbeddingForNeedsReembed(t *testing.T) {
	testVec := make([]float32, 768)
	testVec[0] = 0.1

	existing := []models.KnowledgeEntry{
		{ID: 42, Slug: "delivery_failure", Title: "DF", Body: "Delivery failed", IsActive: true, NeedsReembed: true},
	}
	repo := &mockKnowledgeRepository{entries: existing}
	embClient := &mockEmbeddingClient{vec: testVec}
	store := NewKnowledgeStore(repo, embClient)

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	// Should have called UpdateEmbedding for the entry with NeedsReembed=true
	assert.Equal(t, int64(42), repo.updatedID)
	assert.Equal(t, testVec, repo.updatedEmbedding)
}

func TestKnowledgeStore_Initialize_SkipsEmbeddingWhenClientNil(t *testing.T) {
	existing := []models.KnowledgeEntry{
		{ID: 1, Slug: "delivery_failure", Title: "DF", Body: "body", IsActive: true, NeedsReembed: true},
	}
	repo := &mockKnowledgeRepository{entries: existing}
	store := NewKnowledgeStore(repo, nil) // no embedding client

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	// Should NOT call UpdateEmbedding
	assert.Equal(t, int64(0), repo.updatedID)
}

// ── Tests: ClassifyDriverNote keyword fallback ────────────────────────────────

func TestKnowledgeStore_ClassifyDriverNote_KeywordFallback_NoEmbeddingClient(t *testing.T) {
	existing := []models.KnowledgeEntry{
		{Slug: "delivery_failure", Title: "DB Delivery Failure Title", Body: "DB Delivery Failure Body", IsActive: true},
	}
	repo := &mockKnowledgeRepository{entries: existing}
	store := NewKnowledgeStore(repo, nil) // keyword path

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	// "hàng bị mất" matches deliveryFailureKeywordsKB
	entries := store.ClassifyDriverNote(context.Background(), "hàng bị mất")
	require.Len(t, entries, 1)
	assert.Equal(t, "DB Delivery Failure Title", entries[0].Title)
	assert.Equal(t, "DB Delivery Failure Body", entries[0].Body)
}

func TestKnowledgeStore_ClassifyDriverNote_KeywordFallback_EmbedError(t *testing.T) {
	existing := []models.KnowledgeEntry{
		{Slug: "delivery_failure", Title: "DF", Body: "body", IsActive: true},
	}
	repo := &mockKnowledgeRepository{entries: existing}
	embClient := &mockEmbeddingClient{err: errors.New("embedding API timeout")}
	store := NewKnowledgeStore(repo, embClient)

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	// embed fails → should fall back to keyword matching, not panic
	entries := store.ClassifyDriverNote(context.Background(), "xe hỏng trên đường giao hàng hôm nay")
	// "xe hỏng" matches deliveryFailureKeywordsKB
	assert.NotEmpty(t, entries, "should return keyword-matched entries on embed error")
	assert.Equal(t, "DF", entries[0].Title)
}

// ── Tests: ClassifyDriverNote semantic path ───────────────────────────────────

func TestKnowledgeStore_ClassifyDriverNote_SemanticMatch(t *testing.T) {
	testVec := make([]float32, 768)
	testVec[0] = 0.9

	// FindSimilar returns a delivery_failure entry
	similarEntry := models.KnowledgeEntry{
		ID:    1,
		Slug:  "delivery_failure",
		Title: "Delivery Failure",
		Body:  "delivery_failure_body",
	}
	repo := &mockKnowledgeRepository{
		entries:        []models.KnowledgeEntry{similarEntry},
		similarEntries: []models.KnowledgeEntry{similarEntry},
	}
	embClient := &mockEmbeddingClient{vec: testVec}
	store := NewKnowledgeStore(repo, embClient)

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	// Note has no keyword but semantic search will return the entry
	entries := store.ClassifyDriverNote(context.Background(), "phương tiện của tôi không vận hành được")
	require.Len(t, entries, 1, "semantic search should return 1 matching entry")
	assert.Equal(t, "Delivery Failure", entries[0].Title)
}

func TestKnowledgeStore_ClassifyDriverNote_SemanticBelowThreshold_ReturnsEmpty(t *testing.T) {
	testVec := make([]float32, 768)
	repo := &mockKnowledgeRepository{
		entries:        []models.KnowledgeEntry{{Slug: "state_machine", Title: "SM", Body: "body", IsActive: true}},
		similarEntries: []models.KnowledgeEntry{}, // FindSimilar returns empty — similarity < threshold
	}
	embClient := &mockEmbeddingClient{vec: testVec}
	store := NewKnowledgeStore(repo, embClient)

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	entries := store.ClassifyDriverNote(context.Background(), "some unrelated note")
	assert.Empty(t, entries, "should return empty when no entry passes similarity threshold")
}

func TestKnowledgeStore_ClassifyDriverNote_FindSimilarError_FallbackToKeyword(t *testing.T) {
	testVec := make([]float32, 768)
	existing := []models.KnowledgeEntry{
		{Slug: "delivery_failure", Title: "DF", Body: "body", IsActive: true},
	}
	repo := &mockKnowledgeRepository{
		entries:    existing,
		similarErr: errors.New("pgvector unavailable"),
	}
	embClient := &mockEmbeddingClient{vec: testVec}
	store := NewKnowledgeStore(repo, embClient)

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	// FindSimilar errors → keyword fallback
	// "xe hỏng" matches deliveryFailureKeywordsKB keyword
	entries := store.ClassifyDriverNote(context.Background(), "xe hỏng không giao được")
	assert.NotEmpty(t, entries, "should fallback to keyword on FindSimilar error")
}

// ── Tests: pgvector.Vector integration ───────────────────────────────────────

func TestPgvectorNewVector_CorrectDimensions(t *testing.T) {
	vec := make([]float32, 768)
	for i := range vec {
		vec[i] = float32(i) * 0.001
	}
	pgVec := pgvector.NewVector(vec)
	assert.Equal(t, 768, len(pgVec.Slice()))
}
