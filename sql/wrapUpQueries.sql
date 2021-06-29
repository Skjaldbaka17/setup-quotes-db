UPDATE authors SET tsv = setweight(to_tsvector('english', name), 'A');
UPDATE quotes SET tsv = setweight(to_tsvector('english', quote), 'B');
CREATE INDEX if not exists index_authors_on_name ON authors USING gin(tsv);
CREATE INDEX if not exists index_quotes_on_quote ON quotes USING gin(tsv);
CREATE INDEX if not exists index_quotes_on_author_id ON quotes(author_id);
CREATE INDEX if not exists index_quotes_on_count ON quotes(count);
CREATE INDEX if not exists index_view_on_tsv ON searchview(tsv);
CREATE EXTENSION if not exists pg_trgm;
