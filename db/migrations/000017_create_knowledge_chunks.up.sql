-- Create knowledge_chunks table for chunk-level RAG embeddings
CREATE TABLE knowledge_chunks (
    id              SERIAL PRIMARY KEY,
    entry_id        INTEGER NOT NULL REFERENCES knowledge_entries(id) ON DELETE CASCADE,
    chunk_index     SMALLINT NOT NULL,
    heading         VARCHAR(256) NOT NULL,
    content         TEXT NOT NULL,
    token_count     INTEGER NOT NULL DEFAULT 0,
    embedding       vector(768),
    needs_reembed   BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    UNIQUE(entry_id, chunk_index)
);

-- HNSW index for fast cosine similarity search on chunks
CREATE INDEX knowledge_chunks_embedding_hnsw_idx
    ON knowledge_chunks
    USING hnsw (embedding vector_cosine_ops);

-- Index to query chunks by parent entry
CREATE INDEX knowledge_chunks_entry_id_idx ON knowledge_chunks(entry_id);
