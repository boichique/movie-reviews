CREATE TYPE movie_role AS ENUM ('actor', 'voice actor', 'writer', 'producer', 'director', 'composer');
CREATE TABLE movie_stars (
    movie_id INTEGER NOT NULL REFERENCES movies(id),
    star_id INTEGER NOT NULL REFERENCES stars(id),
    role movie_role NOT NULL,
    details TEXT,
    order_no SMALLINT NOT NULL,
    PRIMARY KEY (movie_id, star_id, role)
);

CREATE INDEX idx_movie_stars_movie_id ON movie_stars(movie_id);
CREATE INDEX idx_movie_stars_star_id ON movie_stars(star_id);


---- create above / drop below ----

DROP INDEX idx_movie_stars_star_id;
DROP INDEX idx_movie_stars_movie_id;
DROP TABLE movie_stars;
DROP TYPE movie_role;