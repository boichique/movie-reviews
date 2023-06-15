CREATE TABLE movie_genres
(
    movie_id INTEGER NOT NULL REFERENCES movies(id),
    genre_id INTEGER NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    order_no SMALLINT NOT NULL ,
    PRIMARY KEY (movie_id, genre_id)
);
CREATE INDEX idx_movie_genres_movie_id ON movie_genres(movie_id);
CREATE INDEX idx_movie_genres_genre_id ON movie_genres(genre_id);

---- create above / drop below ----

DROP INDEX idx_movie_genres_genre_id;
DROP INDEX idx_movie_genres_movie_id;
DROP TABLE movie_genres;