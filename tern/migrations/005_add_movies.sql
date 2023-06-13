CREATE TABLE movies (
                        id SERIAL PRIMARY KEY,
                        title VARCHAR(255) NOT NULL,
                        description TEXT NOT NULL,
                        release_date DATE NOT NULL,
                        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                        deleted_at TIMESTAMP
);
---- create above / drop below ----
DROP TABLE movies;