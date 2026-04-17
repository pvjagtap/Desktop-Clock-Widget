import { GetSettings, SaveSettings } from '../wailsjs/go/main/App.js';

// === Wails Runtime Wrappers ===
function wrt() { return window.runtime || {}; }

function setAlwaysOnTop(v) { wrt().WindowSetAlwaysOnTop?.(v); }
function setPosition(x, y) { wrt().WindowSetPosition?.(x, y); }
function setSize(w, h) { wrt().WindowSetSize?.(w, h); }
function getPosition() { return wrt().WindowGetPosition?.() || Promise.resolve({ x: 0, y: 0 }); }
function getSize() { return wrt().WindowGetSize?.() || Promise.resolve({ w: 380, h: 100 }); }
function quit() { wrt().Quit?.(); }

// === Default Settings ===
const DEFAULTS = {
    bgColor: '#1a1a1a',
    borderColor: '#2979ff',
    digitColor: '#ff0000',
    opacity: 100,
    alwaysOnTop: true,
    format24h: true,
    timerVisible: false,
    windowX: -1,
    windowY: -1,
    windowW: 380,
    windowH: 100,
};

const SETTINGS_HEIGHT = 190;  // height added when settings open
const TIMER_HEIGHT = 40;      // height added when timer visible

let settings = { ...DEFAULTS };
let isOnTop = true;
let settingsOpen = false;
let baseH = 100; // stored window height without extras

// === DOM Elements ===
const appWrapper = document.getElementById('app-wrapper');
const clockContainer = document.getElementById('clock-container');
const digits = {
    h1: document.getElementById('h1'),
    h2: document.getElementById('h2'),
    m1: document.getElementById('m1'),
    m2: document.getElementById('m2'),
    s1: document.getElementById('s1'),
    s2: document.getElementById('s2'),
};

const settingsPanel = document.getElementById('settings-panel');
const contextMenu = document.getElementById('context-menu');
const timerContainer = document.getElementById('timer-container');

const colorBg = document.getElementById('color-bg');
const colorBorder = document.getElementById('color-border');
const colorDigits = document.getElementById('color-digits');
const opacitySlider = document.getElementById('opacity-slider');
const opacityValue = document.getElementById('opacity-value');
const alwaysOnTopCheck = document.getElementById('always-on-top');
const format24hCheck = document.getElementById('format-24h');

const ctxOntop = document.getElementById('ctx-ontop');
const ctxTimer = document.getElementById('ctx-timer');

// Timer DOM
const timerDigits = {
    tm1: document.getElementById('tm1'),
    tm2: document.getElementById('tm2'),
    ts1: document.getElementById('ts1'),
    ts2: document.getElementById('ts2'),
};
const timerStartBtn = document.getElementById('timer-start');
const timerPauseBtn = document.getElementById('timer-pause');
const timerResetBtn = document.getElementById('timer-reset');
const timerUpBtn = document.getElementById('timer-up');
const timerDownBtn = document.getElementById('timer-down');

// === Clock Update ===
function updateClock() {
    const now = new Date();
    let hours = now.getHours();
    if (!settings.format24h) hours = hours % 12 || 12;

    const h = String(hours).padStart(2, '0');
    const m = String(now.getMinutes()).padStart(2, '0');
    const s = String(now.getSeconds()).padStart(2, '0');

    digits.h1.textContent = h[0];
    digits.h2.textContent = h[1];
    digits.m1.textContent = m[0];
    digits.m2.textContent = m[1];
    digits.s1.textContent = s[0];
    digits.s2.textContent = s[1];
}

// === Apply Theme Colors ===
function applyColors() {
    const root = document.documentElement;
    root.style.setProperty('--bg-color', settings.bgColor);
    root.style.setProperty('--border-color', settings.borderColor);
    root.style.setProperty('--digit-color', settings.digitColor);

    const r = parseInt(settings.digitColor.slice(1, 3), 16);
    const g = parseInt(settings.digitColor.slice(3, 5), 16);
    const b = parseInt(settings.digitColor.slice(5, 7), 16);
    root.style.setProperty('--digit-glow', `rgba(${r}, ${g}, ${b}, 0.4)`);

    clockContainer.style.opacity = (settings.opacity / 100).toString();
}

// === Window Size Management ===
function calcTotalHeight() {
    let h = baseH;
    if (settings.timerVisible) h += TIMER_HEIGHT;
    if (settingsOpen) h += SETTINGS_HEIGHT;
    return h;
}

async function resizeWindow() {
    const size = await getSize();
    const newH = calcTotalHeight();
    setSize(size.w, newH);
}

// === Drag to Move ===
let isDragging = false;
let dragX = 0, dragY = 0;

clockContainer.addEventListener('mousedown', (e) => {
    if (e.target.id === 'close-btn' || e.target.id === 'settings-btn') return;
    if (e.button !== 0) return;
    isDragging = true;
    dragX = e.screenX;
    dragY = e.screenY;
});

