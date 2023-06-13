ALTER TABLE movies ADD COLUMN version integer NOT NULL DEFAULT 0;

---- create above / drop below ----

ALTER TABLE movies DROP COLUMN version;