-- enable trigram extension for LIKE performance on search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- indexes
CREATE INDEX idx_movies_title_trigram ON movies USING gin (lower(title) gin_trgm_ops);
CREATE INDEX idx_movies_genres ON movies USING gin (genres);
