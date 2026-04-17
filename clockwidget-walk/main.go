package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

//go:embed fonts/DSEG7Classic-Bold.ttf
var dseg7Font []byte

// Ensure embed is used
var _ embed.FS

// ═══════════════════════════════════════════════════════════════════════════════
// Win32 DLLs and Procs
// ═══════════════════════════════════════════════════════════════════════════════

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")

	// user32
	pRegisterClassEx        = user32.NewProc("RegisterClassExW")
	pCreateWindowEx         = user32.NewProc("CreateWindowExW")
	pDefWindowProc          = user32.NewProc("DefWindowProcW")
	pGetMessage             = user32.NewProc("GetMessageW")
	pTranslateMessage       = user32.NewProc("TranslateMessage")
	pDispatchMessage        = user32.NewProc("DispatchMessageW")
	pPostQuitMessage        = user32.NewProc("PostQuitMessage")
	pShowWindow             = user32.NewProc("ShowWindow")
	pSetWindowPos           = user32.NewProc("SetWindowPos")
	pGetWindowRect          = user32.NewProc("GetWindowRect")
	pGetClientRect          = user32.NewProc("GetClientRect")
	pSetCapture             = user32.NewProc("SetCapture")
	pReleaseCapture         = user32.NewProc("ReleaseCapture")
	pGetCursorPos           = user32.NewProc("GetCursorPos")
	pSetForegroundWindow    = user32.NewProc("SetForegroundWindow")
	pSetLayeredWindowAttrib = user32.NewProc("SetLayeredWindowAttributes")
	pGetWindowLong          = user32.NewProc("GetWindowLongPtrW")
	pSetWindowLong          = user32.NewProc("SetWindowLongPtrW")
	pBeginPaint             = user32.NewProc("BeginPaint")
	pEndPaint               = user32.NewProc("EndPaint")
	pInvalidateRect         = user32.NewProc("InvalidateRect")
	pFillRect               = user32.NewProc("FillRect")
	pLoadCursor             = user32.NewProc("LoadCursorW")
	pLoadIcon               = user32.NewProc("LoadIconW")
	pSetTimer               = user32.NewProc("SetTimer")
	pKillTimer              = user32.NewProc("KillTimer")
	pTrackPopupMenu         = user32.NewProc("TrackPopupMenuEx")
	pCreatePopupMenu        = user32.NewProc("CreatePopupMenu")
	pAppendMenu             = user32.NewProc("AppendMenuW")
	pDestroyMenu            = user32.NewProc("DestroyMenu")
	pMessageBeep            = user32.NewProc("MessageBeep")
	pMessageBox             = user32.NewProc("MessageBoxW")

	// gdi32
	pCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	pCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	pDeleteDC               = gdi32.NewProc("DeleteDC")
	pDeleteObject           = gdi32.NewProc("DeleteObject")
	pSelectObject           = gdi32.NewProc("SelectObject")
	pBitBlt                 = gdi32.NewProc("BitBlt")
	pSetBkMode              = gdi32.NewProc("SetBkMode")
	pSetTextColor           = gdi32.NewProc("SetTextColor")
	pCreateSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	pCreatePen              = gdi32.NewProc("CreatePen")
	pCreateFont             = gdi32.NewProc("CreateFontW")
	pRoundRect              = gdi32.NewProc("RoundRect")
	pDrawText               = gdi32.NewProc("DrawTextW")
	pMoveToEx               = gdi32.NewProc("MoveToEx")
	pLineTo                 = gdi32.NewProc("LineTo")
	pAddFontMemResourceEx   = gdi32.NewProc("AddFontMemResourceEx")
	pGetStockObject         = gdi32.NewProc("GetStockObject")

	// kernel32
	pGetModuleHandle = kernel32.NewProc("GetModuleHandleW")

	// shell32
	pShellNotifyIcon = shell32.NewProc("Shell_NotifyIconW")
)

// ═══════════════════════════════════════════════════════════════════════════════
// Win32 Constants
// ═══════════════════════════════════════════════════════════════════════════════

