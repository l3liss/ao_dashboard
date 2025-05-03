package main

import (
    "net"
    "os/exec"
    "strconv"
    "strings"
    "time"
)

// Pinger periodically measures ICMP latency to the configured host and updates the state
type Pinger struct {
    cfg   Config
    state *TrackerState
}

// NewPinger constructs a Pinger with the given config and state
func NewPinger(cfg Config, state *TrackerState) *Pinger {
    return &Pinger{cfg: cfg, state: state}
}

// Start begins the ping loop in a separate goroutine
func (p *Pinger) Start() {
    go func() {
        // Extract host (ignore port if present)
        host, _, err := net.SplitHostPort(p.cfg.PingAddress)
        if err != nil {
            host = p.cfg.PingAddress
        }
        for {
            // Use system ping for ICMP
            cmd := exec.Command("ping", "-c", "1", "-W", "1", host)
            out, err := cmd.CombinedOutput()
            var latencyMs int
            if err != nil {
                // fmt.Println("[PING] Error pinging", host, err)
                latencyMs = -1
            } else {
                // parse e.g. "time=23.4 ms"
                s := string(out)
                if idx := strings.Index(s, "time="); idx != -1 {
                    part := s[idx+5:]
                    fields := strings.Fields(part)
                    if len(fields) > 0 {
                        msStr := strings.TrimSuffix(fields[0], "ms")
                        msStr = strings.TrimSpace(strings.TrimSuffix(msStr, "ms"))
                        if f, err := strconv.ParseFloat(msStr, 64); err == nil {
                            latencyMs = int(f)
                        }
                    }
                }
                // fallback if parsing failed
                if latencyMs == 0 {
                    latencyMs = int(time.Since(time.Now()).Milliseconds())
                }
            }
            p.state.UpdateLatency(latencyMs)
            time.Sleep(p.cfg.PingInterval)
        }
    }()
}

