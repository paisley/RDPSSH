package main

import (
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/pkcs12"
)

// P12Info holds certificate credentials extracted from PKCS#12 files
type P12Info struct {
	PrivateKey  interface{} // *rsa.PrivateKey or *ecdsa.PrivateKey
	Certificate *x509.Certificate
	CommonName  string
	UPN         string // User Principal Name from certificate SANs or CN
}

// ParseP12 reads and decodes a PKCS#12 (.p12/.pfx) certificate file
func ParseP12(path string, password string) (*P12Info, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	pKey, cert, err := pkcs12.Decode(data, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decode p12: %w", err)
	}

	if pKey == nil {
		return nil, fmt.Errorf("no private key found in p12")
	}
	if cert == nil {
		return nil, fmt.Errorf("no certificate found in p12")
	}

	info := &P12Info{
		PrivateKey:  pKey,
		Certificate: cert,
		CommonName:  cert.Subject.CommonName,
	}

	// Extract UPN from certificate SANs (EmailAddresses or DNSNames) or fallback to CN
	if len(cert.EmailAddresses) > 0 {
		info.UPN = cert.EmailAddresses[0]
	} else if len(cert.DNSNames) > 0 {
		info.UPN = cert.DNSNames[0]
	} else if strings.Contains(info.CommonName, "@") {
		info.UPN = info.CommonName
	}

	return info, nil
}
