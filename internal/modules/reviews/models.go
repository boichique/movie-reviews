package reviews

import "time"

type Review struct {
	ID        int        `json:"id"`
	MovieID   int        `json:"movie_id"`
	UserID    int        `json:"user_id"`
	Rating    int        `json:"rating"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
