package main

//
// Model health measurements
//

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/procyon-projects/chrono"
)

var (
	ErrUnknownEntity = errors.New("unknown entity")
)

// Health schedules service checks, and exposes a method to query for the
// health status of a particular endpoint.
type HealthService struct {
	// Name contains the provider/entity this service is associated with.
	Name string
	// Checker is the service checker implementation associated with this service.
	Checker Checker

	// The idea behind storing current/previous measurement batches is that
	// we can detect (and store) deltas.
	Current  *MeasurementBatch
	Previous *MeasurementBatch
}

// Healthy returns true if the passed address/proto is known to be healthy. It will
// return an error if we don't have that measurement.
func (hs *HealthService) Healthy(addr net.Addr, proto string) (bool, error) {
	if hs.Current == nil {
		return false, ErrNotReady
	}
	healthy, err := hs.Current.Healthy(addr, proto)
	if err != nil {
		return healthy, fmt.Errorf("%w: %v", ErrUnknownEntity, err)
	}
	return healthy, nil
}

// Start initiates the scheduler
func (hs *HealthService) Start() error {
	taskScheduler := chrono.NewDefaultTaskScheduler()
	_, err := taskScheduler.ScheduleWithFixedDelay(hs.RunBatch, runInterval)
	return err
}

func (hs *HealthService) RunBatch(ctx context.Context) {
	log.Printf("[+] %s: launching health-check round", hs.Name)
	currentBatch := NewMeasurementBatch()
	ch := make(chan *Measurement)
	start := time.Now()
	go hs.Checker.Run(ch)
	for m := range ch {
		log.Printf("Measured %s: healthy=%v\n", m.String(), m.Healthy)
		currentBatch.healthMap[m.String()] = m
	}
	// all done, let's replace info from previous run
	if hs.Previous != nil {
		hs.Previous = hs.Current
	}
	hs.Current = currentBatch
	currentBatch.TimeInRound = time.Now().Sub(start)
	log.Printf("Finished measurement round in %v", currentBatch.TimeInRound)

	// TODO: loop to store changed status, this is something we'd like to persist somewhere
	// for historical records (continuity etc).
}

// A MeasuremenMeasurementBatch is a round of measurements.
type MeasurementBatch struct {
	// keys on healthMap should be Measurement.String()
	healthMap   map[string]*Measurement
	TimeInRound time.Duration
}

// NewNewMeasurementBatch returns a pointer to a MeasurementBatch ready to be
// used.
func NewMeasurementBatch() *MeasurementBatch {
	hm := make(map[string]*Measurement)
	mb := &MeasurementBatch{healthMap: hm}
	return mb
}

// Healthy returns true if the passed address is known to be healthy. It will
// return an error if we don't have that measurement.
func (mb *MeasurementBatch) Healthy(addr net.Addr, transport string) (bool, error) {
	key := fmt.Sprintf("%s/%s", addr.String(), transport)
	m, ok := mb.healthMap[key]
	if !ok {
		return false, fmt.Errorf("no such addr/transport")
	}
	return m.Healthy, nil
}

// Measurement is a single measurement. For what we really care about, it can
// only be healthy or unhealthy.
// For the time being, our criteria for "healthy" is having a gateway that can
// route us to the internet has a ICMP packet loss less than a certain threshold.
// There's more info we could share: did we get a handshake but we could not route?
// Can we resolve DNS through the GW? (gateway could have routing problems etc).
type Measurement struct {
	// Healthy is true if the service is up.
	Healthy bool
	// Addr is the address we measured.
	Addr net.Addr
	// Transport is tcp | udp
	Transport string
	// For metrics purposes.
	Timestamp time.Time
	// Recovered is intended to be set by the monitoring service, by
	// comparing with the previous run.
	// TODO remove if I don't end up using it (for instance, to make
	// any service using this one aware that this is a resource tha can be
	// added back into the pool).
	Recovered bool
}

func (m *Measurement) String() string {
	return fmt.Sprintf("%s/%s", m.Addr.String(), m.Transport)
}
