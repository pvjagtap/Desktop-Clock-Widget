# Digital Clock Widget — Product Requirements Document

## 1. Overview

A lightweight, resizable digital clock widget for Windows that displays time in a 7-segment LED style. The widget sits on the desktop as an always-on-top, frameless, draggable window with customizable colors and transparency.

**Inspired by**: Widget Launcher's "Digital Clock" widget (Colorful Dark skin)

---

## 2. Target Platform

- **OS**: Windows 10/11 (64-bit)
- **Output**: Single `.exe` file
- **Size Target**: < 10 MB
- **Runtime Dependency**: WebView2 (pre-installed on Windows 10 21H2+ and all Windows 11)

---

## 3. Tech Stack

| Layer       | Technology         | Rationale                                        |
|-------------|--------------------|--------------------------------------------------|
| Framework   | **Wails v2**       | Go + WebView2, tiny binary, native window control|
| Backend     | **Go 1.21+**       | Settings persistence, system tray, window mgmt   |
| Frontend    | **Vanilla HTML/CSS/JS** | No framework needed — keeps bundle minimal   |
| Clock Font  | **CSS 7-segment**  | Pure CSS digital display, no font files needed    |
| Storage     | **JSON file**      | `~/.clockwidget/settings.json` for preferences   |

---

## 4. Functional Requirements

### 4.1 Clock Display

| ID    | Requirement                                              | Priority |
|-------|----------------------------------------------------------|----------|
| F-01  | Display current time in HH:MM:SS format (24h)           | Must     |
| F-02  | 7-segment LED-style digit rendering (red on black)       | Must     |
| F-03  | Colon separators between HH, MM, SS                     | Must     |
| F-04  | Time updates every second with smooth rendering          | Must     |
| F-05  | Optional 12h/24h toggle                                 | Nice     |

### 4.2 Window Behavior

| ID    | Requirement                                              | Priority |
|-------|----------------------------------------------------------|----------|
| W-01  | Frameless window (no title bar, no OS chrome)            | Must     |
| W-02  | Always-on-top by default (toggleable)                    | Must     |
| W-03  | Draggable by clicking anywhere on the clock face         | Must     |
| W-04  | Resizable — clock digits scale proportionally            | Must     |
| W-05  | Rounded corners matching the reference design            | Must     |
| W-06  | Remember position and size across restarts               | Must     |
| W-07  | Minimize to system tray                                  | Nice     |

### 4.3 Color Customization

| ID    | Requirement                                              | Priority |
|-------|----------------------------------------------------------|----------|
| C-01  | **Background** color picker (default: `#1a1a1a` black)  | Must     |
| C-02  | **Border** color picker (default: `#2979ff` blue)        | Must     |
| C-03  | **Font/Digit** color picker (default: `#ff0000` red)    | Must     |
| C-04  | **Accent** color picker (default: `#2979ff` blue)        | Nice     |
| C-05  | Transparency/opacity slider (0–100%)                     | Must     |
| C-06  | Settings accessible via right-click context menu         | Must     |
| C-07  | Live preview — colors change in real-time                | Must     |

### 4.4 Settings Panel

| ID    | Requirement                                              | Priority |
|-------|----------------------------------------------------------|----------|
| S-01  | Settings panel opens as a separate small window or overlay| Must    |
| S-02  | Color pickers for Background, Border, Font Color         | Must     |
| S-03  | Transparency slider                                      | Must     |
| S-04  | "Reset to Default" button                                | Nice     |
| S-05  | Settings persist to disk (JSON)                          | Must     |

### 4.5 System Integration

| ID    | Requirement                                              | Priority |
|-------|----------------------------------------------------------|----------|
| I-01  | Close button (X) on hover or via context menu            | Must     |
| I-02  | Right-click context menu: Settings, Always on Top, Exit  | Must     |
| I-03  | Single instance — prevent multiple clocks launching      | Nice     |

---

## 5. Non-Functional Requirements

| ID    | Requirement                                              | Target   |
|-------|----------------------------------------------------------|----------|
| NF-01 | Binary size                                              | < 10 MB  |
| NF-02 | Memory usage                                             | < 30 MB  |
| NF-03 | CPU usage (idle)                                         | < 1%     |
| NF-04 | Startup time                                             | < 1s     |
| NF-05 | No external runtime dependencies (besides WebView2)      | Required |

---

## 6. Visual Reference

```
┌──────────────────────────────────┐
│  ╔══════════════════════════╗    │  Border: Blue (#2979ff), rounded
│  ║                          ║    │  Background: Black (#1a1a1a)
│  ║   11 : 49 : 28          ║    │  Digits: Red (#ff0000), 7-segment
│  ║                          ║    │
│  ╚══════════════════════════╝    │
└──────────────────────────────────┘
       ↑ Draggable, resizable, always-on-top
```

**Default "Colorful Dark" Theme:**
- Background: `#1a1a1a` (near-black)
- Border: `#2979ff` (bright blue), 3px, rounded 16px
- Digits: `#ff0000` (bright red), 7-segment LED style
- Accent: `#2979ff` (blue, for settings UI elements)

---

## 7. Out of Scope (v1)

- Multiple clock instances / zones
- Alarm / timer functionality
- Date display
- Analog clock mode
- Skin/theme presets (beyond color pickers)
- Auto-update mechanism
- Installer (single exe, portable)

---

## 8. Delivery

- Single portable `.exe` file
- No installer required
- Settings stored in `%APPDATA%\ClockWidget\settings.json`
