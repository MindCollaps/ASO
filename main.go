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
	"os"
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
	flags()

	logChopper.LogChop()

	log.Println("\n" + env.BANNER + "\nArnolds Super Organiser" + "\nVersion: " + env.VERSION)

	envSetup()

	cryptSetup()

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

func envSetup() {
	if !env.DOCKER {
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

		if os.Getenv("PORT") != "" && !isFlagPassed("port") {
			env.PORT = os.Getenv("PORT")
		}
	}
}

func flags() {
	flag.BoolVar(&env.UNIX, "unix", false, "Run the server in unix mode")
	flag.BoolVar(&env.DOCKER, "docker", false, "Run the server in docker mode")
	flag.StringVar(&env.PORT, "port", "8080", "Port to run the server on")

	flag.Parse()
}

func cryptSetup() {
	err := crypt.KeySetup()
	if err != nil {
		log.Println("Failed to setup keys")
		return
	}
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
