package main

import (
	"syscall"
	"unsafe"
)

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procFindWindow    = user32.NewProc("FindWindowW")
	procGetWindowLong = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLong = user32.NewProc("SetWindowLongPtrW")
	procShowWindow    = user32.NewProc("ShowWindow")
)

const (
	WS_EX_TOOLWINDOW = 0x00000080
	WS_EX_APPWINDOW  = 0x00040000
	SW_HIDE          = 0
	SW_SHOW          = 5
)

// HideFromTaskbar hides the window from the Windows taskbar
// by setting WS_EX_TOOLWINDOW style and removing WS_EX_APPWINDOW
func (a *App) HideFromTaskbar() {
	className, _ := syscall.UTF16PtrFromString("wailsWindow")
	hwnd, _, _ := procFindWindow.Call(
		uintptr(unsafe.Pointer(className)),
		0,
	)
	if hwnd == 0 {
		return
	}

	// Get current extended style (GWL_EXSTYLE = -20, use two's complement)
	gwlExStyle := uintptr(0xFFFFFFFFFFFFFFEC) // -20 as uintptr
	style, _, _ := procGetWindowLong.Call(hwnd, gwlExStyle)

	// Remove WS_EX_APPWINDOW, add WS_EX_TOOLWINDOW
	newStyle := (style &^ WS_EX_APPWINDOW) | WS_EX_TOOLWINDOW

	// Briefly hide the window to apply style change
	procShowWindow.Call(hwnd, SW_HIDE)
	procSetWindowLong.Call(hwnd, gwlExStyle, newStyle)
	procShowWindow.Call(hwnd, SW_SHOW)
}
