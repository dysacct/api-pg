package handlers

import (
	"api-postgre/config"
	"api-postgre/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func FetchALLContacts(c *gin.Context) {
	contacts := []models.Contact{}

	if err := config.DB.Find(&contacts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error fetching contacts: %v", err.Error()),
		})
		return // 作用是退出函数,防止继续执行下面的代码。
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":  "Fetched contacts successfully!",
		"data": contacts,
	})
}

func CreateContact(c *gin.Context) {
	var contact models.Contact
	if err := c.ShouldBindBodyWithJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Error binding body: %v", err.Error()),
		})
		return
	}

	if err := config.DB.Create(&contact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error creating contact: %v", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":  "Contact created successfully!",
		"data": contact,
	})
}

func FetchContact(c *gin.Context) {

}

func DeleteContact(c *gin.Context) {

}
func UpdateContact(c *gin.Context) {

}
