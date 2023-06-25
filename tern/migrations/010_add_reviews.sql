CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    movie_id INTEGER REFERENCES movies(id),
    user_id INTEGER REFERENCES users(id),
    rating INTEGER,
    title varchar(255),
    content varchar(2000),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    UNIQUE (movie_id, user_id)
);
CREATE INDEX idx_reviews_movie_id ON reviews(movie_id);
CREATE INDEX idx_reviews_user_id ON reviews(user_id);

---- create above / drop below ----

DROP INDEX idx_reviews_movie_id;
DROP INDEX idx_reviews_user_id;
DROP TABLE reviews;