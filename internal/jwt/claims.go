package jwt

import "github.com/golang-jwt/jwt"

type AccessClaims struct {
	jwt.StandardClaims
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
}
