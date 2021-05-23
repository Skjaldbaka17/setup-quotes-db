CREATE INDEX index_authors_on_name ON authors USING gin(to_tsvector(name));
CREATE INDEX index_quotes_on_quote ON quotes USING gin(to_tsvector(quote));
CREATE EXTENSION pg_trgm; --https://www.freecodecamp.org/news/fuzzy-string-matching-with-postgresql/
