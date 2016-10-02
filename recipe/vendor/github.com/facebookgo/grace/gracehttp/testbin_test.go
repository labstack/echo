package gracehttp_test

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/facebookgo/grace/gracehttp"
)

func TestMain(m *testing.M) {
	const (
		testbinKey   = "GRACEHTTP_TEST_BIN"
		testbinValue = "1"
	)
	if os.Getenv(testbinKey) == testbinValue {
		testbinMain()
		return
	}
	if err := os.Setenv(testbinKey, testbinValue); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

type response struct {
	Sleep time.Duration
	Pid   int
	Error string `json:",omitempty"`
}

// Wait for 10 consecutive responses from our own pid.
//
// This prevents flaky tests that arise from the fact that we have the
// perfectly acceptable (read: not a bug) condition where both the new and the
// old servers are accepting requests. In fact the amount of time both are
// accepting at the same time and the number of requests that flip flop between
// them is unbounded and in the hands of the various kernels our code tends to
// run on.
//
// In order to combat this, we wait for 10 successful responses from our own
// pid. This is a somewhat reliable way to ensure the old server isn't
// serving anymore.
func wait(wg *sync.WaitGroup, url string) {
	var success int
	defer wg.Done()
	for {
		res, err := http.Get(url)
		if err == nil {
			// ensure it isn't a response from a previous instance
			defer res.Body.Close()
			var r response
			if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
				log.Fatalf("Error decoding json: %s", err)
			}
			if r.Pid == os.Getpid() {
				success++
				if success == 10 {
					return
				}
				continue
			}
		} else {
			success = 0
			// we expect connection refused
			if !strings.HasSuffix(err.Error(), "connection refused") {
				e2 := json.NewEncoder(os.Stderr).Encode(&response{
					Error: err.Error(),
					Pid:   os.Getpid(),
				})
				if e2 != nil {
					log.Fatalf("Error writing error json: %s", e2)
				}
			}
		}
	}
}

func httpsServer(addr string) *http.Server {
	cert, err := tls.X509KeyPair(localhostCert, localhostKey)
	if err != nil {
		log.Fatalf("error loading cert: %v", err)
	}
	return &http.Server{
		Addr:    addr,
		Handler: newHandler(),
		TLSConfig: &tls.Config{
			NextProtos:   []string{"http/1.1"},
			Certificates: []tls.Certificate{cert},
		},
	}
}

func testbinMain() {
	var httpAddr, httpsAddr string
	flag.StringVar(&httpAddr, "http", ":48560", "http address to bind to")
	flag.StringVar(&httpsAddr, "https", ":48561", "https address to bind to")
	flag.Parse()

	// we have self signed certs
	http.DefaultTransport = &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	err := flag.Set("gracehttp.log", "false")
	if err != nil {
		log.Fatalf("Error setting gracehttp.log: %s", err)
	}

	// print json to stderr once we can successfully connect to all three
	// addresses. the ensures we only print the line once we're ready to serve.
	go func() {
		var wg sync.WaitGroup
		wg.Add(2)
		go wait(&wg, fmt.Sprintf("http://%s/sleep/?duration=1ms", httpAddr))
		go wait(&wg, fmt.Sprintf("https://%s/sleep/?duration=1ms", httpsAddr))
		wg.Wait()

		err := json.NewEncoder(os.Stderr).Encode(&response{Pid: os.Getpid()})
		if err != nil {
			log.Fatalf("Error writing startup json: %s", err)
		}
	}()

	err = gracehttp.Serve(
		&http.Server{Addr: httpAddr, Handler: newHandler()},
		httpsServer(httpsAddr),
	)
	if err != nil {
		log.Fatalf("Error in gracehttp.Serve: %s", err)
	}
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/sleep/", func(w http.ResponseWriter, r *http.Request) {
		duration, err := time.ParseDuration(r.FormValue("duration"))
		if err != nil {
			http.Error(w, err.Error(), 400)
		}
		time.Sleep(duration)
		err = json.NewEncoder(w).Encode(&response{
			Sleep: duration,
			Pid:   os.Getpid(),
		})
		if err != nil {
			log.Fatalf("Error encoding json: %s", err)
		}
	})
	return mux
}

// localhostCert is a PEM-encoded TLS cert with SAN IPs
// "127.0.0.1" and "[::1]", expiring at the last second of 2049 (the end
// of ASN.1 time).
// generated from src/pkg/crypto/tls:
// go run generate_cert.go  --rsa-bits 512 --host 127.0.0.1,::1,example.com --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var localhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIIBdzCCASOgAwIBAgIBADALBgkqhkiG9w0BAQUwEjEQMA4GA1UEChMHQWNtZSBD
bzAeFw03MDAxMDEwMDAwMDBaFw00OTEyMzEyMzU5NTlaMBIxEDAOBgNVBAoTB0Fj
bWUgQ28wWjALBgkqhkiG9w0BAQEDSwAwSAJBALyCfqwwip8BvTKgVKGdmjZTU8DD
ndR+WALmFPIRqn89bOU3s30olKiqYEju/SFoEvMyFRT/TWEhXHDaufThqaMCAwEA
AaNoMGYwDgYDVR0PAQH/BAQDAgCkMBMGA1UdJQQMMAoGCCsGAQUFBwMBMA8GA1Ud
EwEB/wQFMAMBAf8wLgYDVR0RBCcwJYILZXhhbXBsZS5jb22HBH8AAAGHEAAAAAAA
AAAAAAAAAAAAAAEwCwYJKoZIhvcNAQEFA0EAr/09uy108p51rheIOSnz4zgduyTl
M+4AmRo8/U1twEZLgfAGG/GZjREv2y4mCEUIM3HebCAqlA5jpRg76Rf8jw==
-----END CERTIFICATE-----`)

// localhostKey is the private key for localhostCert.
var localhostKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIBOQIBAAJBALyCfqwwip8BvTKgVKGdmjZTU8DDndR+WALmFPIRqn89bOU3s30o
lKiqYEju/SFoEvMyFRT/TWEhXHDaufThqaMCAwEAAQJAPXuWUxTV8XyAt8VhNQER
LgzJcUKb9JVsoS1nwXgPksXnPDKnL9ax8VERrdNr+nZbj2Q9cDSXBUovfdtehcdP
qQIhAO48ZsPylbTrmtjDEKiHT2Ik04rLotZYS2U873J6I7WlAiEAypDjYxXyafv/
Yo1pm9onwcetQKMW8CS3AjuV9Axzj6cCIEx2Il19fEMG4zny0WPlmbrcKvD/DpJQ
4FHrzsYlIVTpAiAas7S1uAvneqd0l02HlN9OxQKKlbUNXNme+rnOnOGS2wIgS0jW
zl1jvrOSJeP1PpAHohWz6LOhEr8uvltWkN6x3vE=
-----END RSA PRIVATE KEY-----`)
