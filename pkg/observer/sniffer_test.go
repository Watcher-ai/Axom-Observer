package observer

import (
	"axom-observer/pkg/models"
	"testing"
)

func TestSystemUsageSignal(t *testing.T) {
	ch := make(chan models.Signal, 1)
	sniffer := &TrafficSniffer{signalCh: ch}
	// Simulate sending a system usage signal
	sig := models.Signal{
		Protocol: "system",
		CPUUsage: 42.0,
		MemUsage: 55.0,
		GPUUsage: 10.0,
	}
	sniffer.signalCh <- sig
	got := <-ch
	if got.Protocol != "system" || got.CPUUsage != 42.0 || got.MemUsage != 55.0 || got.GPUUsage != 10.0 {
		t.Errorf("unexpected system usage signal: %+v", got)
	}
}