const (
	// Window styles
	WS_POPUP       = 0x80000000
	WS_VISIBLE     = 0x10000000
	WS_SIZEBOX     = 0x00040000
	WS_EX_LAYERED  = 0x00080000
	WS_EX_TOOLWINDOW = 0x00000080
	WS_EX_APPWINDOW = 0x00040000
	WS_EX_TOPMOST  = 0x00000008

	HWND_TOPMOST   = ^uintptr(0) // -1
	HWND_NOTOPMOST = ^uintptr(1) // -2

	SWP_NOMOVE     = 0x0002
	SWP_NOSIZE     = 0x0001
	SWP_NOZORDER   = 0x0004
	SWP_NOACTIVATE = 0x0010

	SW_SHOW = 5
	SW_HIDE = 0

	LWA_ALPHA = 2

	// Messages
	WM_CREATE      = 0x0001
	WM_DESTROY     = 0x0002
	WM_SIZE        = 0x0005
	WM_PAINT       = 0x000F
	WM_ERASEBKGND  = 0x0014
	WM_LBUTTONDOWN = 0x0201
	WM_LBUTTONUP   = 0x0202
	WM_RBUTTONUP   = 0x0205
	WM_MOUSEMOVE   = 0x0200
	WM_TIMER       = 0x0113
	WM_COMMAND     = 0x0111
	WM_APP         = 0x8000
	WM_TRAYICON    = WM_APP + 1

	// Timer IDs
	TIMER_CLOCK    = 1
	TIMER_COUNTDOWN = 2

	// GDI
	TRANSPARENT     = 1
	SRCCOPY         = 0x00CC0020
	PS_SOLID        = 0
	FW_BOLD         = 700
	DEFAULT_CHARSET = 1
	OUT_TT_PRECIS   = 4
	CLIP_DEFAULT    = 0
	CLEARTYPE_QUALITY = 5
	FF_DONTCARE     = 0

	// DrawText
	DT_CENTER     = 0x0001
	DT_VCENTER    = 0x0004
	DT_SINGLELINE = 0x0020
	DT_NOPREFIX   = 0x0800

	// Menu
	MF_STRING    = 0x0000
	MF_SEPARATOR = 0x0800
	MF_CHECKED   = 0x0008
	TPM_LEFTALIGN  = 0x0000
	TPM_TOPALIGN   = 0x0000
	TPM_RETURNCMD  = 0x0100

	// Cursor
	IDC_ARROW  = 32512

	// Tray
	NIM_ADD    = 0
	NIM_DELETE = 2
	NIF_MESSAGE = 1
	NIF_ICON    = 2
	NIF_TIP     = 4

	// GWL
	GWL_STYLE   = -16
	GWL_EXSTYLE = -20

	// Stock objects
	BLACK_BRUSH = 4
)

// ═══════════════════════════════════════════════════════════════════════════════
// Win32 Structs
// ═══════════════════════════════════════════════════════════════════════════════

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

type POINT struct{ X, Y int32 }

type RECT struct{ Left, Top, Right, Bottom int32 }

type PAINTSTRUCT struct {
	HDC         uintptr
	FErase      int32
	RcPaint     RECT
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

type NOTIFYICONDATA struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
}

// ═══════════════════════════════════════════════════════════════════════════════
// Settings
// ═══════════════════════════════════════════════════════════════════════════════

type Settings struct {
	BgColor      uint32 `json:"bgColor"`
	BorderColor  uint32 `json:"borderColor"`
	DigitColor   uint32 `json:"digitColor"`
	Opacity      int    `json:"opacity"`
	AlwaysOnTop  bool   `json:"alwaysOnTop"`
	Format24h    bool   `json:"format24h"`
	TimerVisible bool   `json:"timerVisible"`
	WindowX      int    `json:"windowX"`
	WindowY      int    `json:"windowY"`
	WindowW      int    `json:"windowW"`
	WindowH      int    `json:"windowH"`
}

var defaultSettings = Settings{
	BgColor: 0x1a1a1a, BorderColor: 0x2979ff, DigitColor: 0xff0000,
	Opacity: 100, AlwaysOnTop: true, Format24h: true,
	WindowX: -1, WindowY: -1, WindowW: 400, WindowH: 110,
}

func settingsPath() string {
	dir, _ := os.UserConfigDir()
	if dir == "" {
		dir = "."
	}
	dir = filepath.Join(dir, "ClockWidget")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "settings.json")
}

func loadSettings() Settings {
	s := defaultSettings
	data, err := os.ReadFile(settingsPath())
	if err == nil {
		json.Unmarshal(data, &s)
	}
	return s
}

