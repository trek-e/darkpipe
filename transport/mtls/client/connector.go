// Package client provides an mTLS client for the home-device side of
// DarkPipe's alternative transport. The client maintains a persistent
// connection to the cloud relay with exponential backoff reconnection
// (1 s initial, 5 min max, never gives up).
//
// The only external dependency is github.com/cenkalti/backoff/v4.
package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// Client is an mTLS client that connects to a cloud relay server and
// maintains the connection using exponential backoff.
type Client struct {
	serverAddr string
	tlsConfig  *tls.Config
}

// NewClient creates a new mTLS client. It loads the CA certificate used to
// verify the server and the client's own certificate/key pair.
//
// serverAddr:     host:port of the mTLS server
// caCertPath:     PEM file containing the CA certificate (used to verify server cert)
// clientCertPath: PEM file containing the client's certificate
// clientKeyPath:  PEM file containing the client's private key
func NewClient(serverAddr, caCertPath, clientCertPath, clientKeyPath string) (*Client, error) {
	// Load CA certificate for server verification.
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("load CA cert %s: %w", caCertPath, err)
	}

	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate from %s", caCertPath)
	}

	// Load client certificate and key.
	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load client cert/key: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      rootCAs,
		MinVersion:   tls.VersionTLS12,
		// Cipher suites left as default -- Go negotiates TLS 1.3 automatically.
	}

	return &Client{
		serverAddr: serverAddr,
		tlsConfig:  tlsConfig,
	}, nil
}

// Connect performs a single TLS dial to the server.
// The caller is responsible for closing the returned connection.
func (c *Client) Connect(ctx context.Context) (*tls.Conn, error) {
	dialer := &tls.Dialer{Config: c.tlsConfig}
	conn, err := dialer.DialContext(ctx, "tcp", c.serverAddr)
	if err != nil {
		return nil, fmt.Errorf("tls dial %s: %w", c.serverAddr, err)
	}

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		conn.Close()
		return nil, fmt.Errorf("unexpected connection type from tls.Dialer")
	}

	return tlsConn, nil
}

// MaintainConnection keeps a persistent mTLS connection alive. When the
// connection drops or handler returns an error, it reconnects with exponential
// backoff (1 s initial interval, 5 min maximum interval, never gives up).
//
// The handler function receives the established connection and should run
// until the connection fails or ctx is cancelled. MaintainConnection only
// returns when ctx is cancelled.
func (c *Client) MaintainConnection(ctx context.Context, handler func(net.Conn) error) error {
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 1 * time.Second
	bo.MaxInterval = 5 * time.Minute
	bo.MaxElapsedTime = 0 // Never give up.

	operation := func() error {
		conn, err := c.Connect(ctx)
		if err != nil {
			return fmt.Errorf("connect: %w", err)
		}
		defer conn.Close()

		// Connection succeeded -- reset backoff so the next failure starts
		// from the initial interval again.
		bo.Reset()

		if err := handler(conn); err != nil {
			return fmt.Errorf("handler: %w", err)
		}

		return nil
	}

	bCtx := backoff.WithContext(bo, ctx)
	err := backoff.Retry(operation, bCtx)
	if err != nil && ctx.Err() != nil {
		// Context cancelled -- expected shutdown path.
		return ctx.Err()
	}
	return err
}
