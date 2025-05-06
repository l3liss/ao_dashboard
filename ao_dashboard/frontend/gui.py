from PyQt6.QtWidgets import (
    QWidget, QVBoxLayout, QLabel, QListWidget, QProgressBar,
    QTextEdit, QSizePolicy, QApplication, QSplitter
)
from PyQt6.QtCore import QTimer, Qt
from PyQt6.QtGui import QFont, QColor, QPalette
import reader
import time
import re
from html import escape

# Replace this with your character's exact name as it appears in chat metadata
MY_NAME = "Rustbank"

# Determine config.json path: try frontend then backend
import json, os
BASE_DIR = os.path.dirname(__file__)
CANDIDATE = [
    os.path.join(BASE_DIR, "config.json"),
    os.path.join(BASE_DIR, "..", "backend", "config.json"),
]
CONFIG_PATH = None
for path in CANDIDATE:
    if os.path.exists(path):
        CONFIG_PATH = path
        break
if not CONFIG_PATH:
    print(f"[GUI] Warning: config.json not found in frontend or backend directories.")
    STATE_FILE = "../shared/state.json"
else:
    try:
        with open(CONFIG_PATH, "r") as cfgf:
            cfg = json.load(cfgf)
            STATE_FILE = cfg.get("state_file_path", cfg.get("StateFilePath", "../shared/state.json"))
    except Exception as e:
        print(f"[GUI] Warning: failed to load config.json at {CONFIG_PATH}: {e}")
        STATE_FILE = "../shared/state.json"

