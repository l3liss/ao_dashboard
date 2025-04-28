package main

import (
    "net"
    "time"
)

// Pinger handles measuring latency to a server
type Pinger struct {
    config Config
    state  *TrackerState
}

// NewPinger creates a new Pinger
func NewPinger(cfg Config, state *TrackerState) *Pinger {
    return &Pinger{
        config: cfg,
        state:  state,
    }
}

// Start begins periodic pinging
func (p *Pinger) Start() {
    go func() {
        for {
            latency := p.pingOnce(p.config.PingAddress)
            if latency >= 0 {
                p.state.UpdateLatency(latency)
            }
            time.Sleep(p.config.PingInterval)
        }
    }()
}

// pingOnce pings the server once using TCP and measures round-trip time
func (p *Pinger) pingOnce(address string) int {
    start := time.Now()

    conn, err := net.DialTimeout("tcp", net.JoinHostPort(address, "80"), 1*time.Second)
    if err != nil {
        return -1
    }
    conn.Close()

    elapsed := time.Since(start)
    return int(elapsed.Milliseconds())
}
