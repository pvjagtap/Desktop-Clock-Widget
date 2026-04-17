package main

import (
	"context"
	"os"
	"path/filepath"
	"time"
)

// App struct
type App struct {
	ctx         context.Context
	settingsDir string
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Settings directory: %APPDATA%\ClockWidget
	appData, err := os.UserConfigDir()
	if err != nil {
		appData = "."
	}
	dir := filepath.Join(appData, "ClockWidget")
	os.MkdirAll(dir, 0755)

	return &App{
		settingsDir: dir,
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Setup system tray icon
	a.setupTray()
	// Hide from taskbar (Windows only)
	go func() {
		// Small delay to ensure window is created
		time.Sleep(500 * time.Millisecond)
		a.HideFromTaskbar()
	}()
}

// settingsPath returns the path to the settings JSON file
func (a *App) settingsPath() string {
	return filepath.Join(a.settingsDir, "settings.json")
}

// GetSettings reads saved settings from disk
func (a *App) GetSettings() (string, error) {
	data, err := os.ReadFile(a.settingsPath())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveSettings writes settings JSON to disk
func (a *App) SaveSettings(jsonStr string) error {
	return os.WriteFile(a.settingsPath(), []byte(jsonStr), 0644)
}

// SetOpacity sets the window opacity (0-100) via CSS
func (a *App) SetOpacity(percent int) {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	// Opacity is handled on the frontend via CSS
	// This method exists so it's accessible via Wails binding
}
