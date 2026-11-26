// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import (
	stdContext "context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartConfig_StartTLS_DoesNotMutateCallerTLSConfig(t *testing.T) {
	e := New()
	e.GET("/ok", func(c *Context) error {
		return c.String(http.StatusOK, "OK")
	})

	callerTLSConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		NextProtos: []string{"http/1.1"},
	}

	addrChan := make(chan string, 1)
	errCh := make(chan error, 1)
	ctx, cancel := stdContext.WithTimeout(stdContext.Background(), 2*time.Second)
	defer cancel()

	go func() {
		sc := StartConfig{
			Address:         "127.0.0.1:0",
			HideBanner:      true,
			HidePort:        true,
			TLSConfig:       callerTLSConfig,
			GracefulTimeout: 100 * time.Millisecond,
			ListenerAddrFunc: func(addr net.Addr) {
				addrChan <- addr.String()
			},
		}
		errCh <- sc.StartTLS(ctx, e, "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
	}()

	addr, err := waitForServerStart(addrChan, errCh)
	require.NoError(t, err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	res, err := client.Get(fmt.Sprintf("https://%s/ok", addr))
	require.NoError(t, err)
	_ = res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	cancel()
	require.NoError(t, <-errCh)

	assert.Nil(t, callerTLSConfig.Certificates, "StartTLS must not mutate caller-provided TLSConfig.Certificates")
}

func TestStartConfig_StartTLS_SharedTLSConfigConcurrent(t *testing.T) {
	sharedTLSConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		NextProtos: []string{"http/1.1"},
	}

	const numServers = 2
	ctx, cancel := stdContext.WithTimeout(stdContext.Background(), 3*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	addrs := make([]string, numServers)
	errChs := make([]chan error, numServers)
	addrChs := make([]chan string, numServers)

	for i := range numServers {
		e := New()
		e.GET("/ok", func(c *Context) error {
			return c.String(http.StatusOK, "OK")
		})
		addrChs[i] = make(chan string, 1)
		errChs[i] = make(chan error, 1)
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sc := StartConfig{
				Address:         "127.0.0.1:0",
				HideBanner:      true,
				HidePort:        true,
				TLSConfig:       sharedTLSConfig,
				GracefulTimeout: 100 * time.Millisecond,
				ListenerAddrFunc: func(addr net.Addr) {
					addrChs[idx] <- addr.String()
				},
			}
			errChs[idx] <- sc.StartTLS(ctx, e, "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
		}(i)
	}

	for i := range numServers {
		addr, err := waitForServerStart(addrChs[i], errChs[i])
		require.NoError(t, err)
		addrs[i] = addr
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	for _, addr := range addrs {
		res, err := client.Get(fmt.Sprintf("https://%s/ok", addr))
		require.NoError(t, err)
		_ = res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	}

	cancel()
	wg.Wait()
	for i := range numServers {
		require.NoError(t, <-errChs[i])
	}
	assert.Nil(t, sharedTLSConfig.Certificates)
}
