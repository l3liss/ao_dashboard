package main

import (
    "bufio"
    "fmt"
    "os"
    "regexp"
    "strconv"
    "strings"
    "time"
)

// Tracker handles reading log files and updating session state
type Tracker struct {
    config Config
    state  *TrackerState
}

// NewTracker constructs a Tracker with provided config and state
func NewTracker(cfg Config, state *TrackerState) *Tracker {
    return &Tracker{config: cfg, state: state}
}

// Start initializes zone and begins tailing logs
func (t *Tracker) Start() {
    t.initZoneFromLog()
    go t.tailFile(t.config.ChatLogPath, t.processChatLine)
    go t.tailFile(t.config.SystemLogPath, t.processSystemLine)
    go t.tailFile(t.config.LootLogPath, t.processLootLine)
}

// initZoneFromLog reads chat log to set initial zone on startup
func (t *Tracker) initZoneFromLog() {
    file, err := os.Open(t.config.ChatLogPath)
    if err != nil {
        return
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        if idx := strings.Index(line, "]"); idx != -1 {
            body := strings.TrimSpace(line[idx+1:])
            if strings.Contains(body, "Entering") {
                parts := strings.SplitN(body, "Entering", 2)
                if len(parts) == 2 {
                    zone := strings.Trim(parts[1], `"' `)
                    t.state.UpdateZone(zone)
                }
            }
        }
    }
}

// tailFile opens a file and continuously reads new lines
func (t *Tracker) tailFile(path string, handler func(string)) {
    for {
        file, err := os.Open(path)
        if err != nil {
            time.Sleep(1 * time.Second)
            continue
        }
        defer file.Close()
        reader := bufio.NewReader(file)
        for {
            rawLine, err := reader.ReadString('\n')
            if err != nil {
                time.Sleep(100 * time.Millisecond)
                continue
            }
            line := strings.TrimSpace(rawLine)
            if line == "" {
                continue
            }
            handler(line)
        }
    }
}

var metaRe = regexp.MustCompile(`\["[^"]*","[^"]*","([^"]*)",\d+\]`)

// processChatLine records chat with player names and zone changes
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
    if name != "" {
        t.state.AddChatMessage(fmt.Sprintf("%s: %s", name, body))
    } else {
        t.state.AddChatMessage(body)
    }
    // Zone detection
    if strings.Contains(body, "Entering") {
        parts := strings.SplitN(body, "Entering", 2)
        if len(parts) == 2 {
            zone := strings.Trim(parts[1], `"' `)
            t.state.UpdateZone(zone)
        }
    }
}

// processSystemLine handles XP, credits, and damage lines
func (t *Tracker) processSystemLine(line string) {
    body := line
    if idx := strings.Index(line, "]"); idx != -1 {
        body = strings.TrimSpace(line[idx+1:])
    }
    if strings.Contains(body, "You received") && strings.Contains(body, "xp") {
        xp := extractFirstNumber(body)
        if xp > 0 {
            t.state.UpdateXP(xp)
        }
    }
    if strings.Contains(body, "You received") && strings.Contains(body, "credits") {
        credits := extractFirstNumber(body)
        if credits > 0 {
            t.state.UpdateCredits(credits)
        }
    }
    if strings.Contains(body, "hit") {
        dmg := extractFirstNumber(body)
        if dmg > 0 {
            t.state.UpdateCrit(dmg)
        }
    }
}

// processLootLine parses HTML-style loot messages
func (t *Tracker) processLootLine(line string) {
    body := line
    if idx := strings.Index(line, "]"); idx != -1 {
        body = strings.TrimSpace(line[idx+1:])
    }
    re := regexp.MustCompile(`(?i)looted <a href="[^"]+">([^<]+)</a>`)  
    if m := re.FindStringSubmatch(body); len(m) > 1 {
        item := strings.TrimSpace(m[1])
        t.state.AddLoot(item)
    }
}

// extractFirstNumber returns the first integer found in text
func extractFirstNumber(text string) int {
    re := regexp.MustCompile(`\d+`)
    match := re.FindString(text)
    if match != "" {
        num, _ := strconv.Atoi(match)
        return num
    }
    return 0
}

