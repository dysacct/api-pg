package util

import (
	"api-postgre/config"
	"api-postgre/models"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetContactById(c *gin.Context, contact *models.Contact) (int, int, error) {
	id, err := strconv.Atoi(c.Param("contactId"))

	if err != nil {
		return 400, 0, errors.New("invalid contact id")
	}

	if err := config.DB.Find(contact).Error; err != nil {

	}
}
