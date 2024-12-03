-- remove indexes
DROP INDEX IF EXISTS idx_movies_title_trigram;
DROP INDEX IF EXISTS idx_movies_genres;

-- remove trgm
DROP EXTENSION IF EXISTS pg_trgm
