package handler

import (
	"DIA_Backend/internal/app/ds"
	"DIA_Backend/internal/app/repository"
	"net/http"

	"github.com/gin-gonic/gin"
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
	IsMod    bool   `json:"is_mod"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Username *string `json:"username"`
	Password *string `json:"password"`
}

func (h *UserHandler) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	user := &ds.User{
		Username: req.Username,
		Password: req.Password,
		IsMod:    req.IsMod,
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

func (h *UserHandler) GetProfile(ctx *gin.Context) {
	userID := GetFixedUserID()

	user, err := h.repo.User.GetUserByID(userID)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	profile := gin.H{
		"id":       user.ID,
		"username": user.Username,
	}

	ctx.JSON(http.StatusOK, profile)
}

func (h *UserHandler) UpdateProfile(ctx *gin.Context) {
	var req UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	userID := GetFixedUserID()

	if err := h.repo.User.UpdateUser(userID, req.Username, req.Password); err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

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

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"id_user": user.ID,
		"is_mod":  user.IsMod,
	})
}

func (h *UserHandler) Logout(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}
