package ds

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type JWTClaims struct {
	jwt.RegisteredClaims
	UserUUID  uuid.UUID `json:"user_uuid"`
	Scopes    []string  `json:"scopes"`
	IsRefresh bool      `json:"is_refresh"`
	NeedOTP   bool      `json:"need_otp"`
}