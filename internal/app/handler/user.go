package handler

import (
	"DIA_Backend/internal/app/ds"
	"DIA_Backend/internal/app/repository"
	"DIA_Backend/internal/app/role"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

var (
	fixedUserID uint64 = 1
)

func GetFixedUserID() uint64 {
	return fixedUserID
}

type UserHandler struct {
	repo *repository.Repository
}

func NewUserHandler(repo *repository.Repository) *UserHandler {
	return &UserHandler{
		repo: repo,
	}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int64    `json:"expires_in"`
	User         UserInfo `json:"user"`
}

type UserInfo struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type UpdateProfileRequest struct {
	Username *string `json:"username"`
	Password *string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// @Summary      Register a new user
// @Description  Create a new user account
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "User registration data"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /users/register [post]
func (h *UserHandler) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	userRole := role.FromString(req.Role)
	if req.Role != "" && (req.Role != "user" && req.Role != "moderator") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be 'user' or 'moderator'"})
		return
	}
	user := &ds.User{
		Username: req.Username,
		Password: req.Password,
		Role:     userRole,
	}

	if err := h.repo.User.CreateUser(user); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"id_user": user.ID,
	})
}

// @Summary      User login
// @Description  Authenticate user and return JWT tokens
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "User credentials"
// @Success      200  {object}  LoginResponse
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /users/login [post]
func (h *UserHandler) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	user, err := h.repo.User.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	jwtConfig := h.repo.GetJWTConfig()
	accessToken, refreshToken, err := h.generateTokens(user, jwtConfig)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	response := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(jwtConfig.ExpiresIn.Seconds()),
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role.String(),
		},
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary      Refresh access token
// @Description  Get new access token using refresh token
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body RefreshTokenRequest true "Refresh token"
// @Success      200  {object}  LoginResponse
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /users/refresh [post]
func (h *UserHandler) RefreshToken(ctx *gin.Context) {
	var req RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	claims := &ds.JWTClaims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.repo.GetJWTSecret()), nil
	})

	if err != nil || !token.Valid {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	redisClient := h.repo.GetRedisClient()
	if redisClient != nil {
		isBlacklisted, err := redisClient.CheckJWTInBlacklist(context.Background(), req.RefreshToken)
		if err != nil {
			logrus.WithError(err).Error("Failed to check Redis blacklist for refresh token")
		} else if isBlacklisted {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token has been revoked"})
			return
		}
	}

	if !claims.IsRefresh {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Not a refresh token"})
		return
	}

	user, err := h.repo.User.GetUserByUUID(claims.UserUUID.String())
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	jwtConfig := h.repo.GetJWTConfig()
	accessToken, refreshToken, err := h.generateTokens(user, jwtConfig)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	response := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(jwtConfig.ExpiresIn.Seconds()),
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role.String(),
		},
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary      Get user profile
// @Description  Get current user's profile information
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /users/profile [get]
func (h *UserHandler) GetProfile(ctx *gin.Context) {
	userUUID, _, ok := GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := h.repo.User.GetUserByUUID(userUUID)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	profile := gin.H{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role.String(),
	}

	ctx.JSON(http.StatusOK, profile)
}

// @Summary      Update user profile
// @Description  Update current user's profile information
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body UpdateProfileRequest true "Profile update data"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /users/profile [put]
func (h *UserHandler) UpdateProfile(ctx *gin.Context) {
	var req UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	userUUID, _, ok := GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := h.repo.User.UpdateUserByUUID(userUUID, req.Username, req.Password); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// @Summary      User logout
// @Description  Logout user (token invalidation)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /users/logout [post]
func (h *UserHandler) Logout(ctx *gin.Context) {
	jwtStr := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header required"})
		return
	}

	jwtStr = jwtStr[len(jwtPrefix):]

	claims := &ds.JWTClaims{}
	token, err := jwt.ParseWithClaims(jwtStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.repo.GetJWTSecret()), nil
	})

	if err != nil || !token.Valid {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		return
	}

	// Add access token to Redis blacklist
	redisClient := h.repo.GetRedisClient()
	if redisClient != nil {
		err = redisClient.WriteJWTToBlacklist(context.Background(), jwtStr, claims.ExpiresAt.Time)
		if err != nil {
			logrus.Error("Failed to add access token to blacklist: ", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func (h *UserHandler) generateTokens(user *ds.User, jwtConfig *repository.JWTConfig) (string, string, error) {
	accessToken := jwt.NewWithClaims(jwtConfig.SigningMethod, &ds.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtConfig.ExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "DIA_Backend",
		},
		UserUUID:  user.UUID,
		Scopes:    user.Role.GetScopes(),
		IsRefresh: false,
		NeedOTP:   false,
	})

	accessTokenStr, err := accessToken.SignedString([]byte(jwtConfig.Secret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %v", err)
	}

	refreshToken := jwt.NewWithClaims(jwtConfig.SigningMethod, &ds.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtConfig.RefreshIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "DIA_Backend",
		},
		UserUUID:  user.UUID,
		Scopes:    []string{"refresh"},
		IsRefresh: true,
		NeedOTP:   false,
	})

	refreshTokenStr, err := refreshToken.SignedString([]byte(jwtConfig.Secret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %v", err)
	}

	return accessTokenStr, refreshTokenStr, nil
}
