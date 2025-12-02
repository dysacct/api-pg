package main

import (
	"api-postgre/config"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	PORT := os.Getenv("PORT")

	config.ConnectDB()
	app := gin.Default()

	app.Run(fmt.Sprintf(":%s", PORT))
}
