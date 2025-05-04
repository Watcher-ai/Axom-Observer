package protocols

import (
	"net"
	"testing"
)

func TestProcessHTTP(t *testing.T) {
	raw := []byte("GET /foo HTTP/1.1\r\nHost: example.com\r\n\r\n")
	src := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
	dst := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 80}
	sig, err := ProcessHTTP(raw, src, dst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sig.Protocol != "http" {
		t.Errorf("expected protocol http, got %s", sig.Protocol)
	}
	if sig.Operation != "GET /foo" {
		t.Errorf("expected operation GET /foo, got %s", sig.Operation)
	}
}