document.addEventListener('mousemove', (e) => {
    if (!isDragging) return;
    const dx = e.screenX - dragX;
    const dy = e.screenY - dragY;
    if (Math.abs(dx) > 2 || Math.abs(dy) > 2) {
        dragX = e.screenX;
        dragY = e.screenY;
        getPosition().then(pos => setPosition(pos.x + dx, pos.y + dy));
    }
});

document.addEventListener('mouseup', () => {
    if (isDragging) {
        isDragging = false;
        savePosition();
    }
});

// === Settings Panel (auto-expand window) ===
function openSettings() {
    hideContextMenu();
    if (settingsOpen) return;
    settingsOpen = true;
    settingsPanel.classList.remove('hidden');
    appWrapper.classList.add('settings-open');

    colorBg.value = settings.bgColor;
    colorBorder.value = settings.borderColor;
    colorDigits.value = settings.digitColor;
    opacitySlider.value = settings.opacity;
    opacityValue.textContent = settings.opacity + '%';
    alwaysOnTopCheck.checked = settings.alwaysOnTop;
    format24hCheck.checked = settings.format24h;

    resizeWindow();
}

function closeSettingsPanel() {
    if (!settingsOpen) return;
    settingsOpen = false;
    settingsPanel.classList.add('hidden');
    appWrapper.classList.remove('settings-open');
    resizeWindow();
    persistSettings();
}

document.getElementById('settings-btn').addEventListener('click', openSettings);

// === Live Color Updates ===
colorBg.addEventListener('input', (e) => { settings.bgColor = e.target.value; applyColors(); });
colorBorder.addEventListener('input', (e) => { settings.borderColor = e.target.value; applyColors(); });
colorDigits.addEventListener('input', (e) => { settings.digitColor = e.target.value; applyColors(); });

opacitySlider.addEventListener('input', (e) => {
    settings.opacity = parseInt(e.target.value);
    opacityValue.textContent = settings.opacity + '%';
    applyColors();
});

alwaysOnTopCheck.addEventListener('change', (e) => {
    settings.alwaysOnTop = e.target.checked;
    isOnTop = settings.alwaysOnTop;
    setAlwaysOnTop(isOnTop);
});

format24hCheck.addEventListener('change', (e) => {
    settings.format24h = e.target.checked;
    updateClock();
});

document.getElementById('reset-defaults').addEventListener('click', () => {
    const wasTimer = settings.timerVisible;
    settings = { ...DEFAULTS, timerVisible: wasTimer };
    applyColors();
    setAlwaysOnTop(true);
    isOnTop = true;
    openSettings();
    persistSettings();
});

// === Timer Logic ===
let timerSeconds = 300; // default 5 minutes
let timerRunning = false;
let timerInterval = null;

function updateTimerDisplay() {
    const m = String(Math.floor(timerSeconds / 60)).padStart(2, '0');
    const s = String(timerSeconds % 60).padStart(2, '0');
    timerDigits.tm1.textContent = m[0];
    timerDigits.tm2.textContent = m[1];
    timerDigits.ts1.textContent = s[0];
    timerDigits.ts2.textContent = s[1];
}

function timerTick() {
    if (timerSeconds <= 0) {
        stopTimer();
        // Flash effect on completion
        timerContainer.style.background = '#442200';
        setTimeout(() => { timerContainer.style.background = ''; }, 300);
        setTimeout(() => { timerContainer.style.background = '#442200'; }, 600);
        setTimeout(() => { timerContainer.style.background = ''; }, 900);
        // Windows notification
        showTimerNotification();
        return;
    }
    timerSeconds--;
    updateTimerDisplay();
}

function showTimerNotification() {
    // Play beep sound
    try {
        const audioCtx = new (window.AudioContext || window.webkitAudioContext)();
        for (let i = 0; i < 3; i++) {
            const osc = audioCtx.createOscillator();
            const gain = audioCtx.createGain();
            osc.connect(gain);
            gain.connect(audioCtx.destination);
            osc.frequency.value = 880;
            osc.type = 'square';
            gain.gain.value = 0.08;
            osc.start(audioCtx.currentTime + i * 0.4);
            osc.stop(audioCtx.currentTime + i * 0.4 + 0.2);
        }
    } catch (e) { /* audio not available */ }

    // Browser/OS notification
    if ('Notification' in window) {
        if (Notification.permission === 'granted') {
            new Notification('Timer Complete', { body: 'Your countdown timer has finished!', icon: './src/fonts/DSEG7Classic-Bold.woff2' });
        } else if (Notification.permission !== 'denied') {
            Notification.requestPermission().then(perm => {
                if (perm === 'granted') {
                    new Notification('Timer Complete', { body: 'Your countdown timer has finished!' });
                }
            });
        }
    }
}

