ALTER TABLE movies ADD COLUMN avg_rating float4;
CREATE INDEX idx_movies_avg_rating ON movies(avg_rating);

---- create above / drop below ----

DROP INDEX idx_movies_avg_rating;
ALTER TABLE movies DROP COLUMN avg_rating;