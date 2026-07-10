package ai

import (
	"context"
	_ "embed"
	"fmt"
	"main/internal/models"
	"main/internal/repositories"
	"main/pkg/embedding"
	"strings"
	"sync"
)

//go:embed knowledge/combined_knowledge.md
var kbBodyCombined string

type KnowledgeEntry struct {
	Title string
	Body  string
}

func GetKnowledgeBase(note string) []KnowledgeEntry {
	fmt.Printf("[DEBUG][LoadKBFromMD] Loading knowledge base from Markdown.\\n")
	return []KnowledgeEntry{
		{
			Title: "Order Fulfillment Knowledge Base",
			Body:  kbBodyCombined,
		},
	}
}

type KnowledgeStore struct {
	Repo            repositories.KnowledgeRepository
	EmbeddingClient embedding.Client
	Cache           map[string]KnowledgeEntry
	mu              sync.RWMutex
}

func NewKnowledgeStore(repo repositories.KnowledgeRepository, embeddingClient embedding.Client) *KnowledgeStore {
	return &KnowledgeStore{
		Repo:            repo,
		EmbeddingClient: embeddingClient,
		Cache:           make(map[string]KnowledgeEntry),
	}
}

func (s *KnowledgeStore) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dbEntries, err := s.Repo.GetAllActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to load knowledge entries from DB: %w", err)
	}

	if len(dbEntries) == 0 {
		seeds := []struct {
			slug  string
			title string
			body  string
		}{
			{"combined_knowledge", "Order Fulfillment Knowledge Base", kbBodyCombined},
		}

		for _, seed := range seeds {
			entry := &models.KnowledgeEntry{
				Slug:         seed.slug,
				Title:        seed.title,
				Body:         seed.body,
				IsActive:     true,
				NeedsReembed: true,
			}
			if err := s.Repo.Save(ctx, entry); err != nil {
				fmt.Printf("[WARNING][KnowledgeStore] Failed to seed slug %q: %v\n", seed.slug, err)
			} else {
				dbEntries = append(dbEntries, *entry)
			}
		}
	}

	s.generateMissingEmbeddings(ctx, dbEntries)

	s.rechunkAndEmbed(ctx, dbEntries)

	s.Cache = make(map[string]KnowledgeEntry)
	for _, dbEntry := range dbEntries {
		s.Cache[dbEntry.Slug] = KnowledgeEntry{
			Title: dbEntry.Title,
			Body:  dbEntry.Body,
		}
	}

	fmt.Printf("[INFO][KnowledgeStore] Loaded %d active knowledge entries into memory cache.\n", len(s.Cache))
	return nil
}

func (s *KnowledgeStore) Reload(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dbEntries, err := s.Repo.GetAllActive(ctx)
	if err != nil {
		return err
	}

	s.generateMissingEmbeddings(ctx, dbEntries)

	s.rechunkAndEmbed(ctx, dbEntries)

	s.Cache = make(map[string]KnowledgeEntry)
	for _, dbEntry := range dbEntries {
		s.Cache[dbEntry.Slug] = KnowledgeEntry{
			Title: dbEntry.Title,
			Body:  dbEntry.Body,
		}
	}

	fmt.Printf("[INFO][KnowledgeStore] Reloaded %d active knowledge entries from DB.\n", len(s.Cache))
	return nil
}

func (s *KnowledgeStore) generateMissingEmbeddings(ctx context.Context, entries []models.KnowledgeEntry) {
	if s.EmbeddingClient == nil {
		return
	}
	for _, entry := range entries {
		if !entry.NeedsReembed {
			continue
		}
		vec, err := s.EmbeddingClient.Embed(ctx, entry.Body)
		if err != nil {
			fmt.Printf("[WARNING][KnowledgeStore] Embed failed for %q: %v\n", entry.Slug, err)
			continue
		}
		if updateErr := s.Repo.UpdateEmbedding(ctx, entry.ID, vec); updateErr != nil {
			fmt.Printf("[WARNING][KnowledgeStore] UpdateEmbedding failed for %q (id=%d): %v\n", entry.Slug, entry.ID, updateErr)
		} else {
			fmt.Printf("[INFO][KnowledgeStore] Embedded entry %q (id=%d, dims=%d)\n", entry.Slug, entry.ID, len(vec))
		}
	}
}

func (s *KnowledgeStore) rechunkAndEmbed(ctx context.Context, entries []models.KnowledgeEntry) {
	if s.EmbeddingClient == nil {
		return
	}

	for _, entry := range entries {
		if !entry.NeedsReembed {
			continue
		}

		if err := s.Repo.DeleteChunksByEntryID(ctx, entry.ID); err != nil {
			fmt.Printf("[WARNING][KnowledgeStore] DeleteChunks failed for %q (id=%d): %v\n", entry.Slug, entry.ID, err)
			continue
		}

		chunks := ChunkMarkdown(entry.Body, entry.Title, DefaultMinChunkTokens)
		if len(chunks) == 0 {
			continue
		}

		dbChunks := make([]models.KnowledgeChunk, 0, len(chunks))
		for i, c := range chunks {
			dbChunks = append(dbChunks, models.KnowledgeChunk{
				EntryID:      entry.ID,
				ChunkIndex:   int16(i),
				Heading:      c.Heading,
				Content:      c.Content,
				TokenCount:   c.TokenCount,
				NeedsReembed: true,
			})
		}

		if err := s.Repo.SaveChunks(ctx, dbChunks); err != nil {
			fmt.Printf("[WARNING][KnowledgeStore] SaveChunks failed for %q: %v\n", entry.Slug, err)
			continue
		}
		fmt.Printf("[INFO][KnowledgeStore] Created %d chunks for entry %q (id=%d)\n", len(dbChunks), entry.Slug, entry.ID)

		for _, dbChunk := range dbChunks {
			vec, err := s.EmbeddingClient.Embed(ctx, dbChunk.Content)
			if err != nil {
				fmt.Printf("[WARNING][KnowledgeStore] Embed chunk failed for %q chunk#%d: %v\n", entry.Slug, dbChunk.ChunkIndex, err)
				continue
			}
			if updateErr := s.Repo.UpdateChunkEmbedding(ctx, dbChunk.ID, vec); updateErr != nil {
				fmt.Printf("[WARNING][KnowledgeStore] UpdateChunkEmbedding failed for chunk id=%d: %v\n", dbChunk.ID, updateErr)
			} else {
				fmt.Printf("[INFO][KnowledgeStore] Embedded chunk %q #%d (id=%d, dims=%d)\n", entry.Slug, dbChunk.ChunkIndex, dbChunk.ID, len(vec))
			}
		}
	}
}

