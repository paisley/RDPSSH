package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// getSSHConfig creates SSH client configuration with cert-based auth
// WARNING: Uses InsecureIgnoreHostKey - should implement known_hosts verification in production
func getSSHConfig(user string, p12Info *P12Info) (*ssh.ClientConfig, error) {
	signer, err := ssh.NewSignerFromKey(p12Info.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer from private key: %w", err)
	}

	return &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}, nil
}

// TestConnection verifies SSH connectivity and authentication
func TestConnection(host, user string, p12Info *P12Info) (string, error) {
	config, err := getSSHConfig(user, p12Info)
	if err != nil {
		return "", err
	}

	addr := host
	if _, _, err := net.SplitHostPort(host); err != nil {
		addr = host + ":22"
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", fmt.Errorf("connection failed: %v", err)
	}
	defer client.Close()

	return fmt.Sprintf("Successfully authenticated to %s as %s", host, user), nil
}

// StartTunnel establishes SSH tunnel, forwards localhost:localPort to remote RDP (3389),
// and launches mstsc.exe. Blocks until RDP client exits or context is cancelled.
func StartTunnel(ctx context.Context, host, user, localPort string, p12Info *P12Info, logFunc func(string), onReady func()) error {
	config, err := getSSHConfig(user, p12Info)
	if err != nil {
		return err
	}

	addr := host
	if _, _, err := net.SplitHostPort(host); err != nil {
		addr = host + ":22"
	}

	logFunc(fmt.Sprintf("Dialing SSH to %s...", addr))
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("SSH dial failed: %w", err)
	}
	defer client.Close()
	logFunc("SSH connection established.")

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, _, err := client.SendRequest("keepalive@openssh.com", true, nil)
				if err != nil {
					logFunc(fmt.Sprintf("Keep-alive failed: %v", err))
					return
				}
			}
		}
	}()

	localAddr := "localhost:" + localPort
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("failed to start local listener on %s: %w", localAddr, err)
	}
	defer listener.Close()
	logFunc(fmt.Sprintf("Tunnel listening on %s -> remote:localhost:3389", localAddr))

	go func() {
		for {
			localConn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleForward(client, localConn, logFunc)
		}
	}()

	rdpContent := fmt.Sprintf("full address:s:%s\nusername:s:%s\n", localAddr, user)
	tmpFile, err := os.CreateTemp("", strings.ToLower(AppName)+"-*.rdp")
	if err != nil {
		return fmt.Errorf("failed to create temp rdp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(rdpContent); err != nil {
		return fmt.Errorf("failed to write rdp file: %w", err)
	}
	tmpFile.Close()

	if onReady != nil {
		onReady()
	}

	logFunc(fmt.Sprintf("Launching Remote Desktop Client with %s...", tmpFile.Name()))
	cmd := exec.CommandContext(ctx, "mstsc.exe", tmpFile.Name())
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start mstsc: %w", err)
	}

	err = cmd.Wait()
	logFunc("Remote Desktop Client exited.")
	return err
}

func handleForward(client *ssh.Client, localConn net.Conn, logFunc func(string)) {
	defer localConn.Close()

	remoteConn, err := client.Dial("tcp", "localhost:3389")
	if err != nil {
		logFunc(fmt.Sprintf("Failed to dial remote RDP port: %v", err))
		return
	}
	defer remoteConn.Close()

	logFunc(fmt.Sprintf("Accepted connection from %s", localConn.RemoteAddr()))

	copyConn := func(writer, reader net.Conn, direction string) {
		written, _ := io.Copy(writer, reader)
		logFunc(fmt.Sprintf("Tunnel connection closed (%s). Bytes: %d", direction, written))
	}

	go copyConn(localConn, remoteConn, "Remote->Local")
	copyConn(remoteConn, localConn, "Local->Remote")
}

// ExportPrivateKey exports the private key in PEM format (PKCS#1 for RSA, SEC1 for ECDSA)
func ExportPrivateKey(key interface{}) ([]byte, error) {
	var pemBlock *pem.Block
	switch k := key.(type) {
	case *rsa.PrivateKey:
		pemBlock = &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(k),
		}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, err
		}
		pemBlock = &pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: b,
		}
	default:
		return nil, fmt.Errorf("unsupported key type: %T", key)
	}
	return pem.EncodeToMemory(pemBlock), nil
}

// ExportPublicKey exports the public key in OpenSSH authorized_keys format
func ExportPublicKey(key interface{}) ([]byte, error) {
	var pubKey interface{}
	switch k := key.(type) {
	case *rsa.PrivateKey:
		pubKey = &k.PublicKey
	case *ecdsa.PrivateKey:
		pubKey = &k.PublicKey
	default:
		return nil, fmt.Errorf("unsupported key type: %T", key)
	}

	sshPub, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return nil, err
	}
	return ssh.MarshalAuthorizedKey(sshPub), nil
}