func saveSettings(s Settings) {
	data, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile(settingsPath(), data, 0644)
}

// RGB -> COLORREF (0xRRGGBB -> 0x00BBGGRR)
func colorRef(rgb uint32) uintptr {
	r := (rgb >> 16) & 0xFF
	g := (rgb >> 8) & 0xFF
	b := rgb & 0xFF
	return uintptr(b<<16 | g<<8 | r)
}

// ═══════════════════════════════════════════════════════════════════════════════
// Global State
// ═══════════════════════════════════════════════════════════════════════════════

var (
	appHwnd  uintptr
	settings Settings
	nid      NOTIFYICONDATA

	timeStr  = "00:00:00"
	timerStr = "05:00"
	mu       sync.RWMutex

	dragging            bool
	dragStartScreenX    int32
	dragStartScreenY    int32
	dragStartWindowX    int32
	dragStartWindowY    int32

	timerSeconds  = 300
	timerRunning  bool
	timerMu       sync.Mutex
)

// ═══════════════════════════════════════════════════════════════════════════════
// Font
// ═══════════════════════════════════════════════════════════════════════════════

func installFont() {
	var numFonts uint32
	pAddFontMemResourceEx.Call(
		uintptr(unsafe.Pointer(&dseg7Font[0])),
		uintptr(len(dseg7Font)),
		0,
		uintptr(unsafe.Pointer(&numFonts)),
	)
}

func makeFont(size int32) uintptr {
	name, _ := syscall.UTF16FromString("DSEG7 Classic")
	ret, _, _ := pCreateFont.Call(
		uintptr(uint32(-size)), 0, 0, 0, FW_BOLD, 0, 0, 0,
		DEFAULT_CHARSET, OUT_TT_PRECIS, CLIP_DEFAULT, CLEARTYPE_QUALITY,
		FF_DONTCARE, uintptr(unsafe.Pointer(&name[0])),
	)
	return ret
}

// ═══════════════════════════════════════════════════════════════════════════════
// Painting
// ═══════════════════════════════════════════════════════════════════════════════

func paint(hwnd uintptr) {
	mu.RLock()
	s := settings
	tStr := timeStr
	tmStr := timerStr
	mu.RUnlock()

	var ps PAINTSTRUCT
	hdc, _, _ := pBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	defer pEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

	var rc RECT
	pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top

	// Double buffer
	memDC, _, _ := pCreateCompatibleDC.Call(hdc)
	memBmp, _, _ := pCreateCompatibleBitmap.Call(hdc, uintptr(w), uintptr(h))
	oldBmp, _, _ := pSelectObject.Call(memDC, memBmp)
	defer func() {
		pSelectObject.Call(memDC, oldBmp)
		pDeleteObject.Call(memBmp)
		pDeleteDC.Call(memDC)
	}()

	// Fill background
	bgBrush, _, _ := pCreateSolidBrush.Call(colorRef(s.BgColor))
	bgR := RECT{0, 0, w, h}
	pFillRect.Call(memDC, uintptr(unsafe.Pointer(&bgR)), bgBrush)
	pDeleteObject.Call(bgBrush)

	// Rounded border
	pen, _, _ := pCreatePen.Call(PS_SOLID, 3, colorRef(s.BorderColor))
	// Need a null brush for the interior
	nullBrush, _, _ := pGetStockObject.Call(5) // HOLLOW_BRUSH
	oldP, _, _ := pSelectObject.Call(memDC, pen)
	oldB, _, _ := pSelectObject.Call(memDC, nullBrush)
	pRoundRect.Call(memDC, 1, 1, uintptr(w-1), uintptr(h-1), 20, 20)
	pSelectObject.Call(memDC, oldP)
	pSelectObject.Call(memDC, oldB)
	pDeleteObject.Call(pen)

	// Clock area vs timer area
	clockH := h
	timerH := int32(0)
	if s.TimerVisible {
		timerH = 35
		clockH = h - timerH
	}

	// Clock text
	pSetBkMode.Call(memDC, TRANSPARENT)
	pSetTextColor.Call(memDC, colorRef(s.DigitColor))

	fontSize := int32(float64(clockH) * 0.72)
	if maxFW := int32(float64(w) * 0.13); maxFW < fontSize {
		fontSize = maxFW
	}
	if fontSize < 16 {
		fontSize = 16
	}

	hFont := makeFont(fontSize)
	oldFont, _, _ := pSelectObject.Call(memDC, hFont)

	clockRect := RECT{0, 0, w, clockH}
	text16, _ := syscall.UTF16FromString(tStr)
	pDrawText.Call(memDC, uintptr(unsafe.Pointer(&text16[0])),
		uintptr(len(text16)-1), uintptr(unsafe.Pointer(&clockRect)),
		DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX)

	pSelectObject.Call(memDC, oldFont)
	pDeleteObject.Call(hFont)

	// Timer
	if s.TimerVisible {
		// Divider
		divPen, _, _ := pCreatePen.Call(PS_SOLID, 1, colorRef(s.BorderColor))
		oldDP, _, _ := pSelectObject.Call(memDC, divPen)
		pMoveToEx.Call(memDC, 10, uintptr(clockH), 0)
		pLineTo.Call(memDC, uintptr(w-10), uintptr(clockH))
		pSelectObject.Call(memDC, oldDP)
		pDeleteObject.Call(divPen)

		timerFontSize := int32(float64(timerH) * 0.55)
		if timerFontSize < 12 {
			timerFontSize = 12
		}
		hTFont := makeFont(timerFontSize)
		oldTF, _, _ := pSelectObject.Call(memDC, hTFont)

		timerRect := RECT{0, clockH + 2, w, h}
		t16, _ := syscall.UTF16FromString(tmStr)
		pDrawText.Call(memDC, uintptr(unsafe.Pointer(&t16[0])),
			uintptr(len(t16)-1), uintptr(unsafe.Pointer(&timerRect)),
			DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX)

		pSelectObject.Call(memDC, oldTF)
		pDeleteObject.Call(hTFont)
	}

	// Blit
	pBitBlt.Call(hdc, 0, 0, uintptr(w), uintptr(h), memDC, 0, 0, SRCCOPY)
}

