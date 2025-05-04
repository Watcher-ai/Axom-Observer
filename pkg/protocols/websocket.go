package protocols

import (
	"net"
	"time"

	"axom-observer/pkg/models"
)

// ProcessWebSocket parses WebSocket messages from raw packets.
// For production, implement full WebSocket frame parsing and message correlation.
func ProcessWebSocket(packet []byte, src, dst net.Addr) (*models.Signal, error) {
	// TODO: Implement WebSocket frame parsing and outcome extraction.
	return &models.Signal{
		Timestamp:   time.Now(),
		Protocol:    "websocket",
		Source:      AddrToEndpoint(src),
		Destination: AddrToEndpoint(dst),
		Operation:   "ws_message",
		Metadata:    map[string]interface{}{},
		RawRequest:  packet,
	}, nil
}
