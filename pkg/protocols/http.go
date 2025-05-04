package protocols

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"strconv"
	"time"

	"axom-observer/pkg/models"
)

// ProcessHTTP parses HTTP requests and (optionally) responses from raw packets.
func ProcessHTTP(packet []byte, src, dst net.Addr) (*models.Signal, error) {
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(packet)))
	if err != nil {
		return nil, err
	}

	signal := &models.Signal{
		Timestamp:   time.Now(),
		Protocol:    "http",
		Source:      AddrToEndpoint(src),
		Destination: AddrToEndpoint(dst),
		Operation:   req.Method + " " + req.URL.Path,
		Status:      0,
		LatencyMS:   0,
		Metadata: map[string]interface{}{
			"host":   req.Host,
			"path":   req.URL.Path,
			"method": req.Method,
		},
		RawRequest: packet,
	}

	// TODO: For production, implement TCP stream reassembly and response correlation.
	// Optionally parse response if available (not always possible in sniffed traffic)
	// Example stub:
	// resp, err := http.ReadResponse(...)
	// if err == nil {
	//     signal.Status = resp.StatusCode
	//     signal.LatencyMS = ... // calculate from timestamps
	//     signal.RawResponse = ... // raw response bytes
	// }

	return signal, nil
}

// AddrToEndpoint converts a net.Addr to models.Endpoint.
// Extend to support IPv6 and error handling as needed.
func AddrToEndpoint(addr net.Addr) models.Endpoint {
	if addr == nil {
		return models.Endpoint{}
	}
	host, portStr, err := net.SplitHostPort(addr.String())
	port := 0
	if err == nil {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}
	return models.Endpoint{IP: host, Port: port}
}
