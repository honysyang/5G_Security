package detector

import (
	"sync"
	"time"
)

// SCTPINITFloodDetector detects SCTP INIT Flood attacks
type SCTPINITFloodDetector struct {
	Window      time.Duration
	Threshold   int
	Events      map[MessageKey]*Event
	CleanupChan chan bool
	StopChan    chan bool
	sync.Mutex
}

// NewSCTPINITFloodDetector creates a new SCTPINITFloodDetector instance
func NewSCTPINITFloodDetector(window time.Duration, threshold int) *SCTPINITFloodDetector {
	return &SCTPINITFloodDetector{
		Window:      window,
		Threshold:   threshold,
		Events:      make(map[MessageKey]*Event),
		CleanupChan: make(chan bool),
		StopChan:    make(chan bool),
	}
}

// Start begins the detector's operation
func (d *SCTPINITFloodDetector) Start() {
	go func() {
		for {
			select {
			case <-d.CleanupChan:
				d.cleanup()
			case <-d.StopChan:
				return
			}
		}
	}()
}

// Stop halts the detector's operation
func (d *SCTPINITFloodDetector) Stop() {
	close(d.StopChan)
}

// AddEvent adds a new event to the detector
func (d *SCTPINITFloodDetector) AddEvent(messageType uint8, destIP string, destPort uint16) {
	key := MessageKey{MessageType: messageType, DestIP: destIP, DestPort: destPort}
	now := time.Now()

	d.Lock()
	defer d.Unlock()

	if event, exists := d.Events[key]; exists {
		// If the event exists, update the count and timestamp
		event.Count++
		event.Timestamp = now
	} else {
		// If the event does not exist, create a new one
		d.Events[key] = &Event{
			MessageType: messageType,
			IP:          destIP,
			Port:        destPort,
			Count:       1,
			Timestamp:   now,
		}
	}
}

// CheckINITFlood checks for a SCTP INIT Flood attack
func (d *SCTPINITFloodDetector) CheckINITFlood() (bool, MessageKey) {
	d.Lock()
	defer d.Unlock()

	for key, event := range d.Events {
		if event.Count > d.Threshold {
			return true, key
		}
	}
	return false, MessageKey{}
}

// cleanup removes expired events from the map
func (d *SCTPINITFloodDetector) cleanup() {
	d.Lock()
	defer d.Unlock()

	now := time.Now()
	for key, event := range d.Events {
		if now.Sub(event.Timestamp) > d.Window {
			// If the event is expired, remove it from the map
			delete(d.Events, key)
		}
	}
}
