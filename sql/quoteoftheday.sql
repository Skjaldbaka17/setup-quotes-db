CREATE TABLE quoteoftheday (
    id serial not null,
    quoteid integer,
    created_at timestamptz default current_timestamp,
    updated_at timestamptz
)