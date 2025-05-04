package observer

import (
	"context"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"axom-observer/pkg/config"
	"axom-observer/pkg/models"
	"axom-observer/pkg/protocols"
)

// TrafficSniffer captures and processes network and system usage signals.
type TrafficSniffer struct {
	rules    *config.Rules
	signalCh chan<- models.Signal
}

func NewTrafficSniffer(rules *config.Rules, signalCh chan<- models.Signal) *TrafficSniffer {
	return &TrafficSniffer{
		rules:    rules,
		signalCh: signalCh,
	}
}

// Start launches packet sniffing and system usage collection.
func (s *TrafficSniffer) Start(ctx context.Context) error {
	go s.sniffLoop(ctx)
	go s.systemUsageLoop(ctx)
	return nil
}

// sniffLoop captures packets and dispatches them for protocol parsing.
func (s *TrafficSniffer) sniffLoop(ctx context.Context) {
	handle, err := pcap.OpenLive("any", 65536, true, pcap.BlockForever)
	if err != nil {
		log.Printf("pcap open error: %v", err)
		return
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for {
		select {
		case <-ctx.Done():
			return
		case packet := <-packetSource.Packets():
			s.processPacket(packet)
		}
	}
}

// systemUsageLoop periodically collects system metrics (CPU, memory, etc.).
func (s *TrafficSniffer) systemUsageLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cpuPercent, _ := cpu.Percent(0, false)
			memStat, _ := mem.VirtualMemory()
			gpuUsage := getGPUUsage()
			sig := models.Signal{
				Timestamp: time.Now(),
				Protocol:  "system",
				CPUUsage:  0,
				MemUsage:  0,
				GPUUsage:  gpuUsage,
			}
			if len(cpuPercent) > 0 {
				sig.CPUUsage = cpuPercent[0]
			}
			if memStat != nil {
				sig.MemUsage = memStat.UsedPercent
			}
			s.signalCh <- sig
		}
	}
}

// getGPUUsage returns the GPU usage percent (stub, works with NVIDIA GPUs and nvidia-smi).
func getGPUUsage() float64 {
	out, err := exec.Command("nvidia-smi", "--query-gpu=utilization.gpu", "--format=csv,noheader,nounits").Output()
	if err != nil {
		return 0
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) > 0 {
		val, err := strconv.ParseFloat(strings.TrimSpace(lines[0]), 64)
		if err == nil {
			return val
		}
	}
	return 0
}

// processPacket detects protocol and dispatches to the correct parser.
// Extend this function to support more DBs/protocols.
func (s *TrafficSniffer) processPacket(packet gopacket.Packet) {
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if tcpLayer == nil || ipLayer == nil {
		return
	}
	tcp, _ := tcpLayer.(*layers.TCP)
	ip, _ := ipLayer.(*layers.IPv4)

	payload := tcp.Payload
	if len(payload) == 0 {
		return
	}

	src := &net.TCPAddr{IP: ip.SrcIP, Port: int(tcp.SrcPort)}
	dst := &net.TCPAddr{IP: ip.DstIP, Port: int(tcp.DstPort)}

	// Protocol detection by port (extend as needed)
	var proto string
	switch int(tcp.DstPort) {
	case 80, 443, 5000, 8000:
		proto = "http"
	case 50051:
		proto = "grpc"
	case 5432:
		proto = "postgres"
	case 3306:
		proto = "mysql"
	// Add more DBs/protocols here, e.g.:
	// case 27017:
	//     proto = "mongodb"
	default:
		return
	}

	var sig *models.Signal
	var err error
	switch proto {
	case "http":
		// Detect TLS/SSL handshake and emit a TLS signal instead of parsing as HTTP
		if len(payload) > 2 && payload[0] == 0x16 && payload[1] == 0x03 {
			// TLS handshake detected
			sig = &models.Signal{
				Timestamp:   time.Now(),
				Protocol:    "tls",
				Source:      protocols.AddrToEndpoint(src),
				Destination: protocols.AddrToEndpoint(dst),
				Operation:   "handshake",
				RawRequest:  payload,
				Metadata: map[string]interface{}{
					"note": "TLS handshake detected",
				},
			}
			s.signalCh <- *sig
			return
		}
		sig, err = protocols.ProcessHTTP(payload, src, dst)
	case "grpc":
		sig, err = protocols.ProcessGRPC(payload, src, dst)
	case "postgres":
		sig, err = protocols.ProcessPostgres(payload, src, dst)
	case "mysql":
		sig, err = protocols.ProcessMySQL(payload, src, dst)
		// Add more DBs/protocols here, e.g.:
		// case "mongodb":
		//     sig, err = protocols.ProcessMongoDB(payload, src, dst)
	}
	if err == nil && sig != nil {
		s.signalCh <- *sig
	} else if err != nil {
		// Suppress noisy TLS/SSL parse errors
		if !strings.Contains(err.Error(), "malformed HTTP request") && !strings.Contains(err.Error(), "invalid method") {
			log.Printf("Packet parse error: %v", err)
		}
	}
}
