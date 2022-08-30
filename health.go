package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/procyon-projects/chrono"
)

//
// Model health measurements
//

var (
	runInterval = 60 * time.Second
)

type HealthService struct {
	Name     string
	Last     *MeasurementBatch
	Previous *MeasurementBatch
	Checker  Checker
}

// Healthy returns true if the passed address is known to be healthy. It will
// return an error if we don't have that measurement.
func (hs *HealthService) Healthy(addr net.Addr) (bool, error) {
	return false, nil
}

// Start initiates the scheduler
func (hs *HealthService) Start() error {
	taskScheduler := chrono.NewDefaultTaskScheduler()
	_, err := taskScheduler.ScheduleWithFixedDelay(hs.RunBatch, runInterval)
	return err
}

func (hs *HealthService) RunBatch(ctx context.Context) {
	log.Printf("[+] %s: running round...", hs.Name)
	ch := make(chan *Measurement)
	go hs.Checker.Run(ch)
	for m := range ch {
		log.Println()
		log.Println(">>>", m.Addr, "healthy:", m.Healthy)
		log.Println()
	}
}

// A MeasuremenMeasurementBatch is a round of measurements.
type MeasurementBatch struct {
	TimeInRound *time.Duration
	healthMap   map[net.Addr]*Measurement
}

// Healthy returns true if the passed address is known to be healthy. It will
// return an error if we don't have that measurement.
func (mb *MeasurementBatch) Healthy(addr net.Addr) (bool, error) {
	return false, nil
}

// Measurement is a single measurement. It can only be healthy or unhealthy.
type Measurement struct {
	// Healthy is true if the service is up.
	Healthy bool
	// Recovered is intended to be set by the monitoring service, by
	// comparing with the previous run.
	Recovered bool
	// Addr is the address we measured.
	Addr net.Addr
	// For metrics purposes.
	LastMeasured *time.Time
}
