package protocols

import (
	"net"
	"time"

	"axom-observer/pkg/models"
)

// ProcessGRPC parses gRPC requests and responses from raw packets.
// For production, use proto descriptors and TCP stream reassembly.
func ProcessGRPC(packet []byte, src, dst net.Addr) (*models.Signal, error) {
	// TODO: Implement gRPC request/response parsing and outcome extraction.
	return &models.Signal{
		Timestamp:   time.Now(),
		Protocol:    "grpc",
		Source:      AddrToEndpoint(src),
		Destination: AddrToEndpoint(dst),
		Operation:   "grpc_call",
		Metadata:    map[string]interface{}{},
		RawRequest:  packet,
	}, nil
}
