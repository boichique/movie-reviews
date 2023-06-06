CREATE TABLE genres
(
    id SERIAL PRIMARY KEY,
    name VARCHAR(32) UNIQUE NOT NULL
);

---- create above / drop below ----

DROP TABLE genres;