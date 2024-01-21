package main

import (
	"ASO/main/crypt"
	"ASO/main/database"
	"ASO/main/router"
	"ASO/main/tasks"
	"fmt"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	fmt.Println("Arnolds Super Organiser")

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
		log.Println("Please create .env file with the following content:")
		log.Println("MONGODB_URI=mongodb://localhost:27017")
		return
	}

	err := crypt.KeySetup()
	if err != nil {
		log.Println("Failed to setup keys")
		return
	}

	ok := database.InitDatabase()
	if ok {
		log.Println("Database connection established")
	} else {
		log.Println("Database connection failed exiting...")
		return
	}

	tasks.StartRepeatingTasks()

	router.InitRouter()
}
