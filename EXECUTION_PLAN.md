# Execution Plan — Digital Clock Widget

## Prerequisites

- [x] Go 1.21+ installed
- [ ] Wails CLI v2 installed (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- [ ] WebView2 runtime (pre-installed on Windows 10 21H2+ / Windows 11)

---

## Phase 1: Project Scaffold (10 min)

| Step | Task                                              | Details                                    |
|------|---------------------------------------------------|--------------------------------------------|
| 1.1  | Install Wails CLI                                 | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| 1.2  | Initialize Wails project                          | `wails init -n ClockWidget -t vanilla`     |
| 1.3  | Configure `wails.json`                            | Frameless, transparent, size, title         |
| 1.4  | Verify scaffold builds and runs                   | `wails dev`                                |

---

## Phase 2: Clock Display Frontend (30 min)

| Step | Task                                              | Details                                    |
|------|---------------------------------------------------|--------------------------------------------|
| 2.1  | Create 7-segment CSS digit system                 | Pure CSS segments, no external fonts        |
| 2.2  | Build clock face layout (HH:MM:SS)                | Flexbox, centered, responsive to container  |
| 2.3  | Implement JS time update loop                     | `setInterval` every 1s, update digits       |
| 2.4  | Style: dark background, red digits, blue border   | Match reference screenshots                 |
| 2.5  | Make clock scale with window resize               | CSS viewport units or container queries     |

---

## Phase 3: Window Behavior (20 min)

| Step | Task                                              | Details                                    |
|------|---------------------------------------------------|--------------------------------------------|
| 3.1  | Frameless window config in `wails.json`            | `"Frameless": true`                        |
| 3.2  | Implement drag-to-move via mouse events            | JS mousedown/mousemove on clock face       |
| 3.3  | Enable window resize                              | Wails window resize handles                 |
| 3.4  | Always-on-top via Wails runtime API                | `WindowSetAlwaysOnTop(true)`               |
| 3.5  | Rounded corners + border styling                   | CSS `border-radius`, Wails transparency    |

---

## Phase 4: Settings & Color Customization (40 min)

| Step | Task                                              | Details                                    |
|------|---------------------------------------------------|--------------------------------------------|
| 4.1  | Go backend: Settings struct + JSON persistence     | Load/Save to `%APPDATA%\ClockWidget\`      |
| 4.2  | Go backend: Expose Get/Save settings to frontend   | Wails binding methods                       |
| 4.3  | Right-click context menu (JS)                      | Settings, Always on Top, Close             |
| 4.4  | Settings overlay/panel UI                          | Color pickers, transparency slider          |
| 4.5  | Wire color pickers to live CSS variable updates    | `--bg-color`, `--border-color`, `--digit-color` |
| 4.6  | Transparency slider → Wails window opacity API     | `WindowSetAlpha()`                         |
| 4.7  | Load saved settings on startup                     | Apply colors + position + size              |

---

## Phase 5: Polish & Build (20 min)

| Step | Task                                              | Details                                    |
|------|---------------------------------------------------|--------------------------------------------|
| 5.1  | Close button (X) on hover in corner                | Small, subtle, appears on mouse enter       |
| 5.2  | Window position/size persistence                   | Save on move/resize, restore on launch      |
| 5.3  | App icon                                          | Simple clock icon embedded in exe           |
| 5.4  | Production build                                   | `wails build` → single `.exe`              |
| 5.5  | Test: resize, drag, color changes, restart          | Full smoke test                             |
| 5.6  | Verify binary size < 10MB                          | Strip debug symbols if needed               |

---

## Estimated Timeline

| Phase                        | Duration |
|------------------------------|----------|
| Phase 1: Scaffold            | 10 min   |
| Phase 2: Clock Display       | 30 min   |
| Phase 3: Window Behavior     | 20 min   |
| Phase 4: Settings            | 40 min   |
| Phase 5: Polish & Build      | 20 min   |
| **Total**                    | **~2 hr** |

---

## Risk Mitigation

| Risk                                    | Mitigation                                          |
|-----------------------------------------|-----------------------------------------------------|
| WebView2 not installed on target PC     | Bundle WebView2 loader or document requirement      |
| Wails window transparency issues        | Fall back to solid background with CSS transparency |
| 7-segment CSS complexity                | Use proven CSS-only segment patterns                |
| Binary size exceeds target              | UPX compression as fallback                         |

---

## File Structure (Final)

```
ClockWidget/
├── wails.json                  # Wails project config
├── main.go                     # App entry point
├── app.go                      # Settings backend (Go)
├── frontend/
│   ├── index.html              # Clock face + settings panel
│   ├── src/
│   │   ├── style.css           # 7-segment digits, themes, layout
│   │   └── main.js             # Clock logic, settings, drag, context menu
│   └── wailsjs/                # Auto-generated Wails bindings
├── build/
│   └── appicon.png             # App icon
├── PRD.md                      # Product Requirements
└── EXECUTION_PLAN.md           # This file
```