// ═══════════════════════════════════════════════════════════════════════════════
// Tray Icon
// ═══════════════════════════════════════════════════════════════════════════════

func addTrayIcon(hwnd uintptr) {
	hInst, _, _ := pGetModuleHandle.Call(0)
	hIcon, _, _ := pLoadIcon.Call(hInst, 1) // icon resource ID 1

	nid = NOTIFYICONDATA{
		CbSize:           uint32(unsafe.Sizeof(NOTIFYICONDATA{})),
		HWnd:             hwnd,
		UID:              1,
		UFlags:           NIF_MESSAGE | NIF_ICON | NIF_TIP,
		UCallbackMessage: WM_TRAYICON,
		HIcon:            hIcon,
	}
	tip := syscall.StringToUTF16("Clock Widget")
	copy(nid.SzTip[:], tip)
	pShellNotifyIcon.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
}

func removeTrayIcon() {
	pShellNotifyIcon.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
}

// ═══════════════════════════════════════════════════════════════════════════════
// Taskbar Hide
// ═══════════════════════════════════════════════════════════════════════════════

func hideFromTaskbar(hwnd uintptr) {
	gwlExStyle := uintptr(0xFFFFFFFFFFFFFFEC) // GWL_EXSTYLE = -20
	style, _, _ := pGetWindowLong.Call(hwnd, gwlExStyle)
	style = (style &^ WS_EX_APPWINDOW) | WS_EX_TOOLWINDOW
	pShowWindow.Call(hwnd, SW_HIDE)
	pSetWindowLong.Call(hwnd, gwlExStyle, style)
	pShowWindow.Call(hwnd, SW_SHOW)
}

// ═══════════════════════════════════════════════════════════════════════════════
// Always on Top
// ═══════════════════════════════════════════════════════════════════════════════

func setOnTop(hwnd uintptr, onTop bool) {
	flag := HWND_NOTOPMOST
	if onTop {
		flag = HWND_TOPMOST
	}
	pSetWindowPos.Call(hwnd, flag, 0, 0, 0, 0,
		SWP_NOMOVE|SWP_NOSIZE|SWP_NOACTIVATE)
}

// ═══════════════════════════════════════════════════════════════════════════════
// Context Menu
// ═══════════════════════════════════════════════════════════════════════════════

