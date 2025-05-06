package main

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "regexp"
    "strconv"
    "strings"
    "time"
)

// Tracker handles reading log files and updating session state
// It watches chat, system, and loot logs to update XP, credits, damage, loot, and zone information.
type Tracker struct {
    config Config
    state  *TrackerState
}

// NewTracker constructs a Tracker with provided config and state
func NewTracker(config Config, state *TrackerState) *Tracker {
    return &Tracker{config: config, state: state}
}

// Start prints paths, initializes zone, and begins tailing logs
func (t *Tracker) Start() {
    fmt.Printf("[TRACKER] Watching logs → chat: %s | system: %s | loot: %s\n",
        t.config.ChatLogPath, t.config.SystemLogPath, t.config.LootLogPath)
    t.initZoneFromLog()

    go t.tailFile(t.config.ChatLogPath, t.processChatLine)
    go t.tailFile(t.config.ChatLogPath, t.processSystemLine)
    go t.tailFile(t.config.SystemLogPath, t.processSystemLine)
    go t.tailFile(t.config.LootLogPath, t.processLootLine)
}

// tailFile streams only new lines from the given file path into handler
func (t *Tracker) tailFile(path string, handler func(string)) {
    for {
        f, err := os.Open(path)
        if err != nil {
            fmt.Printf("[TRACKER] open error (%s): %v\n", path, err)
            time.Sleep(1 * time.Second)
            continue
        }
        // Seek to end so only new entries are read
        if _, err := f.Seek(0, io.SeekEnd); err != nil {
            fmt.Printf("[TRACKER] seek error (%s): %v\n", path, err)
        }
        reader := bufio.NewReader(f)
        for {
            raw, err := reader.ReadString('\n')
            if err != nil {
                time.Sleep(100 * time.Millisecond)
                continue
            }
            handler(strings.TrimSpace(raw))
        }
    }
}

// readLastLines reads up to the last n lines from a file (used for initial zone)
func readLastLines(path string, n int) ([]string, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    var lines []string
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
        if len(lines) > n {
            lines = lines[1:]
        }
    }
    return lines, scanner.Err()
}

// parseZone extracts the zone name from a line containing "Entering"
func parseZone(line string) string {
    if idx := strings.Index(line, "Entering"); idx != -1 {
        raw := line[idx+len("Entering"):]
        return strings.Trim(raw, `"' .`)
    }
    return ""
}

// initZoneFromLog sets initial zone by scanning recent log history
func (t *Tracker) initZoneFromLog() {
    for _, path := range []string{t.config.ChatLogPath, t.config.SystemLogPath} {
        if lines, err := readLastLines(path, 50); err == nil {
            for i := len(lines) - 1; i >= 0; i-- {
                if z := parseZone(lines[i]); z != "" {
                    t.state.UpdateZone(z)
                    fmt.Printf("[TRACKER] Initial zone → %s\n", z)
                    return
                }
            }
        }
    }
}

var (
    metaRe = regexp.MustCompile(`\["[^"]*","[^"]*","([^"]*)",\d+\]`)
    lootRe = regexp.MustCompile(`<a href="itemref://\d+/\d+/\d+">(.*?)</a>`)
)

// processChatLine handles chat; coordinates end-of-combat logic and updates zone/messages
func (t *Tracker) processChatLine(line string) {
    name := ""
    body := line
    if idx := strings.Index(line, "]"); idx != -1 {
        meta := line[:idx+1]
        body = strings.TrimSpace(line[idx+1:])
        if m := metaRe.FindStringSubmatch(meta); len(m) > 1 {
            name = m[1]
        }
    }
    cleaned := cleanHTML(body)

    // End-of-combat trigger: wait 6s after loot remains message
    if strings.Contains(cleaned, "You can loot these remains.") {
        if t.state.CombatStartTime != 0 {
            endTriggerTime := time.Now().UnixMilli()
            go func(remainsTime int64) {
                time.Sleep(6 * time.Second)
                t.state.mutex.Lock()
                defer t.state.mutex.Unlock()
                // Confirm no new hits occurred since remainsTime
                if t.state.CombatStartTime != 0 && t.state.LastHitTime <= remainsTime {
                    // Calculate encounter DPS
                    durMs := t.state.LastHitTime - t.state.CombatStartTime
                    if durMs < 1 {
                        durMs = 1
                    }
                    encounterDPS := float64(t.state.TotalDamage) / (float64(durMs) / 1000.0)
                    // Append to history, capped at maxCombats
                    t.state.CombatDPSHistory = append(t.state.CombatDPSHistory, encounterDPS)
                    if len(t.state.CombatDPSHistory) > maxCombats {
                        t.state.CombatDPSHistory = t.state.CombatDPSHistory[len(t.state.CombatDPSHistory)-maxCombats:]
                    }
                    // Reset combat state
                    t.state.CombatStartTime = 0
                    t.state.LastHitTime = 0
                    t.state.TotalDamage = 0
                }
            }(endTriggerTime)
        }
    }

    // Zone update
    if z := parseZone(cleaned); z != "" {
        t.state.UpdateZone(z)
        fmt.Printf("[TRACKER] Zone updated → %s\n", z)
    }
    // Record chat line
    if name != "" {
        t.state.AddChatMessage(fmt.Sprintf("[%s] %s", name, cleaned))
    }
}

// processSystemLine handles XP, credits, and damage hits
func (t *Tracker) processSystemLine(line string) {
    body := line
    if idx := strings.Index(line, "]"); idx != -1 {
        body = strings.TrimSpace(line[idx+1:])
    }
    // XP/credits parsing remains the same
    if strings.Contains(body, "You received") {
        if strings.Contains(body, "xp") {
            if xp := extractFirstNumber(body); xp > 0 {
                t.state.UpdateXP(xp)
            }
        }
        if strings.Contains(body, "credits") {
            if cr := extractFirstNumber(body); cr > 0 {
                t.state.UpdateCredits(cr)
            }
        }
    }
    // Damage parsing: only count player-originated hits
    if strings.HasPrefix(body, "You hit") || strings.HasPrefix(body, "You critically hit") {
        if dmg := extractFirstNumber(body); dmg > 0 {
            t.state.UpdateCrit(dmg)
        }
    }
}

// processLootLine captures new loot only and saves immediately
func (t *Tracker) processLootLine(line string) {
    fmt.Printf("[TRACKER] Loot event → %s\n", line)
    for _, m := range lootRe.FindAllStringSubmatch(line, -1) {
        item := cleanHTML(m[1])
        fmt.Printf("[TRACKER] Loot detected → %s\n", item)
        t.state.AddLoot(item)
        if err := t.state.SaveToFile(t.config.StateFilePath); err != nil {
            fmt.Printf("[TRACKER] Save error → %v\n", err)
        }
    }
}

// extractFirstNumber returns first integer found in the string
func extractFirstNumber(s string) int {
    re := regexp.MustCompile(`\d+`)
    if m := re.FindString(s); m != "" {
        n, _ := strconv.Atoi(m)
        return n
    }
    return 0
}

// cleanHTML strips HTML tags and normalizes whitespace
func cleanHTML(input string) string {
    re := regexp.MustCompile(`<[^>]+>`)
    t := re.ReplaceAllString(input, "")
    re2 := regexp.MustCompile(`\s+`)
    return strings.TrimSpace(re2.ReplaceAllString(t, " "))
}

