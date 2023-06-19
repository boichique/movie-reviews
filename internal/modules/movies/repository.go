package movies

import (
	"context"
	"time"

	"github.com/boichique/movie-reviews/internal/apperrors"
	"github.com/boichique/movie-reviews/internal/dbx"
	"github.com/boichique/movie-reviews/internal/modules/genres"
	"github.com/boichique/movie-reviews/internal/modules/stars"
	"github.com/boichique/movie-reviews/internal/slices"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db         *pgxpool.Pool
	genresRepo *genres.Repository
	starRepo   *stars.Repository
}

func NewRepository(db *pgxpool.Pool, genresRepo *genres.Repository, starRepo *stars.Repository) *Repository {
	return &Repository{
		db:         db,
		genresRepo: genresRepo,
		starRepo:   starRepo,
	}
}

func (r *Repository) Create(ctx context.Context, movie *MovieDetails) error {
	err := dbx.InTransaction(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		err := tx.
			QueryRow(
				ctx,
				`INSERT INTO movies (title, description, release_date)
				VALUES ($1, $2, $3) 
				RETURNING id, created_at`,
				movie.Title, movie.Description, movie.ReleaseDate).
			Scan(
				&movie.ID,
				&movie.CreatedAt,
			)
		if err != nil {
			return apperrors.Internal(err)
		}

		nextGenres := slices.MapIndex(movie.Genres, func(i int, g *genres.Genre) *genres.MovieGenreRelation {
			return &genres.MovieGenreRelation{
				MovieID: movie.ID,
				GenreID: g.ID,
				OrderNo: i,
			}
		})
		if err = r.updateGenres(ctx, nil, nextGenres); err != nil {
			return err
		}

		nextCast := slices.MapIndex(movie.Cast, func(i int, c *stars.MovieCredit) *stars.MovieStarRelation {
			return &stars.MovieStarRelation{
				MovieID: movie.ID,
				StarID:  c.Star.ID,
				Role:    c.Role,
				Details: c.Details,
				OrderNo: i,
			}
		})

		return r.updateCast(ctx, nil, nextCast)
	})
	if err != nil {
		return apperrors.Internal(err)
	}

	return nil
}

func (r *Repository) GetMoviesPaginated(ctx context.Context, searchTerm *string, starID *int, offset int, limit int) ([]*Movie, int, error) {
	b := &pgx.Batch{}
	selectQuery := dbx.StatementBuilder.
		Select("id, title,  release_date, created_at").
		From("movies").
		Where("deleted_at IS NULL").
		Limit(uint64(limit)).
		Offset(uint64(offset))
	countQuery := dbx.StatementBuilder.
		Select("count(*)").
		From("movies").
		Where("deleted_at IS NULL")

	if starID != nil {
		selectQuery = selectQuery.
			Join("movie_stars on movies.id = movie_stars.movie_id").
			Where("movie_stars.star_id = ?", starID)

		countQuery = countQuery.
			Join("movie_stars on movies.id = movie_stars.movie_id").
			Where("movie_stars.star_id = ?", starID)
	}

	if searchTerm != nil {
		selectQuery = selectQuery.
			Where("search_vector @@ to_tsquery('english', ?)", *searchTerm).
			OrderByClause("ts_rank_cd(search_vector, to_tsquery('english', ?)) DESC", *searchTerm)

		countQuery = countQuery.
			Where("search_vector @@ to_tsquery('english', ?)", *searchTerm)

	}

	if err := dbx.QueueBatchSelect(b, selectQuery); err != nil {
		return nil, 0, apperrors.Internal(err)
	}

	if err := dbx.QueueBatchSelect(b, countQuery); err != nil {
		return nil, 0, apperrors.Internal(err)
	}

	br := r.db.SendBatch(ctx, b)
	defer br.Close()

	rows, err := br.Query()
	if err != nil {
		return nil, 0, apperrors.Internal(err)
	}
	defer rows.Close()

	var movies []*Movie
	for rows.Next() {
		var movie Movie
		if err = rows.
			Scan(&movie.ID,
				&movie.Title,
				&movie.ReleaseDate,
				&movie.CreatedAt,
			); err != nil {
			return nil, 0, apperrors.Internal(err)
		}
		movies = append(movies, &movie)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, apperrors.Internal(err)
	}

	var total int
	if err = br.QueryRow().Scan(&total); err != nil {
		return nil, 0, apperrors.Internal(err)
	}

	return movies, total, err
}

