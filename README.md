# Anarchy Online ― External Dashboard & Log Toolkit

> **Status:** Early Alpha  |  **Audience:** Froob‑friendly traders & AO tinkerers  |  **License:** MIT

A linux based, open‑source overlay that turns Anarchy Online chat logs into **live combat, loot, and economy insights**.
Built as a split‑stack app for second monitor viewing of game statistics.

* **Backend:** Go 1.22 — tails logfiles, parses events, maintains a JSON state & REST‑ish API.
* **Frontend:** Python 3.12 / PyQt6 — reactive GUI that renders DPS meters, loot feeds, credit & XP trackers.

---

## ✨ Core Features

| Module              | What it Does                              | Files                                   |
| ------------------- | ----------------------------------------- | --------------------------------------- |
| **Combat Tracker**  | Calculates DPS, crit rates, combat uptime | `backend/tracker.go`                    |
| **State Manager**   | Persists session data, autosaves JSON     | `backend/state.go`, `shared/state.json` |
| **Latency Pinger**  | \~1 s interval ICMP ping of AO servers    | `backend/pinger.go`                     |
| **GUI**             | Live panels (DPS, Loot, XP, Credits)      | `frontend/gui.py`                       |

---

## 📂 Project Layout 

```text
anarchy/
├── ao_dashboard
│   ├── backend
│   │   ├── config.go          # loads config.json, env overrides
│   │   ├── config.json        # sample paths & tunables
│   │   ├── go.mod
│   │   ├── main.go            # CLI entry; wires everything
│   │   ├── pinger.go          # latency probe goroutine
│   │   ├── state.go           # session state struct + autosave
│   │   └── tracker.go         # log parser & combat maths
│   ├── frontend
│   │   ├── gui.py             # PyQt6 widgets
│   │   ├── main.py            # launches Qt event‑loop
│   │   ├── reader.py          # polling JSON → signals
│   │   └── requirements.txt   # pip deps (PyQt6, pydantic, etc.)
│   ├── run.sh                 # convenience launcher (builds Go, starts GUI)
│   └── shared
│       └── state.json         # live+autosaved session data
├── ao_logs                    # symlinks to actual Steam/Proton logs
│   ├── chat.log   -> ...Window1/Log.txt
│   ├── combat.log -> ...Window2/Log.txt
│   └── loot.log   -> ...Window3/Log.txt
└── README.md                  # you are here
```

---

## 🚀 Quick Start

```bash
# 1. Clone (HTTPS/GPG‑signed commit friendly)
$ git clone https://github.com/YOURNAME/anarchy-dashboard.git
$ cd anarchy-dashboard/anarchy

# 2. Build backend (requires Go 1.22+)
$ cd ao_dashboard/backend && go build -o ../../bin/ao_backend && cd ../..

# 3. Install frontend deps (Python 3.12)
$ python -m venv .venv && source .venv/bin/activate
$ pip install -r ao_dashboard/frontend/requirements.txt

# 4. Update paths in ao_dashboard/backend/config.json (or export env vars)

# 5. Launch
$ ./ao_dashboard/run.sh
```

`run.sh` will: ① build/refresh the Go binary, ② start it in the background, ③ spawn the PyQt GUI.

---

## 🔧 Configuration Cheatsheet (`config.json`)

| Key             | Default              | Notes                             |
| --------------- | -------------------- | --------------------------------- |
| `CombatLogPath` | `ao_logs/combat.log` | Follows symlink; any file is fine |
| `ChatLogPath`   | `ao_logs/chat.log`   | Used for XP/Credit events         |
| `LootLogPath`   | `ao_logs/loot.log`   | Loot feed                         |
| `AutosaveSecs`  | `15`                 | Interval for `shared/state.json`  |
| `PingHost`      | `chat.d1.funcom.com` | EU/US login servers also work     |

All keys can be overridden by env vars: `AO_<KEY>`, e.g. `AO_PingHost=1.2.3.4`.

---

## 📜 License

This project is licensed under the MIT License — see [`LICENSE`](LICENSE) for details.

> **Legal Note:** Anarchy Online is a trademark of Funcom.  This project is third‑party, unofficial, and follows the AO EULA by parsing plain‑text chat logs only.