const (
	CMD_SETTINGS  = 101
	CMD_TIMER     = 102
	CMD_ONTOP     = 103
	CMD_TSTART    = 104
	CMD_TPAUSE    = 105
	CMD_TRESET    = 106
	CMD_TADD      = 107
	CMD_TSUB      = 108
	CMD_EXIT      = 109
)

func showContextMenu(hwnd uintptr) {
	mu.RLock()
	s := settings
	mu.RUnlock()

	hMenu, _, _ := pCreatePopupMenu.Call()
	appendStr := func(id uintptr, text string, checked bool) {
		t, _ := syscall.UTF16PtrFromString(text)
		flags := uintptr(MF_STRING)
		if checked {
			flags |= MF_CHECKED
		}
		pAppendMenu.Call(hMenu, flags, id, uintptr(unsafe.Pointer(t)))
	}
	appendSep := func() {
		pAppendMenu.Call(hMenu, MF_SEPARATOR, 0, 0)
	}

	appendStr(CMD_SETTINGS, "⚙ Settings", false)
	appendStr(CMD_TIMER, "⏱ Timer", s.TimerVisible)
	if s.TimerVisible {
		appendSep()
		appendStr(CMD_TSTART, "▶ Start Timer", false)
		appendStr(CMD_TPAUSE, "⏸ Pause Timer", false)
		appendStr(CMD_TRESET, "↺ Reset Timer (5:00)", false)
		appendStr(CMD_TADD, "+1 Minute", false)
		appendStr(CMD_TSUB, "-1 Minute", false)
	}
	appendSep()
	appendStr(CMD_ONTOP, "📌 Always on Top", s.AlwaysOnTop)
	appendSep()
	appendStr(CMD_EXIT, "✕ Exit", false)

	var pt POINT
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	pSetForegroundWindow.Call(hwnd)
	cmd, _, _ := pTrackPopupMenu.Call(hMenu, TPM_LEFTALIGN|TPM_TOPALIGN|TPM_RETURNCMD,
		uintptr(pt.X), uintptr(pt.Y), 0, hwnd, 0)
	pDestroyMenu.Call(hMenu)

	handleMenuCmd(hwnd, cmd)
}

