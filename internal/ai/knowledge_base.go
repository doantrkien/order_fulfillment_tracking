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

// ClassifyDriverNote is the standalone (no-DB) keyword-based classifier.
// It is used as fallback when KnowledgeStore.embeddingClient is nil.
func ClassifyDriverNote(note string) []KnowledgeEntry {
	fmt.Printf("[DEBUG][ClassifyDriverNote] Standalone fallback triggered. Returning full combined KB.\\n")
	return []KnowledgeEntry{
		{
			Title: "Order Fulfillment Knowledge Base",
			Body:  kbBodyCombined,
		},
	}
}

func removeAccents(s string) string {
	accents := map[rune][]rune{
		'a': []rune("áàảãạăắằẳẵặâấầẩẫậ"),
		'e': []rune("éèẻẽẹêếềểễệ"),
		'i': []rune("íìỉĩị"),
		'o': []rune("óòỏõọôốồổỗộơớờởỡợ"),
		'u': []rune("úùủũụưứừửữự"),
		'y': []rune("ýỳỷỹỵ"),
		'd': []rune("đ"),
	}

	runesList := []rune(s)
	for i, r := range runesList {
		for unaccented, accentedChars := range accents {
			for _, ac := range accentedChars {
				if r == ac {
					runesList[i] = unaccented
					break
				}
			}
		}
	}
	return string(runesList)
}

func containsAny(s string, keywords []string) bool {
	normalizedS := removeAccents(s)
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
		if strings.Contains(normalizedS, removeAccents(kw)) {
			return true
		}
	}
	return false
}

// ── KnowledgeStore ─────────────────────────────────────────────────────────────

// KnowledgeStore manages an in-memory cache of KB entries backed by the database.
// When an embeddingClient is provided, ClassifyDriverNote uses vector similarity
// search instead of keyword matching.
type KnowledgeStore struct {
	repo            repositories.KnowledgeRepository
	embeddingClient embedding.Client // nil → keyword fallback
	cache           map[string]KnowledgeEntry // key: slug
	mu              sync.RWMutex
}

func NewKnowledgeStore(repo repositories.KnowledgeRepository, embeddingClient embedding.Client) *KnowledgeStore {
	return &KnowledgeStore{
		repo:            repo,
		embeddingClient: embeddingClient,
		cache:           make(map[string]KnowledgeEntry),
	}
}

func (s *KnowledgeStore) Initialize(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Query active knowledge entries from DB
	dbEntries, err := s.repo.GetAllActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to load knowledge entries from DB: %w", err)
	}

	// 2. If DB is empty, auto-seed with embedded values
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
			if err := s.repo.Save(ctx, entry); err != nil {
				fmt.Printf("[WARNING][KnowledgeStore] Failed to seed slug %q: %v\n", seed.slug, err)
			} else {
				dbEntries = append(dbEntries, *entry)
			}
		}
	}

	// 3. Generate entry-level embeddings (legacy, kept for backward compat)
	s.generateMissingEmbeddings(ctx, dbEntries)

	// 4. Chunk entries and generate chunk-level embeddings for RAG
	s.rechunkAndEmbed(ctx, dbEntries)

	// 5. Populate memory cache
	s.cache = make(map[string]KnowledgeEntry)
	for _, dbEntry := range dbEntries {
		s.cache[dbEntry.Slug] = KnowledgeEntry{
			Title: dbEntry.Title,
			Body:  dbEntry.Body,
		}
	}

	fmt.Printf("[INFO][KnowledgeStore] Loaded %d active knowledge entries into memory cache.\n", len(s.cache))
	return nil
}

