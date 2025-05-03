package main

import (
    "encoding/json"
    "os"
    "time"
)

// Config holds configurable settings for the tracker
// Includes paths for chat, system, and loot logs.
type Config struct {
    ChatLogPath   string        `json:"chat_log_path"`
    SystemLogPath string        `json:"system_log_path"`
    LootLogPath   string        `json:"loot_log_path"`
    StateFilePath string        `json:"state_file_path"`
    PingAddress   string        `json:"ping_address"`
    PingInterval  time.Duration `json:"ping_interval_ms"`
}

// DefaultConfig returns defaults for all paths and settings
func DefaultConfig() Config {
    return Config{
        ChatLogPath:   "../ao_logs/chat.log",
        SystemLogPath: "../ao_logs/combat.log",
        LootLogPath:   "../ao_logs/loot.log",
        StateFilePath: "../shared/state.json",
        PingAddress:   "8.8.8.8",
        PingInterval:  5000 * time.Millisecond,
    }
}

// LoadConfig loads config.json if present, else uses defaults
func LoadConfig(path string) (Config, error) {
    file, err := os.Open(path)
    if err != nil {
        return DefaultConfig(), nil
    }
    defer file.Close()

    var cfg Config
    dec := json.NewDecoder(file)
    if err := dec.Decode(&cfg); err != nil {
        return DefaultConfig(), nil
    }
    return cfg, nil
}

