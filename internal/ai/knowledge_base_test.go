package ai

import (
	"context"
	"main/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockKnowledgeRepository struct {
	entries []models.KnowledgeEntry
	err     error
	saved   []*models.KnowledgeEntry
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

func TestKnowledgeStore_Initialize_EmptyDB_SeedsData(t *testing.T) {
	repo := &mockKnowledgeRepository{}
	store := NewKnowledgeStore(repo)

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	// Should seed 8 default files
	assert.Len(t, repo.saved, 8)
	assert.Len(t, store.cache, 8)

	// Check if key entries exist in cache
	assert.Contains(t, store.cache, "state_machine")
	assert.Contains(t, store.cache, "delivery_failure")
}

func TestKnowledgeStore_Initialize_NonEmptyDB_DoesNotSeed(t *testing.T) {
	existing := []models.KnowledgeEntry{
		{Slug: "state_machine", Title: "Custom State Machine", Body: "Custom body", IsActive: true},
		{Slug: "delivery_failure", Title: "Custom Delivery Failure", Body: "Custom failure body", IsActive: true},
	}
	repo := &mockKnowledgeRepository{entries: existing}
	store := NewKnowledgeStore(repo)

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	// Should NOT seed any data
	assert.Empty(t, repo.saved)
	assert.Len(t, store.cache, 2)

	// Cache must match existing
	assert.Equal(t, "Custom State Machine", store.cache["state_machine"].Title)
}

func TestKnowledgeStore_ClassifyDriverNote_FromCache(t *testing.T) {
	existing := []models.KnowledgeEntry{
		{Slug: "delivery_failure", Title: "DB Delivery Failure Title", Body: "DB Delivery Failure Body", IsActive: true},
	}
	repo := &mockKnowledgeRepository{entries: existing}
	store := NewKnowledgeStore(repo)

	err := store.Initialize(context.Background())
	require.NoError(t, err)

	// Note matching delivery failure keywords: "hàng bị mất"
	entries := store.ClassifyDriverNote("hàng bị mất")
	require.Len(t, entries, 1)
	assert.Equal(t, "DB Delivery Failure Title", entries[0].Title)
	assert.Equal(t, "DB Delivery Failure Body", entries[0].Body)
}