func handleMenuCmd(hwnd uintptr, cmd uintptr) {
	switch cmd {
	case CMD_SETTINGS:
		// TODO: native settings dialog (color pickers)
		// For now: reset to defaults as a placeholder
		mu.Lock()
		settings = defaultSettings
		mu.Unlock()
		applyOpacity(hwnd)
		setOnTop(hwnd, settings.AlwaysOnTop)
		pInvalidateRect.Call(hwnd, 0, 1)
		saveSettingsAsync()

	case CMD_TIMER:
		mu.Lock()
		settings.TimerVisible = !settings.TimerVisible
		vis := settings.TimerVisible
		mu.Unlock()
		// Resize window
		var rc RECT
		pGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
		curH := rc.Bottom - rc.Top
		if vis {
			updateTimerDisplay()
			pSetWindowPos.Call(hwnd, 0, uintptr(rc.Left), uintptr(rc.Top),
				uintptr(rc.Right-rc.Left), uintptr(curH+35), SWP_NOZORDER)
		} else {
			stopCountdown()
			newH := curH - 35
			if newH < 60 {
				newH = 110
			}
			pSetWindowPos.Call(hwnd, 0, uintptr(rc.Left), uintptr(rc.Top),
				uintptr(rc.Right-rc.Left), uintptr(newH), SWP_NOZORDER)
		}
		pInvalidateRect.Call(hwnd, 0, 1)
		saveSettingsAsync()

	case CMD_ONTOP:
		mu.Lock()
		settings.AlwaysOnTop = !settings.AlwaysOnTop
		onTop := settings.AlwaysOnTop
		mu.Unlock()
		setOnTop(hwnd, onTop)
		saveSettingsAsync()

	case CMD_TSTART:
		startCountdown(hwnd)

	case CMD_TPAUSE:
		stopCountdown()

	case CMD_TRESET:
		stopCountdown()
		timerMu.Lock()
		timerSeconds = 300
		timerMu.Unlock()
		updateTimerDisplay()
		pInvalidateRect.Call(hwnd, 0, 1)

	case CMD_TADD:
		timerMu.Lock()
		if timerSeconds < 5999 {
			timerSeconds += 60
		}
		timerMu.Unlock()
		updateTimerDisplay()
		pInvalidateRect.Call(hwnd, 0, 1)

	case CMD_TSUB:
		timerMu.Lock()
		if timerSeconds >= 60 {
			timerSeconds -= 60
		}
		timerMu.Unlock()
		updateTimerDisplay()
		pInvalidateRect.Call(hwnd, 0, 1)

	case CMD_EXIT:
		saveSettingsFinal()
		removeTrayIcon()
		pPostQuitMessage.Call(0)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// Timer
// ═══════════════════════════════════════════════════════════════════════════════

func updateTimerDisplay() {
	timerMu.Lock()
	secs := timerSeconds
	timerMu.Unlock()
	m := secs / 60
	s := secs % 60
	mu.Lock()
	timerStr = fmt.Sprintf("%02d:%02d", m, s)
	mu.Unlock()
}

func startCountdown(hwnd uintptr) {
	timerMu.Lock()
	defer timerMu.Unlock()
	if timerRunning || timerSeconds <= 0 {
		return
	}
	timerRunning = true
	pSetTimer.Call(hwnd, TIMER_COUNTDOWN, 1000, 0)
}

func stopCountdown() {
	timerMu.Lock()
	defer timerMu.Unlock()
	if timerRunning {
		timerRunning = false
		pKillTimer.Call(appHwnd, TIMER_COUNTDOWN)
	}
}

func tickCountdown(hwnd uintptr) {
	timerMu.Lock()
	if !timerRunning {
		timerMu.Unlock()
		return
	}
	timerSeconds--
	if timerSeconds <= 0 {
		timerSeconds = 0
		timerRunning = false
		pKillTimer.Call(hwnd, TIMER_COUNTDOWN)
		timerMu.Unlock()
		updateTimerDisplay()
		pInvalidateRect.Call(hwnd, 0, 1)
		// Notification
		go func() {
			for i := 0; i < 3; i++ {
				pMessageBeep.Call(0xFFFFFFFF)
				time.Sleep(300 * time.Millisecond)
			}
			title, _ := syscall.UTF16PtrFromString("Timer Complete")
			msg, _ := syscall.UTF16PtrFromString("Your countdown timer has finished!")
			pMessageBox.Call(hwnd, uintptr(unsafe.Pointer(msg)),
				uintptr(unsafe.Pointer(title)), 0x40) // MB_ICONINFORMATION
		}()
		return
	}
	timerMu.Unlock()
	updateTimerDisplay()
	pInvalidateRect.Call(hwnd, 0, 1)
}

// ═══════════════════════════════════════════════════════════════════════════════
// Settings Persistence
// ═══════════════════════════════════════════════════════════════════════════════

func saveSettingsAsync() {
	go func() {
		mu.RLock()
		s := settings
		mu.RUnlock()
		saveSettings(s)
	}()
}

func saveSettingsFinal() {
	mu.RLock()
	s := settings
	mu.RUnlock()
	// Update window position
	var rc RECT
	pGetWindowRect.Call(appHwnd, uintptr(unsafe.Pointer(&rc)))
	s.WindowX = int(rc.Left)
	s.WindowY = int(rc.Top)
	s.WindowW = int(rc.Right - rc.Left)
	s.WindowH = int(rc.Bottom - rc.Top)
	saveSettings(s)
}

func applyOpacity(hwnd uintptr) {
	mu.RLock()
	pct := settings.Opacity
	mu.RUnlock()
	alpha := int(float64(pct) / 100.0 * 255.0)
	if alpha < 50 {
		alpha = 50
	}
	pSetLayeredWindowAttrib.Call(hwnd, 0, uintptr(alpha), LWA_ALPHA)
}

// ═══════════════════════════════════════════════════════════════════════════════
// Window Proc
// ═══════════════════════════════════════════════════════════════════════════════

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_CREATE:
		appHwnd = hwnd
		// Start clock timer (1 second)
		pSetTimer.Call(hwnd, TIMER_CLOCK, 1000, 0)
		// Tray + taskbar
		addTrayIcon(hwnd)
		go func() {
			time.Sleep(200 * time.Millisecond)
			hideFromTaskbar(hwnd)
		}()
		return 0

	case WM_PAINT:
		paint(hwnd)
		return 0

	case WM_ERASEBKGND:
		return 1

	case WM_TIMER:
		if wParam == TIMER_CLOCK {
			updateClockStr()
			pInvalidateRect.Call(hwnd, 0, 1)
		} else if wParam == TIMER_COUNTDOWN {
			tickCountdown(hwnd)
		}
		return 0

	case WM_LBUTTONDOWN:
		dragging = true
		var pt POINT
		pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
		dragStartScreenX = pt.X
		dragStartScreenY = pt.Y
		var rc RECT
		pGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
		dragStartWindowX = rc.Left
		dragStartWindowY = rc.Top
		pSetCapture.Call(hwnd)
		return 0

	case WM_MOUSEMOVE:
		if dragging {
			var pt POINT
			pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
			dx := pt.X - dragStartScreenX
			dy := pt.Y - dragStartScreenY
			newX := dragStartWindowX + dx
			newY := dragStartWindowY + dy
			pSetWindowPos.Call(hwnd, 0, uintptr(newX), uintptr(newY), 0, 0,
				SWP_NOSIZE|SWP_NOZORDER)
		}
		return 0

	case WM_LBUTTONUP:
		if dragging {
			dragging = false
			pReleaseCapture.Call()
			saveSettingsFinal()
		}
		return 0

	case WM_RBUTTONUP:
		showContextMenu(hwnd)
		return 0

	case WM_SIZE:
		pInvalidateRect.Call(hwnd, 0, 1)
		return 0

	case WM_TRAYICON:
		switch lParam {
		case WM_LBUTTONUP:
			pShowWindow.Call(hwnd, SW_SHOW)
			pSetForegroundWindow.Call(hwnd)
		case WM_RBUTTONUP:
			showContextMenu(hwnd)
		}
		return 0

	case WM_DESTROY:
		removeTrayIcon()
		pKillTimer.Call(hwnd, TIMER_CLOCK)
		pKillTimer.Call(hwnd, TIMER_COUNTDOWN)
		saveSettingsFinal()
		pPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := pDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func updateClockStr() {
	now := time.Now()
	mu.RLock()
	fmt24 := settings.Format24h
	mu.RUnlock()

	hours := now.Hour()
	if !fmt24 {
		hours = hours % 12
		if hours == 0 {
			hours = 12
		}
	}
	str := fmt.Sprintf("%02d:%02d:%02d", hours, now.Minute(), now.Second())
	mu.Lock()
	timeStr = str
	mu.Unlock()
}

// ═══════════════════════════════════════════════════════════════════════════════
// Main
// ═══════════════════════════════════════════════════════════════════════════════

func main() {
	installFont()
	settings = loadSettings()

	// Initial time
	updateClockStr()
	if settings.TimerVisible {
		updateTimerDisplay()
	}

	hInst, _, _ := pGetModuleHandle.Call(0)
	className, _ := syscall.UTF16PtrFromString("ClockWidgetNative")

	cursor, _, _ := pLoadCursor.Call(0, IDC_ARROW)
	bgBrush, _, _ := pGetStockObject.Call(BLACK_BRUSH)

	wc := WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
		Style:         0x0003, // CS_HREDRAW | CS_VREDRAW
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     hInst,
		HCursor:       cursor,
		HbrBackground: bgBrush,
		LpszClassName: uintptr(unsafe.Pointer(className)),
	}
	pRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))

	w, h := settings.WindowW, settings.WindowH
	if w <= 0 {
		w = 400
	}
	if h <= 0 {
		h = 110
	}
	x := settings.WindowX
	y := settings.WindowY
	if x < 0 {
		x = 100
	}
	if y < 0 {
		y = 100
	}

	exStyle := uintptr(WS_EX_LAYERED)
	if settings.AlwaysOnTop {
		exStyle |= WS_EX_TOPMOST
	}

	title, _ := syscall.UTF16PtrFromString("ClockWidget")
	hwnd, _, _ := pCreateWindowEx.Call(
		exStyle,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		WS_POPUP|WS_VISIBLE|WS_SIZEBOX,
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		0, 0, hInst, 0,
	)

	if hwnd == 0 {
		panic("Failed to create window")
	}
	appHwnd = hwnd

	// Set opacity
	applyOpacity(hwnd)

	// Message loop
	var msg MSG
	for {
		ret, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		pDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
