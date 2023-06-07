package stars

import "time"

type Star struct {
	ID         int        `json:"id"`
	FirstName  string     `json:"first_name" `
	MiddleName *string    `json:"middle_name,omitempty"`
	LastName   string     `json:"last_name" `
	BirthDate  time.Time  `json:"birth_date" `
	BirthPlace *string    `json:"birth_place,omitempty" `
	DeathDate  *time.Time `json:"death_date,omitempty"`
	Bio        *string    `json:"bio,omitempty"`
	CreatedAt  time.Time  `json:"created_ad"`
	DeletedAt  *time.Time `json:"deleted_ad,omitempty"`
}
