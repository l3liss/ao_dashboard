from PyQt6.QtWidgets import (
    QWidget, QVBoxLayout, QLabel, QListWidget, QProgressBar,
    QTextEdit, QHBoxLayout, QSizePolicy, QApplication
)
from PyQt6.QtCore import QTimer
from PyQt6.QtGui import QFont, QColor, QPalette
import reader

STATE_FILE = "../shared/state.json"

class MainWindow(QWidget):
    def __init__(self):
        super().__init__()
        self.setWindowTitle("AO Dashboard")
        self.setGeometry(100, 100, 600, 800)

        self.init_ui()
        self.apply_dark_theme()

        self.timer = QTimer()
        self.timer.timeout.connect(self.update_ui)
        self.timer.start(500)  # Refresh every 500 ms

    def init_ui(self):
        font = QFont("Arial", 10)
        self.setFont(font)

        layout = QVBoxLayout()

        # Zone and latency
        self.zone_label = QLabel("Zone: Unknown")
        layout.addWidget(self.zone_label)

        self.latency_label = QLabel("Latency: 0 ms")
        layout.addWidget(self.latency_label)

        # XP Progress
        self.xp_progress = QProgressBar()
        self.xp_progress.setMaximum(100)
        layout.addWidget(self.xp_progress)

        # Credits
        self.credits_label = QLabel("Credits: 0")
        layout.addWidget(self.credits_label)

        # Crits
        self.latest_crit_label = QLabel("Latest Crit: 0")
        self.biggest_crit_label = QLabel("Biggest Crit: 0")
        layout.addWidget(self.latest_crit_label)
        layout.addWidget(self.biggest_crit_label)

        # Recent Loot
        self.loot_list = QListWidget()
        self.loot_list.setSizePolicy(QSizePolicy.Policy.Expanding, QSizePolicy.Policy.MinimumExpanding)
        layout.addWidget(QLabel("Recent Loot:"))
        layout.addWidget(self.loot_list)

        # Chat Log
        self.chat_box = QTextEdit()
        self.chat_box.setReadOnly(True)
        self.chat_box.setMaximumHeight(300)
        layout.addWidget(QLabel("Chat Log:"))
        layout.addWidget(self.chat_box)

        self.setLayout(layout)

    def apply_dark_theme(self):
        dark_palette = QPalette()
        dark_palette.setColor(QPalette.ColorRole.Window, QColor(30, 30, 30))
        dark_palette.setColor(QPalette.ColorRole.WindowText, QColor(220, 220, 220))
        dark_palette.setColor(QPalette.ColorRole.Base, QColor(20, 20, 20))
        dark_palette.setColor(QPalette.ColorRole.AlternateBase, QColor(30, 30, 30))
        dark_palette.setColor(QPalette.ColorRole.ToolTipBase, QColor(255, 255, 255))
        dark_palette.setColor(QPalette.ColorRole.ToolTipText, QColor(255, 255, 255))
        dark_palette.setColor(QPalette.ColorRole.Text, QColor(220, 220, 220))
        dark_palette.setColor(QPalette.ColorRole.Button, QColor(45, 45, 45))
        dark_palette.setColor(QPalette.ColorRole.ButtonText, QColor(220, 220, 220))
        dark_palette.setColor(QPalette.ColorRole.Highlight, QColor(0, 120, 215))
        dark_palette.setColor(QPalette.ColorRole.HighlightedText, QColor(0, 0, 0))
        self.setPalette(dark_palette)

    def update_ui(self):
        state = reader.load_state(STATE_FILE)
        if not state:
            return

        self.zone_label.setText(f"Zone: {state.get('zone', 'Unknown')}")
        self.latency_label.setText(f"Latency: {state.get('latency_ms', 0)} ms")

        xp = state.get("xp", 0)
        # Simulated XP percentage
        xp_percent = min(int((xp % 100000) / 1000), 100)
        self.xp_progress.setValue(xp_percent)

        self.credits_label.setText(f"Credits: {state.get('credits', 0)}")
        self.latest_crit_label.setText(f"Latest Crit: {state.get('latest_crit', 0)}")
        self.biggest_crit_label.setText(f"Biggest Crit: {state.get('biggest_crit', 0)}")

        self.loot_list.clear()
        for item in reversed(state.get("recent_loot", [])[-10:]):
            self.loot_list.addItem(item)

        self.chat_box.clear()
        for line in reversed(state.get("chat_history", [])[-50:]):
            self.chat_box.append(line)
