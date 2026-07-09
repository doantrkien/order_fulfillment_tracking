DROP INDEX IF EXISTS knowledge_entries_embedding_hnsw_idx;
ALTER TABLE knowledge_entries DROP COLUMN IF EXISTS needs_reembed;
ALTER TABLE knowledge_entries DROP COLUMN IF EXISTS embedding;
DROP EXTENSION IF EXISTS vector;
