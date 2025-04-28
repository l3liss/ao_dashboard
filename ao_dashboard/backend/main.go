package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    // Load configuration
    config, err := LoadConfig("config.json")
    if err != nil {
        fmt.Println("Error loading config:", err)
        os.Exit(1)
    }
    fmt.Println("Config loaded.")

    // Create tracker state
    trackerState := NewTrackerState()

    // Start tracker (logs)
    tracker := NewTracker(config, trackerState)
    tracker.Start()
    fmt.Println("Log tracker started.")

    // Start pinger (latency checker)
    pinger := NewPinger(config, trackerState)
    pinger.Start()
    fmt.Println("Pinger started.")

    // Start periodic state saving
    go func() {
        for {
            err := trackerState.SaveToFile(config.StateFilePath)
            if err != nil {
                fmt.Println("Error saving state:", err)
            }
            time.Sleep(1 * time.Second) // Save every 1 second
        }
    }()
    fmt.Println("State autosave started.")

    // Wait for CTRL+C to exit
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c

    fmt.Println("\nShutting down cleanly.")
}
