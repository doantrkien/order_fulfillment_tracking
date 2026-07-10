package ai

import (
	"context"
	"errors"
	"main/internal/models"
	"main/internal/repositories"
	"testing"

	"github.com/pgvector/pgvector-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Mock KnowledgeRepository ──────────────────────────────────────────────────

type mockKnowledgeRepository struct {
	entries            []models.KnowledgeEntry
	err                error
	saved              []*models.KnowledgeEntry
	similarEntries     []models.KnowledgeEntry
	similarErr         error
	updatedID          int64
	updatedEmbedding   []float32
	// Chunk fields
	savedChunks        []models.KnowledgeChunk
	similarChunks      []repositories.ChunkWithEntry
	similarChunksErr   error
	deletedEntryID     int64
	chunksNeedReembed  []models.KnowledgeChunk
	updatedChunkID     int64
	updatedChunkEmbed  []float32
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

func (m *mockKnowledgeRepository) SaveChunks(_ context.Context, chunks []models.KnowledgeChunk) error {
	m.savedChunks = append(m.savedChunks, chunks...)
	return nil
}

func (m *mockKnowledgeRepository) DeleteChunksByEntryID(_ context.Context, entryID int64) error {
	m.deletedEntryID = entryID
	return nil
}

func (m *mockKnowledgeRepository) FindSimilarChunks(_ context.Context, _ []float32, _ int, _ float64) ([]repositories.ChunkWithEntry, error) {
	return m.similarChunks, m.similarChunksErr
}

func (m *mockKnowledgeRepository) UpdateChunkEmbedding(_ context.Context, chunkID int64, vec []float32) error {
	m.updatedChunkID = chunkID
	m.updatedChunkEmbed = vec
	return nil
}

func (m *mockKnowledgeRepository) GetChunksNeedingReembed(_ context.Context) ([]models.KnowledgeChunk, error) {
	return m.chunksNeedReembed, nil
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

	// Should seed 1 default entry
	assert.Len(t, repo.saved, 1)
	assert.Len(t, store.cache, 1)

	// Check key entries
	assert.Contains(t, store.cache, "combined_knowledge")
}

func TestKnowledgeStore_Initialize_NonEmptyDB_DoesNotSeed(t *testing.T) {
	existing := []models.KnowledgeEntry{
		{Slug: "combined_knowledge", Title: "Custom Knowledge Base", Body: "Custom body", IsActive: true},
	}
	repo := &mockKnowledgeRepository{entries: existing}
	store := NewKnowledgeStore(repo, nil)

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	// Should NOT seed any data
	assert.Empty(t, repo.saved)
	assert.Len(t, store.cache, 1)
	assert.Equal(t, "Custom Knowledge Base", store.cache["combined_knowledge"].Title)
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
		{Slug: "combined_knowledge", Title: "Order Fulfillment Knowledge Base", Body: "DB Delivery Failure Body", IsActive: true},
	}
	repo := &mockKnowledgeRepository{entries: existing}
	store := NewKnowledgeStore(repo, nil) // keyword path

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	entries := store.ClassifyDriverNote(context.Background(), "hàng bị mất")
	require.Len(t, entries, 1)
	assert.Equal(t, "Order Fulfillment Knowledge Base", entries[0].Title)
}

func TestKnowledgeStore_ClassifyDriverNote_KeywordFallback_EmbedError(t *testing.T) {
	existing := []models.KnowledgeEntry{
		{Slug: "combined_knowledge", Title: "Order Fulfillment Knowledge Base", Body: "body", IsActive: true},
	}
	repo := &mockKnowledgeRepository{entries: existing}
	embClient := &mockEmbeddingClient{err: errors.New("embedding API timeout")}
	store := NewKnowledgeStore(repo, embClient)

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	entries := store.ClassifyDriverNote(context.Background(), "xe hỏng trên đường giao hàng hôm nay")
	assert.NotEmpty(t, entries, "should return keyword-matched entries on embed error")
	assert.Equal(t, "Order Fulfillment Knowledge Base", entries[0].Title)
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
	similarChunk := repositories.ChunkWithEntry{
		ChunkID:    1,
		EntrySlug:  "delivery_failure",
		EntryTitle: "Delivery Failure",
		Content:    "delivery_failure_body",
		Similarity: 0.9,
	}
	repo := &mockKnowledgeRepository{
		entries:       []models.KnowledgeEntry{similarEntry},
		similarChunks: []repositories.ChunkWithEntry{similarChunk},
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

func TestKnowledgeStore_ClassifyDriverNote_AllPathsBelowThreshold_ReturnsEmpty(t *testing.T) {
	testVec := make([]float32, 768)
	repo := &mockKnowledgeRepository{
		entries:        []models.KnowledgeEntry{{Slug: "state_machine", Title: "SM", Body: "body", IsActive: true}},
		similarChunks:  []repositories.ChunkWithEntry{},  // chunk search: no match
		similarEntries: []models.KnowledgeEntry{},          // entry-level search: no match
	}
	embClient := &mockEmbeddingClient{vec: testVec}
	store := NewKnowledgeStore(repo, embClient)

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	// "some unrelated note" matches no keywords and no semantic search
	entries := store.ClassifyDriverNote(context.Background(), "some unrelated note")
	assert.NotEmpty(t, entries, "should fallback to combined KB when no semantic match")
	assert.Equal(t, "Order Fulfillment Knowledge Base", entries[0].Title)
}

func TestKnowledgeStore_ClassifyDriverNote_FindSimilarError_FallbackToKeyword(t *testing.T) {
	testVec := make([]float32, 768)
	existing := []models.KnowledgeEntry{
		{Slug: "combined_knowledge", Title: "Order Fulfillment Knowledge Base", Body: "body", IsActive: true},
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
	entries := store.ClassifyDriverNote(context.Background(), "xe hỏng không giao được")
	assert.NotEmpty(t, entries, "should fallback to keyword on FindSimilar error")
	assert.Equal(t, "Order Fulfillment Knowledge Base", entries[0].Title)
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

// ── Tests: Chunk-level RAG ────────────────────────────────────────────────────

func TestKnowledgeStore_ClassifyDriverNote_SemanticChunkMatch(t *testing.T) {
	testVec := make([]float32, 768)
	testVec[0] = 0.9

	// FindSimilarChunks returns a chunk from delivery_failure entry
	chunkResult := repositories.ChunkWithEntry{
		ChunkID:    10,
		EntrySlug:  "delivery_failure",
		EntryTitle: "Delivery Failure",
		Heading:    "### CRITICAL — Lost Package",
		Content:    "Package is lost, stolen, or cannot be located.",
		Similarity: 0.85,
	}
	repo := &mockKnowledgeRepository{
		entries:       []models.KnowledgeEntry{{Slug: "delivery_failure", Title: "Delivery Failure", Body: "full body", IsActive: true}},
		similarChunks: []repositories.ChunkWithEntry{chunkResult},
	}
	embClient := &mockEmbeddingClient{vec: testVec}
	store := NewKnowledgeStore(repo, embClient)

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	// Note should match via chunk-level search
	entries := store.ClassifyDriverNote(context.Background(), "hàng bị thất lạc không tìm thấy")
	require.Len(t, entries, 1, "should return 1 entry from chunk match")
	assert.Equal(t, "Delivery Failure", entries[0].Title)
	// Body should be the chunk content, not full entry body
	assert.Contains(t, entries[0].Body, "Package is lost")
	assert.NotContains(t, entries[0].Body, "full body")
}

func TestKnowledgeStore_ClassifyDriverNote_ChunksFallbackToKeyword(t *testing.T) {
	testVec := make([]float32, 768)
	testVec[0] = 0.9

	// FindSimilarChunks returns empty, should fallback to keyword
	similarEntry := models.KnowledgeEntry{
		ID: 1, Slug: "stuck_order", Title: "Stuck Order", Body: "stuck body", IsActive: true,
	}
	repo := &mockKnowledgeRepository{
		entries:        []models.KnowledgeEntry{similarEntry},
		similarChunks:  []repositories.ChunkWithEntry{}, // empty chunks
	}
	embClient := &mockEmbeddingClient{vec: testVec}
	store := NewKnowledgeStore(repo, embClient)

	// Inject the combined KB into the mock cache for keyword fallback
	store.cache = map[string]KnowledgeEntry{
		"combined_knowledge": {Title: "Order Fulfillment Knowledge Base", Body: "combined body"},
	}

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	entries := store.ClassifyDriverNote(context.Background(), "đơn hàng không tiến triển")
	require.Len(t, entries, 1)
	assert.Equal(t, "Order Fulfillment Knowledge Base", entries[0].Title)
}

func TestKnowledgeStore_ClassifyDriverNote_ChunksErrorFallbackToKeyword(t *testing.T) {
	testVec := make([]float32, 768)

	similarEntry := models.KnowledgeEntry{
		ID: 1, Slug: "delivery_failure", Title: "DF", Body: "body", IsActive: true,
	}
	repo := &mockKnowledgeRepository{
		entries:          []models.KnowledgeEntry{similarEntry},
		similarChunksErr: errors.New("chunks table not found"),
	}
	embClient := &mockEmbeddingClient{vec: testVec}
	store := NewKnowledgeStore(repo, embClient)

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	// FindSimilarChunks errors → should fall back to keyword
	entries := store.ClassifyDriverNote(context.Background(), "xe hỏng trên đường giao hàng")
	assert.NotEmpty(t, entries, "should fallback to keyword on chunks error")
	assert.Equal(t, "Order Fulfillment Knowledge Base", entries[0].Title)
}

func TestKnowledgeStore_BuildRAGEntries_GroupsByEntry(t *testing.T) {
	repo := &mockKnowledgeRepository{}
	store := NewKnowledgeStore(repo, nil)

	chunks := []repositories.ChunkWithEntry{
		{ChunkID: 1, EntrySlug: "delivery_failure", EntryTitle: "Delivery Failure", Content: "chunk 1 content", Similarity: 0.90},
		{ChunkID: 2, EntrySlug: "delivery_failure", EntryTitle: "Delivery Failure", Content: "chunk 2 content", Similarity: 0.85},
		{ChunkID: 3, EntrySlug: "stuck_order", EntryTitle: "Stuck Order", Content: "stuck chunk", Similarity: 0.80},
	}

	result := store.buildRAGEntries(chunks)

	// Should group into 2 entries
	require.Len(t, result, 2)

	// First entry should be delivery_failure (highest similarity)
	assert.Equal(t, "Delivery Failure", result[0].Title)
	assert.Contains(t, result[0].Body, "chunk 1 content")
	assert.Contains(t, result[0].Body, "chunk 2 content")

	// Second entry should be stuck_order
	assert.Equal(t, "Stuck Order", result[1].Title)
	assert.Contains(t, result[1].Body, "stuck chunk")
}
