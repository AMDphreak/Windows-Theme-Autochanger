package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/getlantern/systray"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"windows-theme-autochanger/service"
)

var mw *walk.MainWindow
var statusLabel *walk.Label
var nextChangeLabel *walk.Label

func main() {
	// Set up logging
	logPath := filepath.Join(filepath.Dir(os.Args[0]), "theme-autochanger.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		showError("Failed to open log file", err)
		return
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// Start systray
	go systray.Run(onReady, onExit)

	// Create and run the main window (hidden initially)
	mainWindow, err := walk.NewMainWindow()
	if err != nil {
		log.Fatal(err)
	}

	// Run() will block until the window is closed
	mainWindow.Run()
}

func onReady() {
	systray.SetIcon(getIcon())
	systray.SetTitle("Theme Autochanger")
	systray.SetTooltip("Windows Theme Autochanger")

	mStatus := systray.AddMenuItem("Status: Checking...", "Service status")
	mStatus.Disable()

	systray.AddSeparator()

	mInstall := systray.AddMenuItem("Install Service", "Install the auto-changer service")
	mUninstall := systray.AddMenuItem("Uninstall Service", "Remove the auto-changer service")

	systray.AddSeparator()

	mStart := systray.AddMenuItem("Start Service", "Start the auto-changer service")
	mStop := systray.AddMenuItem("Stop Service", "Stop the auto-changer service")

	systray.AddSeparator()

	mSettings := systray.AddMenuItem("Settings", "Open settings window")

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	// Update service status periodically
	go func() {
		for {
			isInstalled, isRunning := getServiceStatus()
			updateMenuItems(mInstall, mUninstall, mStart, mStop, isInstalled, isRunning)
			updateStatusMenuItem(mStatus, isInstalled, isRunning)
			time.Sleep(5 * time.Second)
		}
	}()

	// Handle menu items
	go func() {
		for {
			select {
			case <-mInstall.ClickedCh:
				handleServiceInstallation()
			case <-mUninstall.ClickedCh:
				handleServiceRemoval()
			case <-mStart.ClickedCh:
				handleServiceStart()
			case <-mStop.ClickedCh:
				handleServiceStop()
			case <-mSettings.ClickedCh:
				showSettingsWindow()
			case <-mQuit.ClickedCh:
				systray.Quit()
				walk.App().Exit(0)
				return
			}
		}
	}()
}

func onExit() {
	// Cleanup
}

func showSettingsWindow() {
	if mw != nil {
		mw.Show()
		return
	}

	var defaultRB, lightRB, darkRB *walk.RadioButton

	if err := (declarative.MainWindow{
		AssignTo: &mw,
		Title:    "Theme Autochanger Settings",
		MinSize:  declarative.Size{Width: 300, Height: 200},
		Layout:   declarative.VBox{},
		Children: []declarative.Widget{
			declarative.GroupBox{
				Title:  "Theme Selection",
				Layout: declarative.VBox{},
				Children: []declarative.Widget{
					declarative.RadioButton{
						AssignTo: &defaultRB,
						Text:     "System Default (Auto)",
						OnClicked: func() {
							// Re-enable automatic switching
							service.ResetThemeOverride()
						},
					},
					declarative.RadioButton{
						AssignTo: &lightRB,
						Text:     "Light",
						OnClicked: func() {
							if err := service.SetWindowsTheme(false); err != nil {
								showError("Error", err)
							}
						},
					},
					declarative.RadioButton{
						AssignTo: &darkRB,
						Text:     "Dark",
						OnClicked: func() {
							if err := service.SetWindowsTheme(true); err != nil {
								showError("Error", err)
							}
						},
					},
				},
			},
			declarative.Label{
				AssignTo: &statusLabel,
				Text:     "Service Status: Checking...",
			},
			declarative.Label{
				AssignTo: &nextChangeLabel,
				Text:     "Next theme change: Calculating...",
			},
		},
	}.Create()); err != nil {
		log.Printf("Error creating settings window: %v", err)
		return
	}

	defaultRB.SetChecked(true)
	mw.Show()
}

func getServiceStatus() (bool, bool) {
	m, err := mgr.Connect()
	if err != nil {
		return false, false
	}
	defer m.Disconnect()

	s, err := m.OpenService(service.ServiceName)
	if err != nil {
		return false, false // Service not installed
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return true, false // Service installed but can't query status
	}

	return true, status.State == svc.Running
}

func updateMenuItems(mInstall, mUninstall, mStart, mStop *systray.MenuItem, isInstalled, isRunning bool) {
	if isInstalled {
		mInstall.Hide()
		mUninstall.Show()
		if isRunning {
			mStart.Hide()
			mStop.Show()
		} else {
			mStart.Show()
			mStop.Hide()
		}
	} else {
		mInstall.Show()
		mUninstall.Hide()
		mStart.Hide()
		mStop.Hide()
	}
}

func updateStatusMenuItem(mStatus *systray.MenuItem, isInstalled, isRunning bool) {
	if !isInstalled {
		mStatus.SetTitle("Status: Service not installed")
	} else if isRunning {
		mStatus.SetTitle("Status: Service is running")
	} else {
		mStatus.SetTitle("Status: Service is stopped")
	}
}

// Service management functions
func handleServiceInstallation() {
	if !isAdmin() {
		// Request elevation and restart
		err := runMeElevated()
		if err != nil {
			showError("Failed to get administrator privileges", err)
		}
		return
	}

	err := service.Install()
	if err != nil {
		showError("Failed to install service", err)
		return
	}

	showInfo("Success", "Service installed successfully")
}

func handleServiceRemoval() {
	if !isAdmin() {
		err := runMeElevated()
		if err != nil {
			showError("Failed to get administrator privileges", err)
		}
		return
	}

	err := service.Remove()
	if err != nil {
		showError("Failed to remove service", err)
		return
	}

	showInfo("Success", "Service removed successfully")
}

func handleServiceStart() {
	m, err := mgr.Connect()
	if err != nil {
		showError("Failed to connect to service manager", err)
		return
	}
	defer m.Disconnect()

	s, err := m.OpenService(service.ServiceName)
	if err != nil {
		showError("Failed to open service", err)
		return
	}
	defer s.Close()

	err = s.Start()
	if err != nil {
		showError("Failed to start service", err)
		return
	}

	showInfo("Success", "Service started successfully")
}

func handleServiceStop() {
	m, err := mgr.Connect()
	if err != nil {
		showError("Failed to connect to service manager", err)
		return
	}
	defer m.Disconnect()

	s, err := m.OpenService(service.ServiceName)
	if err != nil {
		showError("Failed to open service", err)
		return
	}
	defer s.Close()

	status, err := s.Control(svc.Stop)
	if err != nil {
		showError("Failed to stop service", err)
		return
	}
	_ = status

	showInfo("Success", "Service stopped successfully")
}

func showError(title string, err error) {
	walk.MsgBox(nil, title, err.Error(), walk.MsgBoxIconError)
}

func showInfo(title, message string) {
	walk.MsgBox(nil, title, message, walk.MsgBoxIconInformation)
}

func getIcon() []byte {
	// Return a basic icon - you should replace this with your own icon
	return []byte{
		// ... icon data ...
	}
}
