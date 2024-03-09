package logChopper

import (
	"ASOServer/main/env"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"os"
)

func LogChop() {
	filePath := "aso.log"
	if env.UNIX {
		filePath = "/var/log/aso/aso.log"
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	// Create a multi-writer that writes to both os.Stdout and the log file
	mw := io.MultiWriter(os.Stdout, file)

	// Set log output to the multi-writer
	log.SetOutput(mw)

	// Set Gin's debug mode and writer
	gin.SetMode(gin.DebugMode)
	gin.DefaultWriter = mw
}
