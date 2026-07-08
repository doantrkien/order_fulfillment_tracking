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

//go:embed knowledge/state_machine.md
var kbBodyStateMachine string

//go:embed knowledge/delivery_failure.md
var kbBodyDeliveryFailure string

//go:embed knowledge/stuck_order.md
var kbBodyStuckOrder string

//go:embed knowledge/duplicate_event.md
var kbBodyDuplicateEvent string

//go:embed knowledge/skipped_status.md
var kbBodySkippedStatus string

//go:embed knowledge/alternative_success.md
var kbBodyAlternativeSuccess string

//go:embed knowledge/cancellation_edge_case.md
var kbBodyCancellation string

//go:embed knowledge/refund_edge_case.md
var kbBodyRefund string

type KnowledgeEntry struct {
	Title string
	Body  string
}

// ── Compiled KB entries (one per exception domain) ───────────────────────────
var kbStateMachine = KnowledgeEntry{
	Title: "Order State Machine",
	Body:  kbBodyStateMachine,
}

var kbDeliveryFailure = KnowledgeEntry{
	Title: "Delivery Failure",
	Body:  kbBodyDeliveryFailure,
}

var kbStuckOrder = KnowledgeEntry{
	Title: "Stuck Order",
	Body:  kbBodyStuckOrder,
}

var kbDuplicateEvent = KnowledgeEntry{
	Title: "Duplicate Event",
	Body:  kbBodyDuplicateEvent,
}

var kbSkippedStatus = KnowledgeEntry{
	Title: "Skipped Status",
	Body:  kbBodySkippedStatus,
}

var kbAlternativeSuccess = KnowledgeEntry{
	Title: "Alternative Delivery Success",
	Body:  kbBodyAlternativeSuccess,
}

var kbCancellation = KnowledgeEntry{
	Title: "Cancellation Edge Case",
	Body:  kbBodyCancellation,
}

var kbRefund = KnowledgeEntry{
	Title: "Refund Edge Case",
	Body:  kbBodyRefund,
}

var deliveryFailureKeywordsKB = []string{
	"lost", "stolen", "cannot find", "missing parcel", "hàng bị mất", "nghi thất lạc",

	"accident", "vehicle breakdown", "xe hỏng", "tai nạn", "bad weather", "thời tiết",
	"failed", "failure", "cannot deliver", "could not deliver", "giao thất bại",

	"not home", "no one home", "customer not home", "wrong address", "address not found",
	"không có nhà", "sai địa chỉ", "không liên lạc được",

	"temporarily unreachable", "no answer", "will retry", "khách không nghe máy", "cannot contact", "unreachable",
}

var stateMachineKeywordsKB = []string{
	"tự động đổi trạng thái", "sai trạng thái hiện tại",
}

var duplicateKeywordsKB = []string{
	"bấm lại", "duplicate", "lặp", "trùng", "repeated", "nhiều lần", "times", "hệ thống ghi nhận delivered 2 lần.",
}

var skippedKeywordsKB = []string{
	"skipped", "missed", "bỏ qua", "thiếu", "không đi qua các bước thông thường",
}

var stuckKeywordsKB = []string{
	"stuck", "no progress", "not moved", "delay", "delayed", "late",
	"trễ", "chậm", "không tiến triển", "Đơn hàng bị giữ lại",
}

var successKeywordsKB = []string{
	"reception", "lễ tân", "neighbor", "hàng xóm", "bảo vệ", "security", "front door", "trước cửa", "thành công", "delivered",
}

var cancellationKeywordsKB = []string{
	"cancel",
	"cancelled",
	"cancellation",
	"customer requested cancellation",
	"cancel request",
	"hủy",
	"hủy đơn",
	"yêu cầu hủy",
	"đổi ý",
}

var refundKeywordsKB = []string{
	"refund",
	"refunded",
	"refund request",
	"duplicate payment",
	"chargeback",
	"hoàn tiền",
	"yêu cầu hoàn tiền",
	"thanh toán nhầm",
}

