-- Enable pgvector extension (requires pgvector/pgvector:pg16 docker image)
CREATE EXTENSION IF NOT EXISTS vector;

-- Add embedding column (768 dimensions for Gemini text-embedding-004)
ALTER TABLE knowledge_entries ADD COLUMN embedding vector(768);

-- Flag to mark entries that need (re)embedding
ALTER TABLE knowledge_entries ADD COLUMN needs_reembed BOOLEAN NOT NULL DEFAULT true;

-- HNSW index for fast cosine similarity search
-- (better than IVFFlat for small datasets like 8–20 KB entries)
CREATE INDEX IF NOT EXISTS knowledge_entries_embedding_hnsw_idx
    ON knowledge_entries
    USING hnsw (embedding vector_cosine_ops);
