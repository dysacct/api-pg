package models

import "gorm.io/gorm"

type Contact struct {
	gorm.Model

	FirstName  string `json:"first_name" gorm:"not null"`
	SecondName string `json:"second_name" gorm:"not null"`
	Email      string `json:"email" gorm:"not null"`
	Phone      string `json:"phone" gorm:"not null"`
}