func (s *KnowledgeStore) Reload(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dbEntries, err := s.repo.GetAllActive(ctx)
	if err != nil {
		return err
	}

	// Re-embed any entry that was added/updated and still has needs_reembed=true
	s.generateMissingEmbeddings(ctx, dbEntries)

	// Re-chunk and embed entries that need it
	s.rechunkAndEmbed(ctx, dbEntries)

	s.cache = make(map[string]KnowledgeEntry)
	for _, dbEntry := range dbEntries {
		s.cache[dbEntry.Slug] = KnowledgeEntry{
			Title: dbEntry.Title,
			Body:  dbEntry.Body,
		}
	}

	fmt.Printf("[INFO][KnowledgeStore] Reloaded %d active knowledge entries from DB.\n", len(s.cache))
	return nil
}

// generateMissingEmbeddings calls embeddingClient for entries where NeedsReembed is true.
// It is a best-effort operation — errors are logged but never returned.
func (s *KnowledgeStore) generateMissingEmbeddings(ctx context.Context, entries []models.KnowledgeEntry) {
	if s.embeddingClient == nil {
		return
	}
	for _, entry := range entries {
		if !entry.NeedsReembed {
			continue
		}
		vec, err := s.embeddingClient.Embed(ctx, entry.Body)
		if err != nil {
			fmt.Printf("[WARNING][KnowledgeStore] Embed failed for %q: %v\n", entry.Slug, err)
			continue
		}
		if updateErr := s.repo.UpdateEmbedding(ctx, entry.ID, vec); updateErr != nil {
			fmt.Printf("[WARNING][KnowledgeStore] UpdateEmbedding failed for %q (id=%d): %v\n", entry.Slug, entry.ID, updateErr)
		} else {
			fmt.Printf("[INFO][KnowledgeStore] Embedded entry %q (id=%d, dims=%d)\n", entry.Slug, entry.ID, len(vec))
		}
	}
}

// rechunkAndEmbed splits each entry that needs re-embedding into chunks,
// saves them to the knowledge_chunks table, and generates chunk-level embeddings.
// This is a best-effort operation — errors are logged but never returned.
func (s *KnowledgeStore) rechunkAndEmbed(ctx context.Context, entries []models.KnowledgeEntry) {
	if s.embeddingClient == nil {
		return
	}

	for _, entry := range entries {
		if !entry.NeedsReembed {
			continue
		}

		// 1. Delete old chunks for this entry
		if err := s.repo.DeleteChunksByEntryID(ctx, entry.ID); err != nil {
			fmt.Printf("[WARNING][KnowledgeStore] DeleteChunks failed for %q (id=%d): %v\n", entry.Slug, entry.ID, err)
			continue
		}

		// 2. Chunk the entry body by markdown sections
		chunks := ChunkMarkdown(entry.Body, entry.Title, DefaultMinChunkTokens)
		if len(chunks) == 0 {
			continue
		}

		// 3. Build model objects
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

		// 4. Save chunks to DB
		if err := s.repo.SaveChunks(ctx, dbChunks); err != nil {
			fmt.Printf("[WARNING][KnowledgeStore] SaveChunks failed for %q: %v\n", entry.Slug, err)
			continue
		}
		fmt.Printf("[INFO][KnowledgeStore] Created %d chunks for entry %q (id=%d)\n", len(dbChunks), entry.Slug, entry.ID)

		// 5. Embed each chunk and save embedding
		for _, dbChunk := range dbChunks {
			vec, err := s.embeddingClient.Embed(ctx, dbChunk.Content)
			if err != nil {
				fmt.Printf("[WARNING][KnowledgeStore] Embed chunk failed for %q chunk#%d: %v\n", entry.Slug, dbChunk.ChunkIndex, err)
				continue
			}
			if updateErr := s.repo.UpdateChunkEmbedding(ctx, dbChunk.ID, vec); updateErr != nil {
				fmt.Printf("[WARNING][KnowledgeStore] UpdateChunkEmbedding failed for chunk id=%d: %v\n", dbChunk.ID, updateErr)
			} else {
				fmt.Printf("[INFO][KnowledgeStore] Embedded chunk %q #%d (id=%d, dims=%d)\n", entry.Slug, dbChunk.ChunkIndex, dbChunk.ID, len(vec))
			}
		}
	}
}



