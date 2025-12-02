package routes

import "github.com/gin-gonic/gin"

func RegisterRoutes(app *gin.Engine) {
	api := app.Group("/api")

	ContactRoutes(api)
}
