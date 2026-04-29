package callback

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/lucaskatayama/oauth2-cli/internal/config"

	"context"
	"fmt"
	"net"
	"net/http"
)

// generateSelfSignedCert creates a self-signed certificate for the given host.
// Returns a tls.Certificate ready to use.
func generateSelfSignedCert(host string) (tls.Certificate, error) {
	// Generate ECDSA P-256 key (fast and suitable for dev)
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	// Prepare certificate template
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return tls.Certificate{}, err
	}
	org := "Dev"
	if host != "" {
		org = host
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{org},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour), // 1 year validity

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// If host looks like an IP, we could add IPAddresses, but for simplicity
	// we add the host as DNSName and also localhost for convenience.
	var dnsNames []string
	if host != "" {
		dnsNames = append(dnsNames, host)
	}
	// Always include localhost for loopback testing
	dnsNames = append(dnsNames, "localhost")
	template.DNSNames = dnsNames

	// Create the self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	// PEM encode certificate and key
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	// Parse PEM data to tls.Certificate
	return tls.X509KeyPair(certPEM, keyPEM)
}

type Result struct {
	Code  string
	State string
	Error error
}

type Server struct {
	host     string
	port     int
	scheme   string // "http" or "https"
	path     string
	result   chan Result
	listener net.Listener
	server   *http.Server
}

// NewServer creates a new Server instance and resolves configuration from flags/env/defaults.
func NewServer(_ int) *Server {
	// Resolve callback configuration (host/port) using env vars or flags.
	cfg, err := config.Resolve(nil)
	if err != nil {
		// Fallback to defaults on error
		cfg = &config.CallbackConfig{URL: "http://localhost:9999/callback", Host: "localhost", Port: "9999", Scheme: "http", Path: "/callback"}
	}
	portNum, _ := strconv.Atoi(cfg.Port)
	return &Server{
		host:   cfg.Host,
		port:   portNum,
		scheme: cfg.Scheme,
		path:   cfg.Path,
		result: make(chan Result, 1),
	}
}

func (s *Server) Start() (string, error) {
	// Resolve configuration on each start to capture latest flag values.
	cfg, err := config.Resolve(nil)
	if err != nil {
		return "", err
	}
	// Update server fields with latest config (in case flags changed)
	s.host = cfg.Host
	s.port, _ = strconv.Atoi(cfg.Port)
	s.scheme = cfg.Scheme
	s.path = cfg.Path
	if s.path == "" {
		s.path = "/callback"
	}

	bindHost := s.host
	if bindHost == "" {
		bindHost = "0.0.0.0"
	}
	if strings.EqualFold(bindHost, "localhost") {
		bindHost = "0.0.0.0"
	} else if ip := net.ParseIP(bindHost); ip != nil {
		if ip.IsLoopback() || ip.IsUnspecified() {
			bindHost = "0.0.0.0"
		}
	} else {
		if resolved, err := net.LookupIP(bindHost); err == nil {
			allLoopback := true
			for _, ip := range resolved {
				if !ip.IsLoopback() {
					allLoopback = false
					break
				}
			}
			if allLoopback {
				bindHost = "0.0.0.0"
			}
		}
	}
	addr := fmt.Sprintf("%s:%d", bindHost, s.port)
	var ln net.Listener

	if s.scheme == "https" {
		// Generate self-signed certificate for the host
		cert, err := generateSelfSignedCert(s.host)
		if err != nil {
			return "", fmt.Errorf("failed to generate self-signed certificate: %w", err)
		}
		tlsCfg := &tls.Config{
			Certificates: []tls.Certificate{cert},
			// Use modern TLS settings
			MinVersion: tls.VersionTLS12,
		}
		ln, err = tls.Listen("tcp", addr, tlsCfg)
		if err != nil {
			return "", fmt.Errorf("could not start HTTPS listener on %s: %w", addr, err)
		}
	} else {
		ln, err = net.Listen("tcp", addr)
		if err != nil {
			return "", fmt.Errorf("could not start HTTP listener on %s: %w", addr, err)
		}
	}
	s.listener = ln

	mux := http.NewServeMux()
	// Register handler for configured path (and allow optional trailing slash)
	mux.HandleFunc(s.path, s.handle)
	if trimmed := strings.TrimSuffix(s.path, "/"); trimmed != "" && trimmed != s.path {
		mux.HandleFunc(trimmed, s.handle)
	}
	s.server = &http.Server{Handler: mux}

	// Start server in a goroutine
	go s.server.Serve(ln)

	return cfg.URL, nil
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	code := query.Get("code")
	state := query.Get("state")
	errParam := query.Get("error")

	if errParam != "" {
		errDesc := query.Get("error_description")
		s.result <- Result{Error: fmt.Errorf("%s: %s", errParam, errDesc)}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Authorization failed. You can close this window."))
		return
	}

	if code == "" {
		s.result <- Result{Error: fmt.Errorf("no code in callback")}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No code received. You can close this window."))
		return
	}

	s.result <- Result{Code: code, State: state}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<html><body style='font-family: sans-serif; text-align: center; padding: 50px;'><h2>✅ Authorization complete!</h2><p>You can close this window and return to the CLI.</p></body></html>"))
}

func (s *Server) Wait(ctx context.Context) (string, string, error) {
	select {
	case result := <-s.result:
		if result.Error != nil {
			return "", "", result.Error
		}
		return result.Code, result.State, nil
	case <-ctx.Done():
		return "", "", ctx.Err()
	}
}

func (s *Server) Close() error {
	if s.server != nil {
		return s.server.Shutdown(context.Background())
	}
	return nil
}

func (s *Server) DeepLinkURL() string {
	return "o2x://callback"
}
