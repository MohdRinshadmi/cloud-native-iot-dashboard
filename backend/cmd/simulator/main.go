// Command simulator emulates an IoT fleet for local development:
//
//  1. signs into the API (demo admin) and fetches the registered devices
//  2. opens ONE MQTT CONNECTION PER DEVICE (like real hardware), each with a
//     Last-Will "offline" announcement the broker fires on ungraceful death
//  3. publishes an "online" status, then telemetry on an interval with
//     random-walk metrics (temperature, battery, voltage, cpu, memory,
//     signal) and a GPS drift around a home coordinate
//
// Usage:
//
//	go run ./cmd/simulator                       # defaults: all devices, 3s
//	go run ./cmd/simulator -interval 1s -devices 4
//	go run ./cmd/simulator -broker tcp://localhost:1883 -api http://localhost:8080
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type cliFlags struct {
	api      string
	broker   string
	email    string
	password string
	interval time.Duration
	devices  int
}

func main() {
	var f cliFlags
	flag.StringVar(&f.api, "api", "http://localhost:8080", "API base URL")
	flag.StringVar(&f.broker, "broker", "tcp://localhost:1883", "MQTT broker URL")
	flag.StringVar(&f.email, "email", "admin@demo.local", "API login email")
	flag.StringVar(&f.password, "password", "Password123!", "API login password")
	flag.DurationVar(&f.interval, "interval", 3*time.Second, "telemetry publish interval")
	flag.IntVar(&f.devices, "devices", 0, "number of devices to simulate (0 = all)")
	flag.Parse()

	devices, err := fetchDevices(f)
	if err != nil {
		log.Fatalf("fetch devices: %v", err)
	}
	if f.devices > 0 && f.devices < len(devices) {
		devices = devices[:f.devices]
	}
	if len(devices) == 0 {
		log.Fatal("no devices registered — log into the dashboard and add some first")
	}

	log.Printf("simulating %d devices → %s (every %s)", len(devices), f.broker, f.interval)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	for i, d := range devices {
		wg.Add(1)
		go func(idx int, dev apiDevice) {
			defer wg.Done()
			runDevice(ctx, f, idx, dev)
		}(i, d)
	}
	<-ctx.Done()
	log.Println("shutting down — devices announce offline")
	wg.Wait()
}

// ---- device loop ---------------------------------------------------------

func runDevice(ctx context.Context, f cliFlags, idx int, dev apiDevice) {
	statusTopic := fmt.Sprintf("devices/%s/status", dev.ID)
	telemetryTopic := fmt.Sprintf("devices/%s/telemetry", dev.ID)

	opts := paho.NewClientOptions().
		AddBroker(f.broker).
		SetClientID("sim-"+dev.ID[:8]).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		// Real devices die ungracefully; the broker announces it for them.
		SetWill(statusTopic, `{"status":"offline"}`, 1, false)

	client := paho.NewClient(opts)
	if t := client.Connect(); t.Wait() && t.Error() != nil {
		log.Printf("[%s] connect failed: %v", dev.Name, t.Error())
		return
	}

	client.Publish(statusTopic, 1, false, `{"status":"online"}`)
	log.Printf("[%s] online", dev.Name)

	// Per-device deterministic baselines so the fleet looks heterogeneous.
	rng := rand.New(rand.NewSource(int64(idx)*7919 + 17))
	state := newMetricState(rng, idx)

	ticker := time.NewTicker(f.interval + time.Duration(rng.Intn(500))*time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t := client.Publish(statusTopic, 1, false, `{"status":"offline"}`)
			t.WaitTimeout(time.Second)
			client.Disconnect(250)
			return
		case <-ticker.C:
			payload, _ := json.Marshal(state.next(rng))
			client.Publish(telemetryTopic, 1, false, payload)
		}
	}
}

// ---- metric synthesis -----------------------------------------------------

type metricState struct {
	temp, battery, voltage, cpu, memory, signal float64
	lat, lng                                    float64
}

// newMetricState seeds plausible industrial baselines; devices spread around
// Dubai as the demo fleet's home region.
func newMetricState(rng *rand.Rand, idx int) *metricState {
	return &metricState{
		temp:    45 + rng.Float64()*15,
		battery: 60 + rng.Float64()*40,
		voltage: 11.5 + rng.Float64(),
		cpu:     20 + rng.Float64()*30,
		memory:  30 + rng.Float64()*30,
		signal:  -60 - rng.Float64()*25,
		lat:     25.10 + float64(idx%4)*0.045 + rng.Float64()*0.02,
		lng:     55.15 + float64(idx/4)*0.06 + rng.Float64()*0.02,
	}
}

type telemetryPayload struct {
	TS          time.Time `json:"ts"`
	Temperature float64   `json:"temperature"`
	Battery     float64   `json:"battery"`
	Voltage     float64   `json:"voltage"`
	CPU         float64   `json:"cpu"`
	Memory      float64   `json:"memory"`
	Signal      float64   `json:"signal"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
}

// next advances every metric one random-walk step within realistic bounds.
func (s *metricState) next(rng *rand.Rand) telemetryPayload {
	walk := func(v *float64, step, lo, hi float64) {
		*v += (rng.Float64() - 0.5) * 2 * step
		*v = math.Max(lo, math.Min(hi, *v))
	}
	walk(&s.temp, 1.2, 20, 95)
	walk(&s.battery, 0.4, 5, 100)
	s.battery = math.Max(5, s.battery-0.05) // slow discharge
	walk(&s.voltage, 0.05, 10.5, 13.5)
	walk(&s.cpu, 4, 2, 99)
	walk(&s.memory, 2, 10, 95)
	walk(&s.signal, 2, -110, -50)
	walk(&s.lat, 0.0006, 24.9, 25.5)
	walk(&s.lng, 0.0006, 55.0, 55.6)

	return telemetryPayload{
		TS:          time.Now().UTC(),
		Temperature: round(s.temp, 10),
		Battery:     round(s.battery, 10),
		Voltage:     round(s.voltage, 100),
		CPU:         round(s.cpu, 10),
		Memory:      round(s.memory, 10),
		Signal:      round(s.signal, 10),
		Lat:         round(s.lat, 1e5),
		Lng:         round(s.lng, 1e5),
	}
}

func round(v, scale float64) float64 { return math.Round(v*scale) / scale }

// ---- API access -------------------------------------------------------------

type apiDevice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func fetchDevices(f cliFlags) ([]apiDevice, error) {
	loginBody, _ := json.Marshal(map[string]string{"email": f.email, "password": f.password})
	res, err := http.Post(f.api+"/api/v1/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login returned %d", res.StatusCode)
	}
	var session struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&session); err != nil {
		return nil, err
	}

	req, _ := http.NewRequest(http.MethodGet, f.api+"/api/v1/devices?limit=200", nil)
	req.Header.Set("Authorization", "Bearer "+session.AccessToken)
	res2, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res2.Body.Close()
	var list struct {
		Data []apiDevice `json:"data"`
	}
	if err := json.NewDecoder(res2.Body).Decode(&list); err != nil {
		return nil, err
	}
	return list.Data, nil
}
