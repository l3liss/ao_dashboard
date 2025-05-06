package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    // Load configuration
    config, err := LoadConfig("config.json")
    if err != nil {
        fmt.Println("Error loading config:", err)
        os.Exit(1)
    }
    fmt.Println("Config loaded.")

    // Create tracker state with specified state file path
    trackerState := NewTrackerState(config.StateFilePath)

    // Start periodic autosave to state file
    trackerState.StartAutoSave(config.StateFilePath)
    fmt.Println("State autosave started.")

    // Start tracker (logs)
    tracker := NewTracker(config, trackerState)
    tracker.Start()
    fmt.Println("Log tracker started.")

    // Start pinger (latency checker)
    pinger := NewPinger(config, trackerState)
    pinger.Start()
    fmt.Println("Pinger started.")

    // Block until SIGINT or SIGTERM
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    <-sigCh

    fmt.Println("\nShutting down cleanly.")
}

