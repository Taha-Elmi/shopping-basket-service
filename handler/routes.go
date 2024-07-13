package handler

import "github.com/labstack/echo/v4"

func RegisterRoutes(e *echo.Echo) {
	// Public routes
	e.POST("/signup", SignupHandler)
	e.POST("/login", LoginHandler)

	// Authenticated routes
	authRouter := e.Group("/basket")
	authRouter.Use(AuthMiddleware)
	authRouter.GET("", GetBasketsHandler)
	authRouter.POST("", CreateBasketHandler)
	authRouter.PATCH("/:id", UpdateBasketHandler)
	authRouter.GET("/:id", GetBasketHandler)
	authRouter.DELETE("/:id", DeleteBasketHandler)
}
