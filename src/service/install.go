package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

func Install() error {
	// Set up logging
	f, err := os.OpenFile("install.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("error opening log file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	exepath, err := getServiceExePath()
	if err != nil {
		log.Printf("Failed to get service executable path: %v", err)
		return err
	}
	log.Printf("Service executable path: %s", exepath)

	m, err := mgr.Connect()
	if err != nil {
		log.Printf("Failed to connect to service manager: %v", err)
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(ServiceName)
	if err == nil {
		s.Close()
		log.Printf("Service %s already exists", ServiceName)
		return fmt.Errorf("service %s already exists", ServiceName)
	}
	log.Printf("Creating new service...")

	s, err = m.CreateService(ServiceName, exepath, mgr.Config{
		DisplayName: ServiceName,
		StartType:   mgr.StartAutomatic,
		Description: ServiceDesc,
	})
	if err != nil {
		log.Printf("Failed to create service: %v", err)
		return err
	}
	defer s.Close()
	log.Printf("Service created successfully")

	err = eventlog.InstallAsEventCreate(ServiceName, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		log.Printf("Failed to install eventlog: %v", err)
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}
	log.Printf("Eventlog installed successfully")

	return nil
}

func Remove() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(ServiceName)
	if err != nil {
		return fmt.Errorf("service %s is not installed", ServiceName)
	}
	defer s.Close()

	err = s.Delete()
	if err != nil {
		return err
	}

	err = eventlog.Remove(ServiceName)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}

	return nil
}

// getServiceExePath returns the path to the service executable
// This should be in the same directory as the GUI executable
func getServiceExePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	// Get the directory of the current executable
	dir := filepath.Dir(exePath)

	// The service executable should be named "windows-theme-autochanger-service.exe"
	servicePath := filepath.Join(dir, "windows-theme-autochanger-service.exe")

	return servicePath, nil
}
