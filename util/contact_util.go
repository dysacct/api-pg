package util

import (
	"api-postgre/config"
	"api-postgre/models"
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetContactById(c *gin.Context, contact *models.Contact) (int, int, error) {
	id, err := strconv.Atoi(c.Param("contactId"))

	if err != nil {
		return 400, 0, errors.New("invalid contact id")
	}

	if err := config.DB.Find(contact).Error; err != nil {
		return 500, 0, errors.New(fmt.Sprintf("error fetching contact: %v", err.Error()))
	}

	if contact.Model.ID == 0 {
		return 404, 0, errors.New("contact not found")
	}

	return 200, id, nil
}