// ClassifyDriverNote classifies a driver note against the knowledge base.
//
// If an embeddingClient is available, it uses pgvector cosine-similarity search
// (semantic). Otherwise it falls back to the keyword-based classifier.
//
// ctx is required for the embedding API call; the keyword path ignores it.
func (s *KnowledgeStore) ClassifyDriverNote(ctx context.Context, note string) []KnowledgeEntry {
	// ── Semantic path ────────────────────────────────────────────────────────
	if s.embeddingClient != nil {
		return s.classifyBySemantic(ctx, note)
	}

	// ── Keyword fallback ─────────────────────────────────────────────────────
	return s.classifyByKeyword(note)
}

// classifyBySemantic embeds the note and queries pgvector for the top-k similar chunks.
// It uses chunk-level retrieval for fine-grained semantic matching, then groups chunks
// by parent entry to build enriched KnowledgeEntry results.
// Falls back to entry-level search, then keyword matching on any error.
func (s *KnowledgeStore) classifyBySemantic(ctx context.Context, note string) []KnowledgeEntry {
	vec, err := s.embeddingClient.Embed(ctx, note)
	if err != nil {
		fmt.Printf("[WARN][KnowledgeStore] Embed error, falling back to keyword: %v\n", err)
		return s.classifyByKeyword(note)
	}

	// Try chunk-level retrieval first (RAG path)
	chunks, err := s.repo.FindSimilarChunks(ctx, vec, 5, 0.65)
	if err != nil {
		fmt.Printf("[WARN][KnowledgeStore] FindSimilarChunks error, trying entry-level: %v\n", err)
		// Fall back to entry-level search
		return s.classifyByEntryLevel(ctx, vec, note)
	}

	if len(chunks) == 0 {
		fmt.Printf("[DEBUG][ClassifyDriverNote] Semantic chunks: no match (cosine < 0.65) for note: %q. Trying entry-level.\n", note)
		// Fall back to entry-level search with original threshold
		return s.classifyByEntryLevel(ctx, vec, note)
	}

	// Build enriched entries from chunks (inject only relevant sections)
	result := s.buildRAGEntries(chunks)

	var titles []string
	for _, e := range result {
		titles = append(titles, e.Title)
	}
	fmt.Printf("[DEBUG][ClassifyDriverNote] Semantic chunk match: %d entries from %d chunks: %s\n", len(result), len(chunks), strings.Join(titles, ", "))
	return result
}

// classifyByEntryLevel is the original entry-level semantic search, kept as
// fallback when chunk-level search returns no results or errors.
func (s *KnowledgeStore) classifyByEntryLevel(ctx context.Context, vec []float32, note string) []KnowledgeEntry {
	dbEntries, err := s.repo.FindSimilar(ctx, vec, 2, 0.75)
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
		if cached, ok := s.cache[dbEntry.Slug]; ok {
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

// buildRAGEntries groups chunks by their parent entry and constructs KnowledgeEntry
// objects whose Body contains only the relevant chunk content (not the full entry body).
// This reduces token usage in the AI prompt while maintaining semantic precision.
func (s *KnowledgeStore) buildRAGEntries(chunks []repositories.ChunkWithEntry) []KnowledgeEntry {
	// Group chunks by entry slug, preserving order (first seen = highest similarity)
	type entryChunks struct {
		slug   string
		title  string
		parts  []string
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

// classifyByKeyword is the fallback when embedding is unavailable or yields no results.
// It returns the entire combined knowledge base so the LLM has all context.
func (s *KnowledgeStore) classifyByKeyword(note string) []KnowledgeEntry {
	fmt.Printf("[DEBUG][ClassifyDriverNote] Keyword fallback triggered. Returning full combined KB.\n")
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.cache != nil {
		if entry, ok := s.cache["combined_knowledge"]; ok {
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
