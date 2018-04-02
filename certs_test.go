package echo

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"log"
	"math/big"
	"time"
)

var (
	xcaCert            *x509.Certificate
	caCert, serverCert *tls.Certificate
)

func init() {
	caCert, serverCert = makeCerts()
}

func makeCerts() (*tls.Certificate, *tls.Certificate) {
	tmpl := getTMPL()
	tmpl.IsCA = true
	key := genKey()
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, key.Public(), key)
	if err != nil {
		log.Fatal(err)
	}

	tmpCaCert, getCACertErr := x509.ParseCertificate(certDER)
	if getCACertErr != nil {
		log.Fatal(getCACertErr)
	}
	xcaCert = tmpCaCert

	tmpl.IsCA = false
	serverKey := genKey()
	serverCertDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpCaCert, serverKey.Public(), key)
	if err != nil {
		log.Fatal(err)
	}

	return getCert(key, certDER), getCert(serverKey, serverCertDER)
}

func genKey() *ecdsa.PrivateKey {
	priv, genErr := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if genErr != nil {
		log.Fatal(genErr)
	}

	return priv
}

func getTMPL() *x509.Certificate {
	now := time.Now()
	return &x509.Certificate{
		NotBefore:    now,
		NotAfter:     now.Add(time.Hour),
		IsCA:         false,
		DNSNames:     []string{"localhost"},
		SerialNumber: big.NewInt(1),
	}
}

func getCert(key *ecdsa.PrivateKey, certDER []byte) *tls.Certificate {
	keyDER, marshalPrivateErr := x509.MarshalECPrivateKey(key)
	if marshalPrivateErr != nil {
		log.Fatal(marshalPrivateErr)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyDER})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatal(err)
	}

	return &tlsCert
}
