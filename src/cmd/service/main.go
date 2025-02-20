package main

import (
	"log"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/svc"
	"windows-theme-autochanger/service"
)

func main() {
	// Set up logging
	logPath := filepath.Join(filepath.Dir(os.Args[0]), "service.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// Run as service
	svcHandler, err := service.New()
	if err != nil {
		log.Fatalf("Failed to create service handler: %v", err)
	}

	err = svc.Run(service.ServiceName, svcHandler)
	if err != nil {
		log.Fatalf("Service failed: %v", err)
	}
}