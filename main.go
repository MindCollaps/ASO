package main

import (
	"ASOServer/main/crypt"
	"ASOServer/main/database"
	"ASOServer/main/env"
	"ASOServer/main/logChopper"
	"ASOServer/main/router"
	"ASOServer/main/tasks"
	"embed"
	"flag"
	"github.com/joho/godotenv"
	"log"
)

//go:embed main/public/*
var Files embed.FS

func main() {
	_, err := Files.ReadDir("main/public")
	if err != nil {
		log.Println("Failed to read public files - this is likely a problem during compilation. Exiting...")
		return
	}
	// command line arguments
	flag.BoolVar(&env.UNIX, "unix", false, "Run the server in unix mode")
	flag.Parse()

	logChopper.LogChop()

	log.Println("\n" + env.BANNER + "\nArnolds Super Organiser" + "\nVersion: " + env.VERSION + "\n\n")

	envLocation := ".env"
	if env.UNIX {
		envLocation = "/etc/aso/.env"
	}
	if err := godotenv.Load(envLocation); err != nil {
		log.Println("No .env file found")
		log.Println("Please create .env file with the following content:")
		log.Println("MONGODB_URI=mongodb://localhost:27017")
		log.Println("MONGODB_DB=ASO")
		log.Println("\noptional: ")
		log.Println("PORT=8080")
		return
	}

	err = crypt.KeySetup()
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

	router.InitRouter(Files)
}
