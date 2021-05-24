UPDATE authors SET tsv = setweight(to_tsvector('english', name), 'A');
UPDATE quotes SET tsv = setweight(to_tsvector('english', quote), 'A');
CREATE INDEX index_authors_on_name ON authors USING gin(tsv);
CREATE INDEX index_quotes_on_quote ON quotes USING gin(tsv);
CREATE EXTENSION pg_trgm;
