package handler

import (
	"DIA_Backend/internal/app/repository"

	"github.com/gin-gonic/gin"
)

func RegisterHandlers(router *gin.Engine, repo *repository.Repository) {
	apiRouter := router.Group("/api")

	costHandler := NewCostHandler(repo)
	costRouter := apiRouter.Group("/costs")

	{
		costRouter.GET("", costHandler.GetCosts)
		costRouter.GET("/:id", costHandler.GetCostByID)
		costRouter.POST("", costHandler.CreateCost)
		costRouter.PUT("/:id", costHandler.UpdateCost)
		costRouter.DELETE("/:id", costHandler.DeleteCost)
		costRouter.POST("/:id/image", costHandler.AddCostImage)
		costRouter.POST("/:id/add-to-request", costHandler.AddCostToDraftRequest)
	}

	requestHandler := NewCostRequestHandler(repo)
	requestRouter := apiRouter.Group("/cost-requests")
	{
		requestRouter.GET("/costRequestInfo", requestHandler.GetCostRequestInfo)
		requestRouter.GET("", requestHandler.GetCostRequests)
		requestRouter.GET("/:id", requestHandler.GetCostRequestByID)
		requestRouter.PUT("/:id", requestHandler.UpdateCostRequest)
		requestRouter.PUT("/:id/form", requestHandler.FormCostRequest)
		requestRouter.PUT("/:id/resolve", requestHandler.ResolveCostRequest)
		requestRouter.PUT("/:id/reject", requestHandler.RejectCostRequest)
		requestRouter.DELETE("/:id", requestHandler.DeleteCostRequest)
	}

	requestCostHandler := NewPriceRequestToCostHandler(repo)
	requestCostRouter := apiRouter.Group("/cost-request-costs")
	{
		requestCostRouter.DELETE("/:requestId/costs/:costId", requestCostHandler.RemovePriceToRequestConnection)
		requestCostRouter.PUT("/:requestId/costs/:costId", requestCostHandler.UpdatePriceToRequestConnection)
	}

	userHandler := NewUserHandler(repo)
	userRouter := apiRouter.Group("/users")
	{
		userRouter.POST("/register", userHandler.Register)
		userRouter.GET("/profile", userHandler.GetProfile)
		userRouter.PUT("/profile", userHandler.UpdateProfile)
		userRouter.POST("/login", userHandler.Login)
		userRouter.POST("/logout", userHandler.Logout)
	}
}
