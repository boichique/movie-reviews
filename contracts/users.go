package contracts

import "time"

type User struct {
	ID        int        `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	Bio       *string    `json:"bio,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type GetOrDeleteUserRequest struct {
	UserID int `param:"userid" validate:"nonzero"`
}

type GetUserByUsernameRequest struct {
	Username string `param:"username" validate:"nonzero"`
}

type UpdateUserBioRequest struct {
	UserID int     `param:"userid" validate:"nonzero"`
	Bio    *string `json:"bio"`
}

type UpdateUserRoleRequest struct {
	UserID int    `param:"userid" validate:"nonzero"`
	Role   string `param:"role" validate:"role"`
}