func (r *Repository) GetByID(ctx context.Context, id int) (*MovieDetails, error) {
	var movie MovieDetails

	err := r.db.
		QueryRow(
			ctx,
			`SELECT id, version ,title, description, release_date, created_at 
			FROM movies 
			WHERE id = $1 
			AND deleted_at IS NULL;`,
			id,
		).
		Scan(
			&movie.ID,
			&movie.Version,
			&movie.Title,
			&movie.Description,
			&movie.ReleaseDate,
			&movie.CreatedAt,
		)
	switch {
	case dbx.IsNoRows(err):
		return nil, errMovieWithNotFound(id)
	case err != nil:
		return nil, apperrors.Internal(err)
	}

	return &movie, nil
}

func (r *Repository) Update(ctx context.Context, movie *MovieDetails) error {
	err := dbx.InTransaction(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		n, err := tx.
			Exec(
				ctx,
				`UPDATE movies 
			SET version = version + 1, 
			title = $1,
			description = $2, 
			release_date = $3 
			WHERE id = $4 
			AND version = $5;`,
				movie.Title,
				movie.Description,
				movie.ReleaseDate,
				movie.ID,
				movie.Version,
			)
		if err != nil {
			return apperrors.Internal(err)
		}

		if n.RowsAffected() == 0 {
			_, err = r.GetByID(ctx, movie.ID)
			if err != nil {
				return err
			}

			return apperrors.VersionMismatch("movie", "id", movie.ID, movie.Version)
		}

		currentGenres, err := r.genresRepo.GetRelationByMovieID(ctx, movie.ID)
		if err != nil {
			return err
		}

		nextGenres := slices.MapIndex(movie.Genres, func(i int, g *genres.Genre) *genres.MovieGenreRelation {
			return &genres.MovieGenreRelation{
				GenreID: g.ID,
				MovieID: movie.ID,
				OrderNo: i,
			}
		})

		if err = r.updateGenres(ctx, currentGenres, nextGenres); err != nil {
			return err
		}

		currentCast, err := r.starRepo.GetRelationByMovieID(ctx, movie.ID)
		if err != nil {
			return err
		}

		nextCast := slices.MapIndex(movie.Cast, func(i int, c *stars.MovieCredit) *stars.MovieStarRelation {
			return &stars.MovieStarRelation{
				MovieID: movie.ID,
				StarID:  c.Star.ID,
				Role:    c.Role,
				Details: c.Details,
				OrderNo: i,
			}
		})

		return r.updateCast(ctx, currentCast, nextCast)
	})
	if err != nil {
		return apperrors.EnsureInternal(err)
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, movieID int) error {
	n, err := r.db.
		Exec(
			ctx,
			`UPDATE movies 
			SET deleted_at = $1 
			WHERE id = $2 
			AND deleted_at IS NULL;`,
			time.Now(), movieID,
		)
	if err != nil {
		return apperrors.Internal(err)
	}

	if n.RowsAffected() == 0 {
		return errMovieWithNotFound(movieID)
	}

	return nil
}

func errMovieWithNotFound(movieID int) error {
	return apperrors.NotFound("movie", "id", movieID)
}

func (r *Repository) updateGenres(ctx context.Context, current, next []*genres.MovieGenreRelation) error {
	q := dbx.FromContext(ctx, r.db)

	addFunc := func(mgo *genres.MovieGenreRelation) error {
		_, err := q.Exec(
			ctx,
			`INSERT INTO movie_genres (movie_id, genre_id, order_no)
			VALUES ($1, $2, $3)`,
			mgo.MovieID, mgo.GenreID, mgo.OrderNo)
		return err
	}

	removeFn := func(mgo *genres.MovieGenreRelation) error {
		_, err := q.Exec(
			ctx,
			`DELETE FROM movie_genres
			WHERE movie_id = $1 
			AND genre_id = $2`,
			mgo.MovieID, mgo.GenreID)
		return err
	}

	return dbx.AdjustRelations(current, next, addFunc, removeFn)
}

func (r *Repository) updateCast(ctx context.Context, current, next []*stars.MovieStarRelation) error {
	q := dbx.FromContext(ctx, r.db)

	addFunc := func(mgo *stars.MovieStarRelation) error {
		_, err := q.
			Exec(
				ctx,
				`INSERT INTO movie_stars (movie_id, star_id, role, details, order_no) 
				VALUES ($1, $2, $3, $4, $5)`,
				mgo.MovieID,
				mgo.StarID,
				mgo.Role,
				mgo.Details,
				mgo.OrderNo,
			)
		return err
	}

	removeFn := func(mgo *stars.MovieStarRelation) error {
		_, err := q.
			Exec(
				ctx,
				`DELETE FROM movie_stars 
				WHERE movie_id = $1 
				AND star_id = $2 
				AND role = $3`,
				mgo.MovieID,
				mgo.StarID,
				mgo.Role,
			)
		return err
	}

	return dbx.AdjustRelations(current, next, addFunc, removeFn)
}
