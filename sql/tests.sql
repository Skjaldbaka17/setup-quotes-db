

SELECT word from unique_lexeme
WHERE similarity(word, 'nietshe') > 0.3
ORDER BY word <-> 'nietshe'
LIMIT 3;

REFRESH MATERIALIZED VIEW unique_lexeme;
REFRESH MATERIALIZED VIEW searchview;