// ClassifyDriverNote is the standalone (no-DB) keyword-based classifier.
// It is used as fallback when KnowledgeStore.embeddingClient is nil.
func ClassifyDriverNote(note string) []KnowledgeEntry {
	lower := strings.ToLower(strings.TrimSpace(note))
	entries := []KnowledgeEntry{} // always present

	matched := false
	isSuccessAlternative := containsAny(lower, successKeywordsKB)

	if isSuccessAlternative {
		entries = append(entries, kbAlternativeSuccess)
		matched = true
	}

	if containsAny(lower, deliveryFailureKeywordsKB) && !isSuccessAlternative {
		entries = append(entries, kbDeliveryFailure)
		matched = true
	}

	if containsAny(lower, stateMachineKeywordsKB) && !isSuccessAlternative {
		entries = append(entries, kbStateMachine)
		matched = true
	}
	if containsAny(lower, duplicateKeywordsKB) {
		entries = append(entries, kbDuplicateEvent)
		matched = true
	}
	if containsAny(lower, skippedKeywordsKB) {
		entries = append(entries, kbSkippedStatus)
		matched = true
	}
	if containsAny(lower, stuckKeywordsKB) {
		entries = append(entries, kbStuckOrder)
		matched = true
	}

	if containsAny(lower, cancellationKeywordsKB) {
		entries = append(entries, kbCancellation)
		matched = true
	}

	if containsAny(lower, refundKeywordsKB) {
		entries = append(entries, kbRefund)
		matched = true
	}

	var titles []string
	for _, e := range entries {
		titles = append(titles, e.Title)
	}
	fmt.Printf("[DEBUG][ClassifyDriverNote] Note: %q | Matched: %v | KB Sent: %d (%s)\n", note, matched, len(entries), strings.Join(titles, ", "))

	return entries
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
			{"state_machine", "Order State Machine", kbBodyStateMachine},
			{"delivery_failure", "Delivery Failure", kbBodyDeliveryFailure},
			{"stuck_order", "Stuck Order", kbBodyStuckOrder},
			{"duplicate_event", "Duplicate Event", kbBodyDuplicateEvent},
			{"skipped_status", "Skipped Status", kbBodySkippedStatus},
			{"alternative_success", "Alternative Delivery Success", kbBodyAlternativeSuccess},
			{"cancellation_edge_case", "Cancellation Edge Case", kbBodyCancellation},
			{"refund_edge_case", "Refund Edge Case", kbBodyRefund},
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

	// 3. Generate embeddings for entries that need it (NeedsReembed=true or embedding IS NULL)
	s.generateMissingEmbeddings(ctx, dbEntries)

	// 4. Populate memory cache
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

func (s *KnowledgeStore) GetStateMachine() KnowledgeEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cache != nil {
		if entry, ok := s.cache["state_machine"]; ok {
			return entry
		}
	}
	return kbStateMachine
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

// classifyBySemantic embeds the note and queries pgvector for the top-k similar entries.
// Falls back to keyword matching on any error.
func (s *KnowledgeStore) classifyBySemantic(ctx context.Context, note string) []KnowledgeEntry {
	vec, err := s.embeddingClient.Embed(ctx, note)
	if err != nil {
		fmt.Printf("[WARN][KnowledgeStore] Embed error, falling back to keyword: %v\n", err)
		return s.classifyByKeyword(note)
	}

	dbEntries, err := s.repo.FindSimilar(ctx, vec, 2, 0.75)
	if err != nil {
		fmt.Printf("[WARN][KnowledgeStore] FindSimilar error, falling back to keyword: %v\n", err)
		return s.classifyByKeyword(note)
	}

	if len(dbEntries) == 0 {
		fmt.Printf("[DEBUG][ClassifyDriverNote] Semantic: no match (cosine < 0.75) for note: %q. Falling back to keyword match.\n", note)
		return s.classifyByKeyword(note)
	}

	// Map DB entries back to in-memory KnowledgeEntry (use cache to get latest body)
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
	fmt.Printf("[DEBUG][ClassifyDriverNote] Semantic match: %d entries (cosine ≥ 0.75): %s\n", len(result), strings.Join(titles, ", "))
	return result
}

// classifyByKeyword is the original keyword-matching implementation,
// preserved as fallback when embedding is unavailable.
func (s *KnowledgeStore) classifyByKeyword(note string) []KnowledgeEntry {
	lower := strings.ToLower(strings.TrimSpace(note))
	entries := []KnowledgeEntry{}

	getEntry := func(slug string, fallback KnowledgeEntry) KnowledgeEntry {
		s.mu.RLock()
		defer s.mu.RUnlock()
		if s.cache != nil {
			if entry, ok := s.cache[slug]; ok {
				return entry
			}
		}
		return fallback
	}

	matched := false
	isSuccessAlternative := containsAny(lower, successKeywordsKB)

	if isSuccessAlternative {
		entries = append(entries, getEntry("alternative_success", kbAlternativeSuccess))
		matched = true
	}

	if containsAny(lower, deliveryFailureKeywordsKB) && !isSuccessAlternative {
		entries = append(entries, getEntry("delivery_failure", kbDeliveryFailure))
		matched = true
	}

	if containsAny(lower, stateMachineKeywordsKB) && !isSuccessAlternative {
		entries = append(entries, getEntry("state_machine", kbStateMachine))
		matched = true
	}
	if containsAny(lower, duplicateKeywordsKB) {
		entries = append(entries, getEntry("duplicate_event", kbDuplicateEvent))
		matched = true
	}
	if containsAny(lower, skippedKeywordsKB) {
		entries = append(entries, getEntry("skipped_status", kbSkippedStatus))
		matched = true
	}
	if containsAny(lower, stuckKeywordsKB) {
		entries = append(entries, getEntry("stuck_order", kbStuckOrder))
		matched = true
	}
	if containsAny(lower, cancellationKeywordsKB) {
		entries = append(entries, getEntry("cancellation_edge_case", kbCancellation))
		matched = true
	}
	if containsAny(lower, refundKeywordsKB) {
		entries = append(entries, getEntry("refund_edge_case", kbRefund))
		matched = true
	}

	var titles []string
	for _, e := range entries {
		titles = append(titles, e.Title)
	}
	fmt.Printf("[DEBUG][ClassifyDriverNote] Keyword match: %v | KB Sent: %d (%s)\n", matched, len(entries), strings.Join(titles, ", "))

	return entries
}