class MainWindow(QWidget):
    def __init__(self):
        super().__init__()
        self.setWindowTitle("AO Dashboard")
        self.setGeometry(100, 100, 700, 900)

        self.last_xp = 0
        self.last_credits = 0
        self.last_crit = 0
        self.last_biggest_crit = 0

        self.init_ui()
        self.apply_dark_theme()

        self.timer = QTimer()
        self.timer.timeout.connect(self.update_ui)
        self.timer.start(500)

    def init_ui(self):
        font = QFont("Consolas", 10)
        self.setFont(font)

        layout = QVBoxLayout(self)

        # Top stats
        self.zone_label = QLabel("Zone: Unknown")
        self.latency_label = QLabel("Latency: 0 ms")
        self.session_label = QLabel("Session: 0 XP/hr | 0 Credits/hr")
        layout.addWidget(self.zone_label)
        layout.addWidget(self.latency_label)
        layout.addWidget(self.session_label)

        # XP, credits, crits, DPS
        self.xp_progress = QProgressBar()
        self.xp_progress.setMaximum(100)
        self.credits_label = QLabel("Credits: 0")
        self.latest_crit_label = QLabel("Latest Crit: 0")
        self.biggest_crit_label = QLabel("Biggest Crit: 0")
        self.dps_burst_label = QLabel("Burst DPS (4s): 0")
        self.dps_session_label = QLabel("Session DPS: 0")

        layout.addWidget(QLabel("XP Progress:"))
        layout.addWidget(self.xp_progress)
        layout.addWidget(self.credits_label)
        layout.addWidget(self.latest_crit_label)
        layout.addWidget(self.biggest_crit_label)
        layout.addWidget(self.dps_burst_label)
        layout.addWidget(self.dps_session_label)

        # Splitter for loot and chat
        self.splitter = QSplitter(Qt.Orientation.Vertical)

        # Loot section
        loot_container = QWidget()
        loot_layout = QVBoxLayout(loot_container)
        loot_layout.addWidget(QLabel("Recent Loot:"))
        self.loot_list = QListWidget()
        loot_layout.addWidget(self.loot_list)

        # Chat section
        chat_container = QWidget()
        chat_layout = QVBoxLayout(chat_container)
        chat_layout.addWidget(QLabel("Chat History:"))
        self.chat_box = QTextEdit()
        self.chat_box.setReadOnly(True)
        self.chat_box.setSizePolicy(QSizePolicy.Policy.Expanding, QSizePolicy.Policy.Expanding)
        chat_layout.addWidget(self.chat_box)

        self.splitter.addWidget(loot_container)
        self.splitter.addWidget(chat_container)
        layout.addWidget(self.splitter)

    def apply_dark_theme(self):
        palette = QPalette()
        palette.setColor(QPalette.ColorRole.Window, QColor("#1a1b24"))
        palette.setColor(QPalette.ColorRole.Base, QColor("#1a1b24"))
        palette.setColor(QPalette.ColorRole.WindowText, QColor("#e2e2dc"))
        palette.setColor(QPalette.ColorRole.Text, QColor("#e2e2dc"))
        palette.setColor(QPalette.ColorRole.Button, QColor("#1a1b24"))
        palette.setColor(QPalette.ColorRole.ButtonText, QColor("#e2e2dc"))
        palette.setColor(QPalette.ColorRole.AlternateBase, QColor("#333344"))
        palette.setColor(QPalette.ColorRole.Highlight, QColor("#333344"))
        palette.setColor(QPalette.ColorRole.HighlightedText, QColor("#e2e2dc"))
        palette.setColor(QPalette.ColorRole.PlaceholderText, QColor("#5a5a7d"))
        palette.setColor(QPalette.ColorRole.Link, QColor("#5da9a8"))
        palette.setColor(QPalette.ColorRole.LinkVisited, QColor("#8a6fb5"))
        self.setPalette(palette)

    def get_name_color(self, name: str) -> str:
        return "#3faf6f" if name == MY_NAME else "#5da9a8"

    def update_ui(self):
        state = reader.load_state(STATE_FILE)
        if not state:
            return

        # Update zone and latency
        self.zone_label.setText(f"Zone: {state.get('zone', 'Unknown')}")
        self.latency_label.setText(f"Latency: {state.get('latency_ms', 0)} ms")

        # XP and credits
        xp = state.get('xp', 0)
        credits = state.get('credits', 0)
        self.xp_progress.setValue(xp % 100)
        self.credits_label.setText(f"Credits: {credits}")

        # Crits
        self.latest_crit_label.setText(f"Latest Crit: {state.get('latest_crit', 0)}")
        self.biggest_crit_label.setText(f"Biggest Crit: {state.get('biggest_crit', 0)}")

        # DPS
        self.dps_burst_label.setText(f"Burst DPS (4s): {state.get('dps_12s', 0)}")
        self.dps_session_label.setText(f"Session DPS: {state.get('dps_session', 0)}")

        # Session rates
        start_time = state.get('start_time', time.time())
        elapsed = max(time.time() - start_time, 1)
        xp_hr = int((xp / elapsed) * 3600)
        cr_hr = int((credits / elapsed) * 3600)
        self.session_label.setText(f"Session: {xp_hr} XP/hr | {cr_hr} Credits/hr")

        # Loot
        self.loot_list.clear()
        for item in reversed(state.get('recent_loot', [])[-10:]):
            self.loot_list.addItem(item)

        # Chat with colors, but only auto-scroll if at bottom
        chat_lines = state.get('chat_history', [])[-50:]
        html_parts = []
        for ln in chat_lines:
            esc = escape(ln)
            if 'Attacked by' in ln:
                html_parts.append(f'<span style="color:#cc4444">{esc}</span>')
            elif 'credits' in ln.lower():
                html_parts.append(f'<span style="color:#e0c96f">{esc}</span>')
            else:
                m = re.findall(r'^(.*?):\s*(.*)$', ln)
                if m:
                    name, msg = m[0]
                    html_parts.append(
                        f'<b><span style="color:{self.get_name_color(name)}">{name}:</span></b> {escape(msg)}'
                    )
                else:
                    html_parts.append(esc)
        new_html = '<br>'.join(html_parts)

        sb = self.chat_box.verticalScrollBar()
        if sb.value() == sb.maximum():
            self.chat_box.setHtml(new_html)
            sb.setValue(sb.maximum())

if __name__ == "__main__":
    app = QApplication([])
    w = MainWindow()
    w.show()
    app.exec()

