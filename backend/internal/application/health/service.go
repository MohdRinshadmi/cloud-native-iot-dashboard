// Package health is an application service that reports liveness and
// readiness. Liveness = "the process is up". Readiness = "every critical
// dependency is reachable". Dependencies register themselves as Checkers,
// so the service has zero knowledge of Postgres/Redis/MQTT specifics.
package health

import (
	"context"
	"sync"
	"time"
)

// Status is the rollup health state.
type Status string

const (
	StatusUp   Status = "up"
	StatusDown Status = "down"
)

// Checker is a port implemented by each critical dependency.
type Checker interface {
	// Name identifies the dependency in the report (e.g. "postgres").
	Name() string
	// Check returns nil when healthy, or an error describing the failure.
	Check(ctx context.Context) error
}

// ComponentReport is the per-dependency result.
type ComponentReport struct {
	Status Status `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Report is the aggregate readiness result.
type Report struct {
	Status     Status                     `json:"status"`
	Version    string                     `json:"version"`
	UptimeSec  float64                    `json:"uptime_seconds"`
	Components map[string]ComponentReport `json:"components"`
}

// Service runs all registered checkers concurrently and rolls up the result.
type Service struct {
	version   string
	startedAt time.Time
	checkers  []Checker
}

// New builds a health service. Checkers are appended via Register at wire time.
func New(version string, startedAt time.Time) *Service {
	return &Service{version: version, startedAt: startedAt}
}

// Register adds a dependency checker. Not safe for concurrent registration;
// call only during startup wiring.
func (s *Service) Register(c Checker) { s.checkers = append(s.checkers, c) }

// Live always returns up — it answers "is the process running?".
func (s *Service) Live() Status { return StatusUp }

// Ready runs every checker concurrently (fan-out/fan-in) and aggregates.
func (s *Service) Ready(ctx context.Context) Report {
	report := Report{
		Status:     StatusUp,
		Version:    s.version,
		UptimeSec:  time.Since(s.startedAt).Seconds(),
		Components: make(map[string]ComponentReport, len(s.checkers)),
	}

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)
	for _, c := range s.checkers {
		wg.Add(1)
		go func(c Checker) {
			defer wg.Done()
			cr := ComponentReport{Status: StatusUp}
			if err := c.Check(ctx); err != nil {
				cr.Status = StatusDown
				cr.Error = err.Error()
			}
			mu.Lock()
			report.Components[c.Name()] = cr
			if cr.Status == StatusDown {
				report.Status = StatusDown
			}
			mu.Unlock()
		}(c)
	}
	wg.Wait()

	return report
}
