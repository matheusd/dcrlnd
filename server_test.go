// +build !rpctest

package dcrlnd

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"testing"
	"time"
)

func TestParseHexColor(t *testing.T) {
	var colorTestCases = []struct {
		test  string
		valid bool // If valid format
		R     byte
		G     byte
		B     byte
	}{
		{"#123", false, 0, 0, 0},
		{"#1234567", false, 0, 0, 0},
		{"$123456", false, 0, 0, 0},
		{"#12345+", false, 0, 0, 0},
		{"#fFGG00", false, 0, 0, 0},
		{"", false, 0, 0, 0},
		{"#123456", true, 0x12, 0x34, 0x56},
		{"#C0FfeE", true, 0xc0, 0xff, 0xee},
	}

	// Perform the table driven tests.
	for _, ct := range colorTestCases {

		color, err := parseHexColor(ct.test)
		if !ct.valid && err == nil {
			t.Fatalf("Invalid color string: %s, should return "+
				"error, but did not", ct.test)
		}

		if ct.valid && err != nil {
			t.Fatalf("Color %s valid to parse: %s", ct.test, err)
		}

		// Ensure that the string to hex decoding is working properly.
		if color.R != ct.R || color.G != ct.G || color.B != ct.B {
			t.Fatalf("Color %s incorrectly parsed as %v", ct.test, color)
		}
	}
}

// TestTLSAutoRegeneration creates an expired TLS certificate, to test that a
// new TLS certificate pair is regenerated when the old pair expires. This is
// necessary because the pair expires after a little over a year.
func TestTLSAutoRegeneration(t *testing.T) {
	tempDirPath, err := ioutil.TempDir("", ".testLnd")
	if err != nil {
		t.Fatalf("couldn't create temporary cert directory")
	}
	defer os.RemoveAll(tempDirPath)

	certPath := tempDirPath + "/tls.cert"
	keyPath := tempDirPath + "/tls.key"

	certDerBytes, keyBytes := genExpiredCertPair(t, tempDirPath)

	expiredCert, err := x509.ParseCertificate(certDerBytes)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}

	certBuf := bytes.Buffer{}
	err = pem.Encode(
		&certBuf, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certDerBytes,
		},
	)
	if err != nil {
		t.Fatalf("failed to encode certificate: %v", err)
	}

	keyBuf := bytes.Buffer{}
	err = pem.Encode(
		&keyBuf, &pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: keyBytes,
		},
	)
	if err != nil {
		t.Fatalf("failed to encode private key: %v", err)
	}

	// Write cert and key files.
	err = ioutil.WriteFile(tempDirPath+"/tls.cert", certBuf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}
	err = ioutil.WriteFile(tempDirPath+"/tls.key", keyBuf.Bytes(), 0600)
	if err != nil {
		os.Remove(tempDirPath + "tls.cert")
		t.Fatalf("failed to write key file: %v", err)
	}

	rpcListener := net.IPAddr{IP: net.ParseIP("127.0.0.1"), Zone: ""}
	rpcListeners := make([]net.Addr, 0)
	rpcListeners = append(rpcListeners, &rpcListener)

	// Now let's run getTLSConfig. If it works properly, it should delete
	// the cert and create a new one.
	_, _, _, err = getTLSConfig(certPath, keyPath, rpcListeners)
	if err != nil {
		t.Fatalf("couldn't retrieve TLS config")
	}

	// Grab the certificate to test that getTLSConfig did its job correctly
	// and generated a new cert.
	newCertData, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		t.Fatalf("couldn't grab new certificate")
	}

	newCert, err := x509.ParseCertificate(newCertData.Certificate[0])
	if err != nil {
		t.Fatalf("couldn't parse new certificate")
	}

	// Check that the expired certificate was successfully deleted and
	// replaced with a new one.
	if !newCert.NotAfter.After(expiredCert.NotAfter) {
		t.Fatalf("New certificate expiration is too old")
	}
}

// genExpiredCertPair generates an expired key/cert pair to test that expired
// certificates are being regenerated correctly.
func genExpiredCertPair(t *testing.T, certDirPath string) ([]byte, []byte) {
	// Max serial number.
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	// Generate a serial number that's below the serialNumberLimit.
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		t.Fatalf("failed to generate serial number: %s", err)
	}

	host := "lightning"

	// Create a simple ip address for the fake certificate.
	ipAddresses := []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}

	dnsNames := []string{host, "unix", "unixpacket"}

	// Construct the certificate template.
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"lnd autogenerated cert"},
			CommonName:   host,
		},
		NotBefore: time.Now().Add(-time.Hour * 24),
		NotAfter:  time.Now(),

		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:                  true, // so can sign self.
		BasicConstraintsValid: true,

		DNSNames:    dnsNames,
		IPAddresses: ipAddresses,
	}

	// Generate a private key for the certificate.
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate a private key")
	}

	certDerBytes, err := x509.CreateCertificate(
		rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatalf("unable to encode privkey: %v", err)
	}

	return certDerBytes, keyBytes
}
