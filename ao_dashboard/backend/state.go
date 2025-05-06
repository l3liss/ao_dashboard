package main

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "sync"
    "time"
)

// TrackerState holds all tracked session data
// Includes stats, chat/loot history, and custom DPS/combat calculations.
type TrackerState struct {
    XP               int       `json:"xp"`
    Level            int       `json:"level"`
    Credits          int       `json:"credits"`
    Zone             string    `json:"zone"`
    ChatHistory      []string  `json:"chat_history"`
    RecentLoot       []string  `json:"recent_loot"`
    BiggestCrit      int       `json:"biggest_crit"`
    LatestCrit       int       `json:"latest_crit"`
    LatencyMS        int       `json:"latency_ms"`
    PlayersOnline    int       `json:"players_online"`
    StartTime        int64     `json:"start_time"`

    // Combat tracking
    CombatStartTime  int64     `json:"-"` // time of first hit in current combat (ms)
    TotalDamage      int       `json:"-"` // sum of damage in current combat
    LastHitTime      int64     `json:"-"` // time of last hit (ms)
    CombatDPSHistory []float64 `json:"-"` // recent combat DPS values

    // Persisted DPS metrics
    LastCombatDPS    int       `json:"dps_12s"`    // current combat DPS (first to last hit)
    LastSessionDPS   int       `json:"dps_session"` // average over recent combats

    mutex            sync.Mutex
}

const (
    maxChatLines = 50
    maxLootItems = 20
    maxCombats   = 30
)

// NewTrackerState loads existing state file or returns a new state
func NewTrackerState(path string) *TrackerState {
    state, err := LoadTrackerState(path)
    if err != nil {
        fmt.Println("[STATE] initializing new state:", err)
        state = &TrackerState{
            Level:            1,
            Zone:             "Unknown",
            StartTime:        time.Now().Unix(),
            ChatHistory:      make([]string, 0, maxChatLines),
            RecentLoot:       make([]string, 0, maxLootItems),
            CombatDPSHistory: make([]float64, 0, maxCombats),
        }
    }
    if state.StartTime == 0 {
        state.StartTime = time.Now().Unix()
    }
    return state
}

// LoadTrackerState reads JSON state from disk
func LoadTrackerState(path string) (*TrackerState, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    data, err := io.ReadAll(f)
    if err != nil {
        return nil, err
    }
    var state TrackerState
    if err := json.Unmarshal(data, &state); err != nil {
        return nil, err
    }
    return &state, nil
}

// SaveToFile writes state back to disk atomically
func (s *TrackerState) SaveToFile(path string) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    tmp := path + ".tmp"
    f, err := os.Create(tmp)
    if err != nil {
        return err
    }
    enc := json.NewEncoder(f)
    enc.SetIndent("", "  ")
    if err := enc.Encode(s); err != nil {
        f.Close()
        return err
    }
    f.Close()
    return os.Rename(tmp, path)
}

// StartAutoSave begins periodic DPS updates and disk saves
func (s *TrackerState) StartAutoSave(path string) {
    go func() {
        for {
            s.UpdateDPS()
            if err := s.SaveToFile(path); err != nil {
                fmt.Println("[STATE] save error:", err)
            }
            time.Sleep(1 * time.Second)
        }
    }()
}

// UpdateDPS recalculates current combat and session DPS
func (s *TrackerState) UpdateDPS() {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    // Current combat DPS: from first hit to last hit
    current := 0
    if s.CombatStartTime > 0 {
        durationMs := s.LastHitTime - s.CombatStartTime
        if durationMs < 1 {
            durationMs = 1
        }
        // convert ms to seconds for DPS calculation
        current = int(float64(s.TotalDamage) / (float64(durationMs) / 1000.0))
    }
    s.LastCombatDPS = current

    // Session DPS: average over stored combat DPS values
    if len(s.CombatDPSHistory) > 0 {
        var sum float64
        for _, d := range s.CombatDPSHistory {
            sum += d
        }
        avg := sum / float64(len(s.CombatDPSHistory))
        s.LastSessionDPS = int(avg)
    } else {
        s.LastSessionDPS = current
    }
}

// UpdateCrit processes a damage hit and accumulates combat damage
func (s *TrackerState) UpdateCrit(damage int) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    now := time.Now().UnixMilli()

    // Track crit values
    s.LatestCrit = damage
    if damage > s.BiggestCrit {
        s.BiggestCrit = damage
    }

    // Start new combat if none active
    if s.CombatStartTime == 0 {
        s.CombatStartTime = now
        s.LastHitTime = now
        s.TotalDamage = damage
        return
    }

    // Continue current combat: update last hit and accumulate damage
    s.LastHitTime = now
    s.TotalDamage += damage
}

// UpdateXP increments the XP counter
func (s *TrackerState) UpdateXP(xp int) {
    s.mutex.Lock()
    s.XP += xp
    s.mutex.Unlock()
}

// UpdateCredits increments the credits counter
func (s *TrackerState) UpdateCredits(c int) {
    s.mutex.Lock()
    s.Credits += c
    s.mutex.Unlock()
}

// UpdateZone sets the current zone
func (s *TrackerState) UpdateZone(zone string) {
    s.mutex.Lock()
    if zone != "" {
        s.Zone = zone
    }
    s.mutex.Unlock()
}

// UpdateLatency sets the latest latency measurement
func (s *TrackerState) UpdateLatency(ms int) {
    s.mutex.Lock()
    if ms >= 0 {
        s.LatencyMS = ms
    }
    s.mutex.Unlock()
}

// AddChatMessage appends a chat line to the history buffer
func (s *TrackerState) AddChatMessage(msg string) {
    s.mutex.Lock()
    s.ChatHistory = append(s.ChatHistory, msg)
    if len(s.ChatHistory) > maxChatLines {
        s.ChatHistory = s.ChatHistory[len(s.ChatHistory)-maxChatLines:]
    }
    s.mutex.Unlock()
}

// AddLoot appends a loot item, ignoring "Lootable Corpse"
func (s *TrackerState) AddLoot(item string) {
    if item == "Lootable Corpse" {
        return
    }
    s.mutex.Lock()
    s.RecentLoot = append(s.RecentLoot, item)
    if len(s.RecentLoot) > maxLootItems {
        s.RecentLoot = s.RecentLoot[len(s.RecentLoot)-maxLootItems:]
    }
    s.mutex.Unlock()
}

