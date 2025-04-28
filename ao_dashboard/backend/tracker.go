package main

import (
    "bufio"
    "os"
    "regexp"
    "strconv"
    "strings"
    "time"
)

// Tracker handles reading log files and updating the session state
type Tracker struct {
    config Config
    state  *TrackerState
}

// NewTracker creates a new Tracker
func NewTracker(cfg Config, state *TrackerState) *Tracker {
    return &Tracker{
        config: cfg,
        state:  state,
    }
}

// Start begins tailing the logs
func (t *Tracker) Start() {
    go t.tailFile(t.config.ChatLogPath, t.processChatLine)
    go t.tailFile(t.config.SystemLogPath, t.processSystemLine)
}

// tailFile tails a given file and processes new lines
func (t *Tracker) tailFile(path string, handler func(string)) {
    for {
        file, err := os.Open(path)
        if err != nil {
            time.Sleep(1 * time.Second)
            continue
        }
        defer file.Close()

        file.Seek(0, os.SEEK_END)
        reader := bufio.NewReader(file)

        for {
            line, err := reader.ReadString('\n')
            if err != nil {
                time.Sleep(0.1 * time.Second)
                continue
            }
            line = strings.TrimSpace(line)
            handler(line)
        }
    }
}

// processSystemLine handles lines from System.log
func (t *Tracker) processSystemLine(line string) {
    if strings.Contains(line, "You gained") && strings.Contains(line, "XP") {
        xp := extractFirstNumber(line)
        t.state.UpdateXP(xp)
    }
    if strings.Contains(line, "You received") && strings.Contains(line, "credits") {
        credits := extractFirstNumber(line)
        t.state.UpdateCredits(credits)
    }
    if strings.Contains(line, "Entering") {
        parts := strings.Split(line, "Entering")
        if len(parts) > 1 {
            zone := strings.TrimSpace(parts[1])
            t.state.UpdateZone(zone)
        }
    }
}

// processChatLine handles lines from Chat.log
func (t *Tracker) processChatLine(line string) {
    t.state.AddChatMessage(line)

    if strings.Contains(line, "You looted") {
        parts := strings.Split(line, "You looted")
        if len(parts) > 1 {
            item := strings.TrimSpace(parts[1])
            t.state.AddLoot(item)
        }
    }

    // Placeholder for crit detection (broad match for now)
    if strings.Contains(line, "hit") || strings.Contains(line, "Hit") {
        damage := extractFirstNumber(line)
        if damage > 0 {
            t.state.UpdateCrit(damage)
        }
    }
}

// extractFirstNumber pulls the first number out of a string
func extractFirstNumber(text string) int {
    re := regexp.MustCompile(`\\d+`)
    match := re.FindString(text)
    if match != "" {
        num, _ := strconv.Atoi(match)
        return num
    }
    return 0
}
