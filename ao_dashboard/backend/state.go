package main

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "sync"
    "time"
)

// TimedHit represents a timestamped damage event
type TimedHit struct {
    Time   int64
    Amount int
}

// TrackerState holds all tracked session data
type TrackerState struct {
    XP              int      `json:"xp"`
    Level           int      `json:"level"`
    Credits         int      `json:"credits"`
    Zone            string   `json:"zone"`
    ChatHistory     []string `json:"chat_history"`
    RecentLoot      []string `json:"recent_loot"`
    BiggestCrit     int      `json:"biggest_crit"`
    LatestCrit      int      `json:"latest_crit"`
    LatencyMS       int      `json:"latency_ms"`
    PlayersOnline   int      `json:"players_online"`
    StartTime       int64    `json:"start_time"`

    // Internal, not serialized
    DamageHistory   []TimedHit `json:"-"`
    TotalDamage     int        `json:"-"`
    CombatStartTime int64      `json:"-"`

    // Persisted DPS metrics
    LastBurstDPS    int        `json:"dps_12s"`
    LastSessionDPS  int        `json:"dps_session"`

    mutex           sync.Mutex
}

const (
    maxChatLines = 50
    maxLootItems = 20
)

// NewTrackerState loads existing state or returns defaults
func NewTrackerState() *TrackerState {
    state, err := LoadTrackerState("../shared/state.json")
    if err != nil {
        fmt.Println("[STATE] initializing new state:", err)
        return &TrackerState{
            Level:          1,
            Zone:           "Unknown",
            StartTime:      time.Now().Unix(),
            ChatHistory:    make([]string, 0, maxChatLines),
            RecentLoot:     make([]string, 0, maxLootItems),
            LastBurstDPS:   0,
            LastSessionDPS: 0,
        }
    }
    if state.StartTime == 0 {
        state.StartTime = time.Now().Unix()
    }
    return state
}

// LoadTrackerState reads state from JSON file
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

// SaveToFile writes state atomically to disk
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

// StartAutoSave begins periodic state saves and DPS updates
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

// UpdateDPS recalculates burst and session DPS
func (s *TrackerState) UpdateDPS() {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.LastBurstDPS = s.calculateBurstDPSLocked()
    s.LastSessionDPS = s.calculateSessionDPSLocked()
}

// UpdateXP safely increments XP
func (s *TrackerState) UpdateXP(xp int) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.XP += xp
}

// UpdateCredits safely increments credits
func (s *TrackerState) UpdateCredits(c int) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.Credits += c
}

// UpdateZone sets the current zone
func (s *TrackerState) UpdateZone(zone string) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    if zone != "" {
        s.Zone = zone
    }
}

// UpdateLatency records ping latency
func (s *TrackerState) UpdateLatency(ms int) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    if ms >= 0 {
        s.LatencyMS = ms
    }
}

// UpdateCrit records a damage hit and updates history
func (s *TrackerState) UpdateCrit(damage int) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    now := time.Now().Unix()

    s.LatestCrit = damage
    if damage > s.BiggestCrit {
        s.BiggestCrit = damage
    }

    // reset session if idle >60s
    if s.CombatStartTime == 0 || now-s.CombatStartTime > 60 {
        s.CombatStartTime = now
        s.TotalDamage = 0
    }
    s.TotalDamage += damage

    // append and prune damage history older than 4s
    s.DamageHistory = append(s.DamageHistory, TimedHit{Time: now, Amount: damage})
    cutoff := now - 4
    i := 0
    for _, hit := range s.DamageHistory {
        if hit.Time >= cutoff {
            s.DamageHistory[i] = hit
            i++
        }
    }
    s.DamageHistory = s.DamageHistory[:i]
}

// calculateBurstDPSLocked computes DPS over last 4s using actual time span
func (s *TrackerState) calculateBurstDPSLocked() int {
    now := time.Now().Unix()
    window := int64(4)
    var total int
    var earliest int64 = now
    for _, hit := range s.DamageHistory {
        if hit.Time >= now-window {
            total += hit.Amount
            if hit.Time < earliest {
                earliest = hit.Time
            }
        }
    }
    if total == 0 {
        return s.LastBurstDPS
    }
    duration := now - earliest
    if duration < 1 {
        duration = 1
    }
    return int(float64(total) / float64(duration))
}

// calculateSessionDPSLocked computes average DPS since combat start
func (s *TrackerState) calculateSessionDPSLocked() int {
    if s.CombatStartTime == 0 {
        return s.LastSessionDPS
    }
    elapsed := time.Now().Unix() - s.CombatStartTime
    if elapsed < 1 {
        return s.LastSessionDPS
    }
    return int(float64(s.TotalDamage) / float64(elapsed))
}

// AddChatMessage appends a chat line
func (s *TrackerState) AddChatMessage(msg string) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.ChatHistory = append(s.ChatHistory, msg)
    if len(s.ChatHistory) > maxChatLines {
        s.ChatHistory = s.ChatHistory[len(s.ChatHistory)-maxChatLines:]
    }
}

// AddLoot appends a loot item, ignoring "Lootable Corpse"
func (s *TrackerState) AddLoot(item string) {
    if item == "Lootable Corpse" {
        return
    }
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.RecentLoot = append(s.RecentLoot, item)
    if len(s.RecentLoot) > maxLootItems {
        s.RecentLoot = s.RecentLoot[len(s.RecentLoot)-maxLootItems:]
    }
}

