CREATE TABLE quoteoftheday (
    id serial not null,
    quoteid integer,
    date date unique default current_date,
    created_at timestamptz default current_timestamp,
    updated_at timestamptz
)