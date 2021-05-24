ALTER TABLE quotes ADD COLUMN tsv tsvector; --Create tsvector column
UPDATE quotes SET tsv =
    setweight(to_tsvector('english', quote), 'A');

ALTER TABLE authors ADD COLUMN tsv tsvector; --Create tsvector column
UPDATE authors SET tsv =
    setweight(to_tsvector('english', name), 'A');

CREATE INDEX index_authors_on_name ON authors USING gin(tsv);
CREATE INDEX index_quotes_on_quote ON quotes USING gin(tsv);
CREATE INDEX index_authors_on_sim_name on authors using gin()
CREATE EXTENSION pg_trgm; --https://www.freecodecamp.org/news/fuzzy-string-matching-with-postgresql/ needed for SIMILARITY()-function
