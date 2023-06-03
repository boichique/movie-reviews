package contracts

type RegisterUserRequest struct {
	Username string `json:"username" validate:"min=5,max=16"`
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"password"`
}

type LoginUserRequest struct {
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"password"`
}

type LoginUserResponse struct {
	AccessToken string `json:"access_token" validate:"nonzero"`
}

type AuthenticatedRequest[T any] struct {
	AccessToken string
	Request     T
}

func NewAuthenticated[T any](request T, accessToken string) *AuthenticatedRequest[T] {
	return &AuthenticatedRequest[T]{
		Request:     request,
		AccessToken: accessToken,
	}
}
