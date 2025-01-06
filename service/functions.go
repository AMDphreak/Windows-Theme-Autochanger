package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

type SunriseSunsetResponse struct {
	Results struct {
		Sunrise string `json:"sunrise"`
		Sunset  string `json:"sunset"`
	} `json:"results"`
}

// Cursor-related registry paths and keys
const (
	cursorSchemeKey    = `Control Panel\Cursors`
	cursorDefaultKey   = `Control Panel\Cursors\Schemes`
	cursorsRegistryKey = `HKEY_CURRENT_USER\` + cursorSchemeKey
)

// List of cursor types to preserve
var cursorTypes = []string{
	"Arrow",
	"Hand",
	"Help",
	"IBeam",
	"Wait",
	"Cross",
	"Crosshair",
	"NWPen",
	"No",
	"SizeNS",
	"SizeWE",
	"SizeNWSE",
	"SizeNESW",
	"SizeAll",
	"UpArrow",
	"AppStarting",
}

type CursorSettings struct {
	Cursors map[string]string
}

func getCurrentLocation() (float64, float64, error) {
	resp, err := http.Get("https://ipapi.co/json/")
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	var data struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return 0, 0, err
	}

	return data.Latitude, data.Longitude, nil
}

func getSunriseSunset(lat, lon float64) (time.Time, time.Time, error) {
	url := fmt.Sprintf("https://api.sunrise-sunset.org/json?lat=%f&lng=%f&formatted=0", lat, lon)
	resp, err := http.Get(url)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	var sunData SunriseSunsetResponse
	if err := json.Unmarshal(body, &sunData); err != nil {
		return time.Time{}, time.Time{}, err
	}

	sunrise, err := time.Parse(time.RFC3339, sunData.Results.Sunrise)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	sunset, err := time.Parse(time.RFC3339, sunData.Results.Sunset)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	// Adjust to local timezone
	local := time.Local
	sunrise = sunrise.In(local)
	sunset = sunset.In(local)

	return sunrise, sunset, nil
}

func getCurrentCursorSettings() (CursorSettings, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, cursorSchemeKey, registry.QUERY_VALUE)
	if err != nil {
		return CursorSettings{}, err
	}
	defer k.Close()

	cursors := make(map[string]string)
	for _, cursorType := range cursorTypes {
		value, _, err := k.GetStringValue(cursorType)
		if err == nil {
			cursors[cursorType] = value
		}
	}

	return CursorSettings{Cursors: cursors}, nil
}

func restoreCursorSettings(settings CursorSettings) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, cursorSchemeKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	for cursorType, path := range settings.Cursors {
		if err := k.SetStringValue(cursorType, path); err != nil {
			fmt.Printf("Error setting cursor %s: %v", cursorType, err)
		}
	}

	user32 := windows.NewLazySystemDLL("user32.dll")
	systemParametersInfo := user32.NewProc("SystemParametersInfoW")
	systemParametersInfo.Call(0x2029, 0, 0, 0) // SPI_SETCURSORS

	return nil
}

func SetWindowsTheme(isDark bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`,
		registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	var lightModeValue uint32 = 1
	if isDark {
		lightModeValue = 0
	}

	// Preserve current cursor settings before theme change
	currentCursors, err := getCurrentCursorSettings()
	if err != nil {
		return err
	}

	// Change theme
	if err := k.SetDWordValue("AppsUseLightTheme", lightModeValue); err != nil {
		return err
	}

	if err := k.SetDWordValue("SystemUsesLightTheme", lightModeValue); err != nil {
		return err
	}

	// Restore cursor settings after theme change
	time.Sleep(500 * time.Millisecond) // Give some time for theme to settle
	if err := restoreCursorSettings(currentCursors); err != nil {
		return err
	}

	return nil
}

func ResetThemeOverride() error {
	// Add any logic needed to reset to automatic mode
	return nil
}

func isFullScreenAppRunning() bool {
	user32 := windows.NewLazySystemDLL("user32.dll")

	getForegroundWindow := user32.NewProc("GetForegroundWindow")
	hwnd, _, _ := getForegroundWindow.Call()
	if hwnd == 0 {
		return false
	}

	getMonitorInfo := user32.NewProc("GetMonitorInfoW")
	monitorFromWindow := user32.NewProc("MonitorFromWindow")

	type monitorInfo struct {
		Size     uint32
		Monitor  struct{ Left, Top, Right, Bottom int32 }
		WorkArea struct{ Left, Top, Right, Bottom int32 }
		Flags    uint32
	}

	mi := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}

	monitorHandle, _, _ := monitorFromWindow.Call(hwnd, 0x00000002)
	if monitorHandle == 0 {
		return false
	}

	ret, _, _ := getMonitorInfo.Call(
		monitorHandle,
		uintptr(unsafe.Pointer(&mi)),
	)

	if ret == 0 {
		return false
	}

	type rect struct {
		Left, Top, Right, Bottom int32
	}
	getWindowRect := user32.NewProc("GetWindowRect")
	windowRect := rect{}
	getWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&windowRect)))

	monitorWidth := mi.Monitor.Right - mi.Monitor.Left
	monitorHeight := mi.Monitor.Bottom - mi.Monitor.Top
	windowWidth := windowRect.Right - windowRect.Left
	windowHeight := windowRect.Bottom - windowRect.Top

	return windowWidth >= monitorWidth-10 &&
		windowHeight >= monitorHeight-10
}
