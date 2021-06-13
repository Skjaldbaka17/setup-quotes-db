CREATE TABLE users(
   id SERIAL PRIMARY KEY,
   email varchar not null unique,
   name VARCHAR NOT NULL UNIQUE,
   apikeyhash varchar not null unique,
   passwordhash text not null,
   email varchar,
   tier varchar not null default 'free',
   createdat timestamptz,
   updatedat timestamptz,
   deletedat timestamptz
);