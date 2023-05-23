package users

import "time"

const (
	UserRole   = "user"
	EditorRole = "editor"
	AdminRole  = "admin"
)

type User struct {
	ID        int64      `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	Bio       *string    `json:"bio,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

type UserWithPassword struct {
	*User
	PasswordHash string
}

func newUserWithPassword() *UserWithPassword {
	return &UserWithPassword{
		User: &User{},
	}
}
