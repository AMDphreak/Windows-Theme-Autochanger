# Installer

We have a WiX (Windows Installer XML) installer specification defined in installer-specification.wxs. WiX is a toolset that builds Windows installation packages from XML source code.

## Current WiX specification:
1. Creates an installer for "Windows Theme Autochanger"
2. Requires elevated (administrator) privileges
3. Installs to Program Files
4. Registers and configures the Windows service
5. Handles service control during install/uninstall

However, there are a few incomplete items in our WiX specification:
1. Missing unique GUIDs (placeholders: "PUT-UNIQUE-GUID-HERE")
2. Only includes the main executable, missing the service executable
3. No Start Menu shortcuts
4. No registry entries
5. No custom installation options
