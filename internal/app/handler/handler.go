package handler

import (
	"DIA_Backend/internal/app/repository"

	"github.com/gin-gonic/gin"
)

func RegisterHandlers(router *gin.Engine, repo *repository.Repository) {
	apiRouter := router.Group("/api")
	userHandler := NewUserHandler(repo)

	publicRouter := apiRouter.Group("")
	{
		publicRouter.POST("/users/register", userHandler.Register)
		publicRouter.POST("/users/login", userHandler.Login)
		publicRouter.POST("/users/refresh", userHandler.RefreshToken)

		// Public costs routes
		costHandler := NewCostHandler(repo)
		publicRouter.GET("/costs", costHandler.GetCosts)
		publicRouter.GET("/costs/:id", costHandler.GetCostByID)

		requestHandler := NewCostRequestHandler(repo)
		publicRouter.GET("/cost-requests/costRequestInfo", requestHandler.GetCostRequestInfo)
	}

	protectedRouter := apiRouter.Group("")
	protectedRouter.Use(userHandler.AuthMiddleware())
	{
		protectedRouter.GET("/users/profile", userHandler.GetProfile)
		protectedRouter.PUT("/users/profile", userHandler.UpdateProfile)
		protectedRouter.POST("/users/logout", userHandler.Logout)

		// Cost routes with role-based permissions
		costHandler := NewCostHandler(repo)
		costRouter := protectedRouter.Group("/costs")
		{

			costRouter.POST("", userHandler.ScopeMiddleware("create:costs"), costHandler.CreateCost)
			costRouter.PUT("/:id", userHandler.ScopeMiddleware("update:costs"), costHandler.UpdateCost)
			costRouter.DELETE("/:id", userHandler.ScopeMiddleware("delete:costs"), costHandler.DeleteCost)
			costRouter.POST("/:id/image", userHandler.ScopeMiddleware("update:costs"), costHandler.AddCostImage)
			costRouter.POST("/:id/add-to-request", userHandler.ScopeMiddleware("create:requests"), costHandler.AddCostToDraftRequest)
		}
	}

	requestHandler := NewCostRequestHandler(repo)
	requestRouter := protectedRouter.Group("/cost-requests")
	{
		requestRouter.GET("", requestHandler.GetCostRequests)
		requestRouter.GET("/:id", requestHandler.GetCostRequestByID)
		requestRouter.PUT("/:id", userHandler.ScopeMiddleware("update:requests"), requestHandler.UpdateCostRequest)
		requestRouter.PUT("/:id/form", userHandler.ScopeMiddleware("update:requests"), requestHandler.FormCostRequest)
		requestRouter.PUT("/:id/resolve", userHandler.ScopeMiddleware("resolve:requests"), requestHandler.ResolveCostRequest)
		requestRouter.PUT("/:id/reject", userHandler.ScopeMiddleware("reject:requests"), requestHandler.RejectCostRequest)
		requestRouter.DELETE("/:id", userHandler.ScopeMiddleware("update:requests"), requestHandler.DeleteCostRequest)
	}

	requestCostHandler := NewPriceRequestToCostHandler(repo)
	requestCostRouter := protectedRouter.Group("/cost-request-costs")
	{
		requestCostRouter.DELETE("/:requestId/costs/:costId", userHandler.ScopeMiddleware("update:requests"), requestCostHandler.RemovePriceToRequestConnection)
		requestCostRouter.PUT("/:requestId/costs/:costId", userHandler.ScopeMiddleware("update:requests"), requestCostHandler.UpdatePriceToRequestConnection)
	}
}
