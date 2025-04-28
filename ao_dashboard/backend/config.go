package main

import (
    "encoding/json"
    "os"
    "time"
)

// Config holds configurable settings for the tracker
type Config struct {
    ChatLogPath   string        `json:"chat_log_path"`
    SystemLogPath string        `json:"system_log_path"`
    StateFilePath string        `json:"state_file_path"`
    PingAddress   string        `json:"ping_address"`
    PingInterval  time.Duration `json:"ping_interval_ms"` // Ping every X milliseconds
}

// DefaultConfig returns a default configuration if none is provided
func DefaultConfig() Config {
    return Config{
        ChatLogPath:   "./Chat.log",            // default: working directory (replace later)
        SystemLogPath: "./System.log",          // default: working directory (replace later)
        StateFilePath: "../shared/state.json",  // shared folder relative to backend/
        PingAddress:   "8.8.8.8",                // Google's DNS (you can change to AO server later)
        PingInterval:  5000 * time.Millisecond,  // every 5 seconds
    }
}

// LoadConfig loads config.json if available, else uses default
func LoadConfig(path string) (Config, error) {
    file, err := os.Open(path)
    if err != nil {
        // Config file not found: return default
        return DefaultConfig(), nil
    }
    defer file.Close()

    decoder := json.NewDecoder(file)
    var cfg Config
    err = decoder.Decode(&cfg)
    if err != nil {
        return DefaultConfig(), nil
    }
    return cfg, nil
}
