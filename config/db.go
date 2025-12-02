package config

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDb() {
	var err error
	fmt.Println("Connecting db...")

	dbURL := fmt.Sprintf("")
	DB, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: " + err.Error())
	}

	// TODO: add migration
	fmt.Println("Connected to database")
}