func (s *KnowledgeStore) ClassifyDriverNote(ctx context.Context, note string) []KnowledgeEntry {
	if s.EmbeddingClient != nil {
		return s.classifyBySemantic(ctx, note)
	}

	return s.classifyByKeyword(note)
}

func (s *KnowledgeStore) classifyBySemantic(ctx context.Context, note string) []KnowledgeEntry {
	vec, err := s.EmbeddingClient.Embed(ctx, note)
	if err != nil {
		fmt.Printf("[WARN][KnowledgeStore] Embed error, falling back to keyword: %v\n", err)
		return s.classifyByKeyword(note)
	}

	chunks, err := s.Repo.FindSimilarChunks(ctx, vec, 5, 0.65)
	if err != nil {
		fmt.Printf("[WARN][KnowledgeStore] FindSimilarChunks error, trying entry-level: %v\n", err)
		return s.classifyByEntryLevel(ctx, vec, note)
	}

	if len(chunks) == 0 {
		fmt.Printf("[DEBUG][ClassifyDriverNote] Semantic chunks: no match (cosine < 0.65) for note: %q. Trying entry-level.\n", note)
		return s.classifyByEntryLevel(ctx, vec, note)
	}

	result := s.BuildRAGEntries(chunks)

	var titles []string
	for _, e := range result {
		titles = append(titles, e.Title)
	}
	fmt.Printf("[DEBUG][ClassifyDriverNote] Semantic chunk match: %d entries from %d chunks: %s\n", len(result), len(chunks), strings.Join(titles, ", "))
	return result
}

func (s *KnowledgeStore) classifyByEntryLevel(ctx context.Context, vec []float32, note string) []KnowledgeEntry {
	dbEntries, err := s.Repo.FindSimilar(ctx, vec, 2, 0.75)
	if err != nil {
		fmt.Printf("[WARN][KnowledgeStore] FindSimilar error, falling back to keyword: %v\n", err)
		return s.classifyByKeyword(note)
	}

	if len(dbEntries) == 0 {
		fmt.Printf("[DEBUG][ClassifyDriverNote] Semantic entry-level: no match (cosine < 0.75) for note: %q. Falling back to keyword match.\n", note)
		return s.classifyByKeyword(note)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]KnowledgeEntry, 0, len(dbEntries))
	for _, dbEntry := range dbEntries {
		if cached, ok := s.Cache[dbEntry.Slug]; ok {
			result = append(result, cached)
		} else {
			result = append(result, KnowledgeEntry{Title: dbEntry.Title, Body: dbEntry.Body})
		}
	}

	var titles []string
	for _, e := range result {
		titles = append(titles, e.Title)
	}
	fmt.Printf("[DEBUG][ClassifyDriverNote] Semantic entry-level match: %d entries (cosine ≥ 0.75): %s\n", len(result), strings.Join(titles, ", "))
	return result
}

func (s *KnowledgeStore) BuildRAGEntries(chunks []repositories.ChunkWithEntry) []KnowledgeEntry {
	type entryChunks struct {
		slug  string
		title string
		parts []string
	}
	seenOrder := []string{}
	grouped := make(map[string]*entryChunks)

	for _, c := range chunks {
		if _, ok := grouped[c.EntrySlug]; !ok {
			grouped[c.EntrySlug] = &entryChunks{
				slug:  c.EntrySlug,
				title: c.EntryTitle,
			}
			seenOrder = append(seenOrder, c.EntrySlug)
		}
		grouped[c.EntrySlug].parts = append(grouped[c.EntrySlug].parts, c.Content)
	}

	result := make([]KnowledgeEntry, 0, len(seenOrder))
	for _, slug := range seenOrder {
		ec := grouped[slug]
		result = append(result, KnowledgeEntry{
			Title: ec.title,
			Body:  strings.Join(ec.parts, "\n\n"),
		})
	}

	return result
}

func (s *KnowledgeStore) classifyByKeyword(note string) []KnowledgeEntry {
	fmt.Printf("[DEBUG][ClassifyDriverNote] Keyword fallback triggered. Returning full combined KB.\n")
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.Cache != nil {
		if entry, ok := s.Cache["combined_knowledge"]; ok {
			return []KnowledgeEntry{entry}
		}
	}

	return []KnowledgeEntry{
		{
			Title: "Order Fulfillment Knowledge Base",
			Body:  kbBodyCombined,
		},
	}
}
