package main

import (
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	shell32               = syscall.NewLazyDLL("shell32.dll")
	trayKernel32          = syscall.NewLazyDLL("kernel32.dll")
	procShellNotifyIcon   = shell32.NewProc("Shell_NotifyIconW")
	procLoadImage         = user32.NewProc("LoadImageW")
	procCreatePopupMenu   = user32.NewProc("CreatePopupMenu")
	procAppendMenu        = user32.NewProc("AppendMenuW")
	procTrackPopupMenu    = user32.NewProc("TrackPopupMenu")
	procDestroyMenu       = user32.NewProc("DestroyMenu")
	procCreateWindowEx    = user32.NewProc("CreateWindowExW")
	procDefWindowProc     = user32.NewProc("DefWindowProcW")
	procRegisterClass     = user32.NewProc("RegisterClassExW")
	procGetMessage        = user32.NewProc("GetMessageW")
	procTranslateMessage  = user32.NewProc("TranslateMessage")
	procDispatchMessage   = user32.NewProc("DispatchMessageW")
	procPostQuitMessage   = user32.NewProc("PostQuitMessage")
	procSetForegroundWnd  = user32.NewProc("SetForegroundWindow")
	procGetCursorPos      = user32.NewProc("GetCursorPos")
	procExtractIconEx     = shell32.NewProc("ExtractIconExW")
	procGetModuleFileName = trayKernel32.NewProc("GetModuleFileNameW")
)

const (
	NIM_ADD         = 0x00000000
	NIM_MODIFY      = 0x00000001
	NIM_DELETE      = 0x00000002
	NIF_MESSAGE     = 0x00000001
	NIF_ICON        = 0x00000002
	NIF_TIP         = 0x00000004
	WM_APP          = 0x8000
	WM_TRAYICON     = WM_APP + 1
	WM_COMMAND      = 0x0111
	WM_LBUTTONUP    = 0x0202
	WM_RBUTTONUP    = 0x0205
	MF_STRING       = 0x0000
	MF_SEPARATOR    = 0x0800
	TPM_BOTTOMALIGN = 0x0020
	TPM_LEFTALIGN   = 0x0000
	IDM_SHOW        = 1001
	IDM_QUIT        = 1002
	IMAGE_ICON      = 1
	LR_LOADFROMFILE = 0x0010
	LR_DEFAULTSIZE  = 0x0040
)

type NOTIFYICONDATA struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
}

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  uintptr
	LpszClassName uintptr
	HIconSm       uintptr
}

type MSG struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type POINT struct {
	X, Y int32
}

var (
	trayHwnd uintptr
	nid      NOTIFYICONDATA
	appRef   *App
)

func trayWndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_TRAYICON:
		switch lParam {
		case WM_LBUTTONUP:
			if appRef != nil {
				runtime.WindowShow(appRef.ctx)
			}
		case WM_RBUTTONUP:
			showTrayMenu(hwnd)
		}
		return 0
	case WM_COMMAND:
		switch wParam {
		case IDM_SHOW:
			if appRef != nil {
				runtime.WindowShow(appRef.ctx)
			}
		case IDM_QUIT:
			removeTrayIcon()
			if appRef != nil {
				runtime.Quit(appRef.ctx)
			}
		}
		return 0
	}
	ret, _, _ := procDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func showTrayMenu(hwnd uintptr) {
	hMenu, _, _ := procCreatePopupMenu.Call()
	showStr, _ := syscall.UTF16PtrFromString("Show/Hide")
	quitStr, _ := syscall.UTF16PtrFromString("Quit")

	procAppendMenu.Call(hMenu, MF_STRING, IDM_SHOW, uintptr(unsafe.Pointer(showStr)))
	procAppendMenu.Call(hMenu, MF_SEPARATOR, 0, 0)
	procAppendMenu.Call(hMenu, MF_STRING, IDM_QUIT, uintptr(unsafe.Pointer(quitStr)))

	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	procSetForegroundWnd.Call(hwnd)
	procTrackPopupMenu.Call(hMenu, TPM_LEFTALIGN|TPM_BOTTOMALIGN, uintptr(pt.X), uintptr(pt.Y), 0, hwnd, 0)
	procDestroyMenu.Call(hMenu)
}

func (a *App) setupTray() {
	appRef = a
	go func() {
		// Register window class
		className, _ := syscall.UTF16PtrFromString("ClockWidgetTray")
		getModuleHandle := trayKernel32.NewProc("GetModuleHandleW")
		hInstance, _, _ := getModuleHandle.Call(0)

		wc := WNDCLASSEX{
			CbSize:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
			LpfnWndProc:   syscall.NewCallback(trayWndProc),
			HInstance:     hInstance,
			LpszClassName: uintptr(unsafe.Pointer(className)),
		}
		procRegisterClass.Call(uintptr(unsafe.Pointer(&wc)))

		// Create hidden message window
		hwnd, _, _ := procCreateWindowEx.Call(
			0,
			uintptr(unsafe.Pointer(className)),
			0,
			0,
			0, 0, 0, 0,
			0, 0, hInstance, 0,
		)
		trayHwnd = hwnd

		// Load icon from our own exe
		var exeBuf [260]uint16
		procGetModuleFileName.Call(0, uintptr(unsafe.Pointer(&exeBuf[0])), 260)
		var hIconSmall uintptr
		procExtractIconEx.Call(uintptr(unsafe.Pointer(&exeBuf[0])), 0, 0, uintptr(unsafe.Pointer(&hIconSmall)), 1)

		// Setup NOTIFYICONDATA
		nid = NOTIFYICONDATA{
			CbSize:           uint32(unsafe.Sizeof(NOTIFYICONDATA{})),
			HWnd:             hwnd,
			UID:              1,
			UFlags:           NIF_MESSAGE | NIF_ICON | NIF_TIP,
			UCallbackMessage: WM_TRAYICON,
			HIcon:            hIconSmall,
		}
		tip := syscall.StringToUTF16("Clock Widget")
		copy(nid.SzTip[:], tip)

		procShellNotifyIcon.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))

		// Message loop
		var msg MSG
		for {
			ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
			if ret == 0 {
				break
			}
			procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
		}
	}()
}

func removeTrayIcon() {
	procShellNotifyIcon.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
	procPostQuitMessage.Call(0)
}
