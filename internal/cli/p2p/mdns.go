package p2p

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/mdns"
)

const (
	serviceType = "_envsync._tcp"
	domain      = "local."
	txtKey      = "pubkey="
)

// BroadcastSession holds the server and a channel to notify when an ACK is received.
type BroadcastSession struct {
	Server  *mdns.Server
	AckChan <-chan struct{}
}

// BroadcastKey announces the public key and listens for an ACK ping.
func BroadcastKey(code string, pubKey string) (*BroadcastSession, error) {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "envsync-node"
	}
	if !strings.HasSuffix(host, ".") {
		host += "."
	}

	// Start a temporary TCP listener on a random port (Port 0)
	listener, err := net.Listen("tcp4", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to start ACK listener: %w", err)
	}

	// Extract the actual port assigned by the OS
	port := listener.Addr().(*net.TCPAddr).Port
	ackChan := make(chan struct{}, 1)

	// Wait for exactly ONE connection (The ACK), then close everything
	go func() {
		defer listener.Close()
		conn, err := listener.Accept()
		if err == nil {
			conn.Close() // We don't need to read data, the connection ITSELF is the signal
			ackChan <- struct{}{}
		}
	}()

	txtRecord := []string{fmt.Sprintf("%s%s", txtKey, pubKey)}

	// Register the mDNS service USING the dynamic TCP port
	service, err := mdns.NewMDNSService(code, serviceType, domain, host, port, nil, txtRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to create mDNS service: %w", err)
	}

	server, err := mdns.NewServer(&mdns.Config{
		Zone: service,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start mDNS server: %w", err)
	}

	return &BroadcastSession{
		Server:  server,
		AckChan: ackChan,
	}, nil
}

// DiscoverResult groups the retrieved key and the trigger to send the ACK.
type DiscoverResult struct {
	PubKey  string
	SendAck func()
}

// DiscoverKey scans the local network and returns the key and an ACK trigger.
func DiscoverKey(ctx context.Context, code string) (*DiscoverResult, error) {
	entries := make(chan *mdns.ServiceEntry, 100)
	resultChan := make(chan *DiscoverResult, 1)

	go func() {
		for entry := range entries {
			if strings.Contains(entry.Name, code) {
				for _, txt := range entry.InfoFields {
					if after, ok := strings.CutPrefix(txt, txtKey); ok {
						pubKey := after

						// Capture connection info for the ACK
						var targetIP net.IP
						if entry.AddrV4 != nil {
							targetIP = entry.AddrV4
						}
						targetPort := entry.Port

						// Create the ACK closure
						ackFunc := func() {
							if targetIP != nil && targetPort > 0 {
								addr := net.JoinHostPort(targetIP.String(), fmt.Sprintf("%d", targetPort))
								// Brief timeout: if it fails, it fails silently, no big deal.
								conn, _ := net.DialTimeout("tcp4", addr, 1*time.Second)
								if conn != nil {
									conn.Close()
								}
							}
						}

						select {
						case resultChan <- &DiscoverResult{PubKey: pubKey, SendAck: ackFunc}:
						default:
						}
						return
					}
				}
			}
		}
	}()

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		params := &mdns.QueryParam{
			Service:     serviceType,
			Domain:      "local",
			Timeout:     10 * time.Second,
			Entries:     entries,
			DisableIPv6: true,
			Logger:      newLogger(),
		}
		err := mdns.Query(params)
		if err != nil {
			errCh <- fmt.Errorf("mDNS query failed: %w", err)
		}
		close(entries)
	}()

	select {
	case res := <-resultChan:
		return res, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("code not found on local network (timeout or canceled)")
	}
}
