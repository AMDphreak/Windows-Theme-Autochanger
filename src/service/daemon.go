package service

import (
	"fmt"
	"time"
)

func (s *ThemeService) runThemeDaemon() {
	// Get current location
	lat, lon, err := getCurrentLocation()
	if err != nil {
		s.elog.Error(1, fmt.Sprintf("Failed to get location: %v", err))
		return
	}

	s.elog.Info(1, fmt.Sprintf("Location detected: Latitude %.4f, Longitude %.4f", lat, lon))

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.cancel:
			return
		case <-ticker.C:
			// Get sunrise and sunset times
			sunrise, sunset, err := getSunriseSunset(lat, lon)
			if err != nil {
				s.elog.Error(1, fmt.Sprintf("Error getting sunrise/sunset: %v", err))
				continue
			}

			now := time.Now()
			isDarkMode := now.Before(sunrise) || now.After(sunset)

			if !isFullScreenAppRunning() {
				err = SetWindowsTheme(isDarkMode)
				if err != nil {
					s.elog.Error(1, fmt.Sprintf("Theme setting error: %v", err))
				} else {
					s.elog.Info(1, fmt.Sprintf("Set theme to %s mode", map[bool]string{true: "dark", false: "light"}[isDarkMode]))
				}
			} else {
				s.elog.Info(1, "Full-screen app running. Theme change postponed.")
			}
		}
	}
}
