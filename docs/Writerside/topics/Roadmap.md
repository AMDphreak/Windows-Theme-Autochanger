# Roadmap

## Existing Features:
1. Core Functionality:
    - Automatic theme switching based on sunrise/sunset times
    - Location detection using IP geolocation
    - Sunrise/sunset time calculation using external API
    - Support for both light and dark themes
2. Service Management:
    - Windows service installation/uninstallation 
    - Service start/stop controls
    - Automatic startup with Windows
    - Event logging
3. User Interface:
    - System tray integration with status indicators
    - Settings window with theme override options
    - Administrator privilege elevation when needed
4. Special Considerations:
    - Cursor scheme preservation during theme changes
    - Full-screen application detection (prevents theme switching during full-screen apps)
    - Logging system for both service and GUI components


## Potential Incomplete or Missing Items:
1. Error Handling & Recovery:
    - No retry mechanism for failed API calls (location and sunrise/sunset)
    - Limited error recovery for network failures
2. Configuration:
   - No user-configurable settings for:
   - Custom sunrise/sunset offsets
   - Update frequency
   - Custom theme schedules
   - Manual location override
3. UI Improvements:
   - Missing custom system tray icon (placeholder in getIcon())
   - No progress indicators during service operations
   - No real-time status updates in settings window
4. Installation:
   - WiX installer specification needs unique GUIDs
   - No upgrade path defined for future versions
   - No uninstall cleanup procedures
5. Security:
   - No signature verification for the executables
   - No encrypted communication between GUI and service
6. Testing:
   - No visible test suite
   - No error simulation handling
   - No integration tests
7. Documentation:
   - No user documentation
   - No API documentation
   - No deployment guide
8. Monitoring & Maintenance:
   - No health monitoring
   - No crash reporting
   - No automatic updates system

## Release Channels
- Download is available on website. Install via installer in Windows.
- Aspiring to release on Windows App Store as well. <a href="Windows-App-Store.md">Details</a>