package main

import (
    "encoding/json"
    "os"
    "sync"
    "time"
)

// SessionState holds live session information
type SessionState struct {
    XP            int      `json:"xp"`
    Level         int      `json:"level"`
    Credits       int      `json:"credits"`
    Zone          string   `json:"zone"`
    ChatHistory   []string `json:"chat_history"`
    RecentLoot    []string `json:"recent_loot"`
    BiggestCrit   int      `json:"biggest_crit"`
    LatestCrit    int      `json:"latest_crit"`
    LatencyMS     int      `json:"latency_ms"`
    PlayersOnline int      `json:"players_online"`
    StartTime     int64    `json:"start_time"` // UNIX timestamp
}

// TrackerState manages the session state with locking
type TrackerState struct {
    State SessionState
    mu    sync.Mutex
}

// NewTrackerState initializes a fresh TrackerState
func NewTrackerState() *TrackerState {
    return &TrackerState{
        State: SessionState{
            XP:            0,
            Level:         1,
            Credits:       0,
            Zone:          "Unknown",
            ChatHistory:   []string{},
            RecentLoot:    []string{},
            BiggestCrit:   0,
            LatestCrit:    0,
            LatencyMS:     0,
            PlayersOnline: 0,
            StartTime:     time.Now().Unix(),
        },
    }
}

// SaveToFile saves the current session state to JSON
func (ts *TrackerState) SaveToFile(filePath string) error {
    ts.mu.Lock()
    defer ts.mu.Unlock()

    file, err := os.Create(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ") // Pretty print
    return encoder.Encode(ts.State)
}

// Update functions: these safely update parts of the state

func (ts *TrackerState) AddChatMessage(msg string) {
    ts.mu.Lock()
    defer ts.mu.Unlock()

    ts.State.ChatHistory = append(ts.State.ChatHistory, msg)
    if len(ts.State.ChatHistory) > 100 {
        ts.State.ChatHistory = ts.State.ChatHistory[len(ts.State.ChatHistory)-100:]
    }
}

func (ts *TrackerState) AddLoot(item string) {
    ts.mu.Lock()
    defer ts.mu.Unlock()

    ts.State.RecentLoot = append(ts.State.RecentLoot, item)
    if len(ts.State.RecentLoot) > 20 {
        ts.State.RecentLoot = ts.State.RecentLoot[len(ts.State.RecentLoot)-20:]
    }
}

func (ts *TrackerState) UpdateLatency(ms int) {
    ts.mu.Lock()
    defer ts.mu.Unlock()

    ts.State.LatencyMS = ms
}

func (ts *TrackerState) UpdateXP(xp int) {
    ts.mu.Lock()
    defer ts.mu.Unlock()

    ts.State.XP += xp
}

func (ts *TrackerState) UpdateCredits(credits int) {
    ts.mu.Lock()
    defer ts.mu.Unlock()

    ts.State.Credits += credits
}

func (ts *TrackerState) UpdateZone(zone string) {
    ts.mu.Lock()
    defer ts.mu.Unlock()

    ts.State.Zone = zone
}

func (ts *TrackerState) UpdateCrit(damage int) {
    ts.mu.Lock()
    defer ts.mu.Unlock()

    ts.State.LatestCrit = damage
    if damage > ts.State.BiggestCrit {
        ts.State.BiggestCrit = damage
    }
}
