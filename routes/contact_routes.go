package routes

import (
	"api-postgre/handlers"

	"github.com/gin-gonic/gin"
)

func ContactRoutes(api *gin.RouterGroup) {

	router := api.Group("/contacts")

	router.GET("/", handlers.FetchALLContacts)
	router.POST("/", handlers.CreateContact)
	router.GET("/:contactId", handlers.FetchContact)
	router.DELETE("/:contactId", handlers.DeleteContact)
	router.PUT("/:contactId", handlers.UpdateContact)
}
