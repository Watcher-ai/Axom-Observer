package observer

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"sync"
	"time"
)

// MITMProxy handles HTTPS interception with a self-signed CA
// For local/dev use only. In production, use a trusted CA and secure key management.
type MITMProxy struct {
	Addr       string
	CAKeyPath  string
	CACertPath string
	logger     *log.Logger
	server     *http.Server
	mu         sync.Mutex
	certCache  map[string]*tls.Certificate
}

func NewMITMProxy(addr, caCertPath, caKeyPath string, logger *log.Logger) *MITMProxy {
	return &MITMProxy{
		Addr:       addr,
		CAKeyPath:  caKeyPath,
		CACertPath: caCertPath,
		logger:     logger,
		certCache:  make(map[string]*tls.Certificate),
	}
}

// Start launches the MITM HTTPS proxy
func (p *MITMProxy) Start(ctx context.Context, handler http.Handler) error {
	p.logger.Printf("[MITM] Starting HTTPS proxy on %s", p.Addr)

	// Ensure CA cert/key exist
	if err := ensureCA(p.CACertPath, p.CAKeyPath, p.logger); err != nil {
		return err
	}

	caCert, caKey, err := loadCA(p.CACertPath, p.CAKeyPath)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return p.getOrCreateCert(hello.ServerName, caCert, caKey)
		},
	}

	p.server = &http.Server{
		Addr:      p.Addr,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	go func() {
		if err := p.server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			p.logger.Printf("[MITM] Proxy error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.server.Shutdown(shutdownCtx)
}

// getOrCreateCert returns a leaf cert for the given server name
func (p *MITMProxy) getOrCreateCert(serverName string, caCert *x509.Certificate, caKey *rsa.PrivateKey) (*tls.Certificate, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if cert, ok := p.certCache[serverName]; ok {
		return cert, nil
	}
	cert, err := generateLeafCert(serverName, caCert, caKey)
	if err != nil {
		return nil, err
	}
	p.certCache[serverName] = cert
	return cert, nil
}

// ensureCA generates a CA cert/key if not present
func ensureCA(certPath, keyPath string, logger *log.Logger) error {
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		logger.Printf("[MITM] Generating new CA cert/key at %s, %s", certPath, keyPath)
		return generateCA(certPath, keyPath)
	}
	return nil
}

// generateCA creates a new self-signed CA cert/key
func generateCA(certPath, keyPath string) error {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: "Axom Observer MITM CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return err
	}
	certOut, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyOut, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return nil
}

// loadCA loads the CA cert/key from disk
func loadCA(certPath, keyPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, err
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, err
	}
	certBlock, _ := pem.Decode(certPEM)
	keyBlock, _ := pem.Decode(keyPEM)
	if certBlock == nil || keyBlock == nil {
		return nil, nil, io.ErrUnexpectedEOF
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}
	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}
	return cert, key, nil
}

// generateLeafCert creates a leaf cert for a given server name
func generateLeafCert(serverName string, caCert *x509.Certificate, caKey *rsa.PrivateKey) (*tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: serverName},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		DNSNames:     []string{serverName},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &priv.PublicKey, caKey)
	if err != nil {
		return nil, err
	}
	cert := &tls.Certificate{
		Certificate: [][]byte{certDER, caCert.Raw},
		PrivateKey:  priv,
	}
	return cert, nil
}