function startTimer() {
    if (timerSeconds <= 0) return;
    timerRunning = true;
    timerStartBtn.classList.add('hidden');
    timerPauseBtn.classList.remove('hidden');
    timerInterval = setInterval(timerTick, 1000);
}

function pauseTimer() {
    timerRunning = false;
    timerStartBtn.classList.remove('hidden');
    timerPauseBtn.classList.add('hidden');
    clearInterval(timerInterval);
    timerInterval = null;
}

function stopTimer() {
    pauseTimer();
    timerRunning = false;
}

function resetTimer() {
    stopTimer();
    timerSeconds = 300;
    updateTimerDisplay();
}

function toggleTimer() {
    settings.timerVisible = !settings.timerVisible;
    applyTimerVisibility();
    resizeWindow();
    persistSettings();
}

function applyTimerVisibility() {
    if (settings.timerVisible) {
        timerContainer.classList.remove('hidden');
        appWrapper.classList.add('timer-active');
    } else {
        timerContainer.classList.add('hidden');
        appWrapper.classList.remove('timer-active');
        stopTimer();
    }
    updateTimerDisplay();
}

timerStartBtn.addEventListener('click', startTimer);
timerPauseBtn.addEventListener('click', pauseTimer);
timerResetBtn.addEventListener('click', resetTimer);
timerUpBtn.addEventListener('click', () => {
    timerSeconds = Math.min(timerSeconds + 60, 5999); // max 99:59
    updateTimerDisplay();
});
timerDownBtn.addEventListener('click', () => {
    timerSeconds = Math.max(timerSeconds - 60, 0);
    updateTimerDisplay();
});

// === Context Menu ===
clockContainer.addEventListener('contextmenu', (e) => {
    e.preventDefault();
    if (settingsOpen) closeSettingsPanel();
    contextMenu.classList.remove('hidden');
    contextMenu.style.left = e.clientX + 'px';
    contextMenu.style.top = e.clientY + 'px';
    ctxOntop.innerHTML = (isOnTop ? '&#9745;' : '&#9744;') + ' Always on Top';
    ctxTimer.innerHTML = (settings.timerVisible ? '&#9745;' : '&#9744;') + ' Timer';
});

function hideContextMenu() {
    contextMenu.classList.add('hidden');
}

document.addEventListener('click', (e) => {
    if (!contextMenu.contains(e.target)) hideContextMenu();
    // Click outside settings to close (but not if clicking timer or settings itself)
    if (settingsOpen &&
        !settingsPanel.contains(e.target) &&
        e.target.id !== 'settings-btn') {
        closeSettingsPanel();
    }
});

document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        if (settingsOpen) closeSettingsPanel();
        hideContextMenu();
    }
});

document.getElementById('ctx-settings').addEventListener('click', () => {
    hideContextMenu();
    openSettings();
});

ctxTimer.addEventListener('click', () => {
    hideContextMenu();
    toggleTimer();
});

ctxOntop.addEventListener('click', () => {
    isOnTop = !isOnTop;
    settings.alwaysOnTop = isOnTop;
    setAlwaysOnTop(isOnTop);
    hideContextMenu();
    persistSettings();
});

document.getElementById('ctx-exit').addEventListener('click', () => {
    persistSettings().then(() => quit());
});

// === Close Button ===
document.getElementById('close-btn').addEventListener('click', () => {
    persistSettings().then(() => quit());
});

// === Persistence ===
async function savePosition() {
    try {
        const pos = await getPosition();
        const size = await getSize();
        settings.windowX = pos.x;
        settings.windowY = pos.y;
        settings.windowW = size.w;
        // Store the base height (without settings/timer additions)
        baseH = size.h;
        if (settings.timerVisible) baseH -= TIMER_HEIGHT;
        if (settingsOpen) baseH -= SETTINGS_HEIGHT;
        if (baseH < 50) baseH = 100;
        settings.windowH = baseH;
        await persistSettings();
    } catch (e) { /* ignore */ }
}

async function persistSettings() {
    try {
        await SaveSettings(JSON.stringify(settings));
    } catch (err) {
        console.error('Save failed:', err);
    }
}

// === Init ===
async function init() {
    try {
        const saved = await GetSettings();
        if (saved) settings = { ...DEFAULTS, ...JSON.parse(saved) };
    } catch (e) {
        console.log('Using defaults');
    }

    baseH = settings.windowH > 0 ? settings.windowH : DEFAULTS.windowH;

    applyColors();
    setAlwaysOnTop(settings.alwaysOnTop);
    isOnTop = settings.alwaysOnTop;

    if (settings.windowX >= 0 && settings.windowY >= 0) {
        setPosition(settings.windowX, settings.windowY);
    }

    // Apply timer visibility before setting size
    applyTimerVisibility();

    if (settings.windowW > 0) {
        setSize(settings.windowW, calcTotalHeight());
    }

    updateClock();
    setInterval(updateClock, 1000);
}

init();
