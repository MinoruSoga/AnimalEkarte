package smtp

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModeForPortNormalizesSubmissionPort(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		wantMode transportMode
	}{
		{name: "submission leading plus", port: "+587", wantMode: requiredSTARTTLS},
		{name: "submission leading zeroes", port: "00587", wantMode: requiredSTARTTLS},
		{name: "submission plus and zeroes", port: "+00587", wantMode: requiredSTARTTLS},
		{name: "implicit TLS leading plus", port: "+465", wantMode: implicitTLS},
		{name: "implicit TLS leading zeroes", port: "00465", wantMode: implicitTLS},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantMode, modeForPort(tt.port))
		})
	}
}

func TestModeForConfigResolvesImplicitTLSServiceAliases(t *testing.T) {
	for _, alias := range []string{"submissions", "ssmtp", "urd"} {
		t.Run(alias, func(t *testing.T) {
			mode, err := modeForConfigWithLookup(Config{
				Host:                  "127.0.0.1",
				Port:                  alias,
				AllowInsecureLoopback: true,
			}, func(network, service string) (int, error) {
				assert.Equal(t, "tcp", network)
				assert.Equal(t, alias, service)
				return 465, nil
			})

			require.NoError(t, err)
			assert.Equal(t, implicitTLS, mode)
		})
	}
}

func TestModeForConfigFailsClosed(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		wantMode transportMode
		wantErr  bool
	}{
		{
			name:     "unknown remote port requires STARTTLS",
			config:   Config{Host: "smtp.example.com", Port: "2525"},
			wantMode: requiredSTARTTLS,
		},
		{
			name: "explicit loopback development mode permits plaintext",
			config: Config{
				Host:                  "127.0.0.1",
				Port:                  "2525",
				AllowInsecureLoopback: true,
			},
			wantMode: opportunisticSTARTTLS,
		},
		{
			name: "submission port cannot be weakened",
			config: Config{
				Host:                  "127.0.0.1",
				Port:                  "+00587",
				AllowInsecureLoopback: true,
			},
			wantMode: requiredSTARTTLS,
		},
		{
			name: "implicit TLS port cannot be weakened",
			config: Config{
				Host:                  "127.0.0.1",
				Port:                  "+00465",
				AllowInsecureLoopback: true,
			},
			wantMode: implicitTLS,
		},
		{
			name: "implicit TLS service alias cannot be weakened",
			config: Config{
				Host:                  "127.0.0.1",
				Port:                  "submissions",
				AllowInsecureLoopback: true,
			},
			wantMode: implicitTLS,
		},
		{
			name: "non-loopback development mode is rejected",
			config: Config{
				Host:                  "smtp.example.com",
				Port:                  "2525",
				AllowInsecureLoopback: true,
			},
			wantMode: requiredSTARTTLS,
			wantErr:  true,
		},
		{
			name:     "unknown service name is rejected before dial",
			config:   Config{Host: "smtp.example.com", Port: "not-a-real-smtp-service"},
			wantMode: requiredSTARTTLS,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode, err := modeForConfig(tt.config)

			assert.Equal(t, tt.wantMode, mode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSendRejectsIncompleteAuthenticationBeforeDial(t *testing.T) {
	tests := []struct {
		name string
		user string
		pass string
	}{
		{name: "user only", user: "mailer"},
		{name: "password only", pass: "test-password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			t.Cleanup(func() { _ = listener.Close() })
			tcpListener := listener.(*net.TCPListener)
			require.NoError(t, tcpListener.SetDeadline(time.Now().Add(150*time.Millisecond)))

			connectionAccepted := make(chan bool, 1)
			go func() {
				connection, acceptErr := listener.Accept()
				if acceptErr != nil {
					connectionAccepted <- false
					return
				}
				_ = connection.Close()
				connectionAccepted <- true
			}()

			host, port, err := net.SplitHostPort(listener.Addr().String())
			require.NoError(t, err)
			sendErr := Send(context.Background(), Config{
				Host: host,
				Port: port,
				User: tt.user,
				Pass: tt.pass,
			}, "sender@example.com", "owner@example.com", nil)

			require.Error(t, sendErr)
			assert.Contains(t, sendErr.Error(), "user and password")
			assert.False(t, <-connectionAccepted)
		})
	}
}

func TestSendRejectsEnvelopeInjectionBeforeDial(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
	}{
		{
			name: "mail from",
			from: "sender@example.com\r\nRCPT TO:<attacker@example.com>",
			to:   "owner@example.com",
		},
		{
			name: "recipient",
			from: "sender@example.com",
			to:   "owner@example.com\nDATA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Send(context.Background(), Config{
				Host: "127.0.0.1",
				Port: "1",
			}, tt.from, tt.to, nil)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "CR or LF")
		})
	}
}

func TestNewClientImplicitTLSPolicy(t *testing.T) {
	serverCertificate, roots := newTestServerCertificate(t, "smtp.test")
	tests := []struct {
		name             string
		serverMaxVersion uint16
		serverName       string
		wantErr          bool
		wantHostnameErr  bool
	}{
		{
			name:             "TLS 1.1 is rejected",
			serverMaxVersion: tls.VersionTLS11,
			serverName:       "smtp.test",
			wantErr:          true,
		},
		{
			name:             "TLS 1.2 succeeds",
			serverMaxVersion: tls.VersionTLS12,
			serverName:       "smtp.test",
		},
		{
			name:             "certificate SAN mismatch is rejected",
			serverMaxVersion: tls.VersionTLS12,
			serverName:       "wrong.test",
			wantErr:          true,
			wantHostnameErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientConnection, serverConnection := net.Pipe()
			t.Cleanup(func() {
				_ = clientConnection.Close()
				_ = serverConnection.Close()
			})
			deadline := time.Now().Add(time.Second)
			require.NoError(t, clientConnection.SetDeadline(deadline))
			require.NoError(t, serverConnection.SetDeadline(deadline))

			serverResult := serveImplicitTLSSMTP(
				serverConnection,
				serverCertificate,
				tt.serverMaxVersion,
			)
			clientTLS := strictTLSConfig(tt.serverName)
			clientTLS.RootCAs = roots
			assert.Equal(t, uint16(tls.VersionTLS12), clientTLS.MinVersion)
			assert.Equal(t, tt.serverName, clientTLS.ServerName)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			client, err := newClientWithTLSConfig(
				ctx,
				clientConnection,
				tt.serverName,
				implicitTLS,
				clientTLS,
			)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, client)
				if tt.wantHostnameErr {
					var hostnameErr x509.HostnameError
					assert.ErrorAs(t, err, &hostnameErr)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, client)
				_ = client.Close()
			}
			<-serverResult
		})
	}
}

func TestSendRequiresSTARTTLSOnUnknownPortByDefault(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	commandAfterEHLO := serveSMTPWithoutExtensions(listener)
	host, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)

	sendErr := Send(context.Background(), Config{
		Host: host,
		Port: port,
	}, "sender@example.com", "owner@example.com", nil)

	require.Error(t, sendErr)
	assert.Contains(t, sendErr.Error(), "STARTTLS")
	assert.Empty(t, <-commandAfterEHLO)
}

func TestSendRejectsConfiguredAuthenticationWhenServerDoesNotAdvertiseAUTH(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	commandAfterEHLO := serveSMTPWithoutExtensions(listener)
	host, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)

	sendErr := Send(context.Background(), Config{
		Host:                  host,
		Port:                  port,
		User:                  "mailer",
		Pass:                  "secret",
		AllowInsecureLoopback: true,
	}, "sender@example.com", "owner@example.com", nil)

	require.Error(t, sendErr)
	assert.Contains(t, sendErr.Error(), "AUTH")
	assert.Empty(t, <-commandAfterEHLO)
}

func TestSendFailsClosedWhenAdvertisedSTARTTLSHandshakeFails(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	observedCommand := serveBrokenSTARTTLS(listener)
	host, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)

	sendErr := Send(context.Background(), Config{
		Host: host,
		Port: port,
	}, "sender@example.com", "owner@example.com", nil)

	require.Error(t, sendErr)
	assert.Equal(t, "STARTTLS", <-observedCommand)
}

func TestSendAllowsUnauthenticatedDevelopmentPortWithoutSTARTTLS(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	delivered := servePlainSMTP(listener)
	host, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)

	sendErr := Send(context.Background(), Config{
		Host:                  host,
		Port:                  port,
		AllowInsecureLoopback: true,
	}, "sender@example.com", "owner@example.com", []byte(
		"Subject: development delivery\r\n\r\nbody",
	))

	require.NoError(t, sendErr)
	result := <-delivered
	require.NoError(t, result.err)
	assert.Contains(t, result.message, "Subject: development delivery")
	assert.Contains(t, result.message, "body")
}

func TestSendRejectsInsecureDevelopmentModeForNonLoopbackHostBeforeDial(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	tcpListener := listener.(*net.TCPListener)
	require.NoError(t, tcpListener.SetDeadline(time.Now().Add(150*time.Millisecond)))

	connectionAccepted := make(chan bool, 1)
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			connectionAccepted <- false
			return
		}
		_ = connection.Close()
		connectionAccepted <- true
	}()

	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	sendErr := Send(context.Background(), Config{
		Host:                  "0.0.0.0",
		Port:                  port,
		AllowInsecureLoopback: true,
	}, "sender@example.com", "owner@example.com", nil)

	require.Error(t, sendErr)
	assert.Contains(t, sendErr.Error(), "loopback")
	assert.False(t, <-connectionAccepted)
}

func TestSendRespectsContextDuringSMTPGreeting(t *testing.T) {
	tests := []struct {
		name              string
		newContext        func() (context.Context, context.CancelFunc)
		cancelAfterAccept bool
		wantDeadline      bool
		wantContextErr    error
	}{
		{
			name: "deadline",
			newContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 100*time.Millisecond)
			},
			wantDeadline:   true,
			wantContextErr: context.DeadlineExceeded,
		},
		{
			name: "cancellation without deadline",
			newContext: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			cancelAfterAccept: true,
			wantDeadline:      false,
			wantContextErr:    context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			t.Cleanup(func() { _ = listener.Close() })

			accepted := make(chan struct{})
			go func() {
				connection, acceptErr := listener.Accept()
				if acceptErr != nil {
					return
				}
				defer connection.Close()
				close(accepted)
				_, _ = io.Copy(io.Discard, connection)
			}()

			host, port, err := net.SplitHostPort(listener.Addr().String())
			require.NoError(t, err)
			ctx, cancel := tt.newContext()
			defer cancel()
			_, hasDeadline := ctx.Deadline()
			assert.Equal(t, tt.wantDeadline, hasDeadline)

			sendResult := make(chan error, 1)
			startedAt := time.Now()
			go func() {
				sendResult <- Send(ctx, Config{
					Host: host,
					Port: port,
				}, "sender@example.com", "owner@example.com", nil)
			}()

			<-accepted
			if tt.cancelAfterAccept {
				cancel()
			}

			select {
			case sendErr := <-sendResult:
				require.Error(t, sendErr)
				assert.True(t, errors.Is(sendErr, tt.wantContextErr))
				assert.Less(t, time.Since(startedAt), time.Second)
			case <-time.After(time.Second):
				t.Fatal("Send did not return after context cancellation")
			}
		})
	}
}

func TestNewClientPropagatesEHLOCommunicationError(t *testing.T) {
	clientConnection, serverConnection := net.Pipe()
	t.Cleanup(func() {
		_ = clientConnection.Close()
		_ = serverConnection.Close()
	})

	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		reader := bufio.NewReader(serverConnection)
		writer := bufio.NewWriter(serverConnection)
		if writeSMTPResponse(writer, "220 localhost ESMTP\r\n") != nil {
			return
		}
		_, _ = reader.ReadString('\n')
		_ = serverConnection.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	client, err := newClient(ctx, clientConnection, "localhost", opportunisticSTARTTLS)

	require.Error(t, err)
	assert.Nil(t, client)
	<-serverDone
}

func newTestServerCertificate(
	t *testing.T,
	dnsName string,
) (tls.Certificate, *x509.CertPool) {
	t.Helper()

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	now := time.Now()
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "SMTP test CA"},
		NotBefore:             now.Add(-time.Minute),
		NotAfter:              now.Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(
		rand.Reader,
		caTemplate,
		caTemplate,
		&caKey.PublicKey,
		caKey,
	)
	require.NoError(t, err)
	caCertificate, err := x509.ParseCertificate(caDER)
	require.NoError(t, err)

	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: dnsName},
		DNSNames:     []string{dnsName},
		NotBefore:    now.Add(-time.Minute),
		NotAfter:     now.Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	serverDER, err := x509.CreateCertificate(
		rand.Reader,
		serverTemplate,
		caCertificate,
		&serverKey.PublicKey,
		caKey,
	)
	require.NoError(t, err)

	roots := x509.NewCertPool()
	roots.AddCert(caCertificate)
	return tls.Certificate{
		Certificate: [][]byte{serverDER, caDER},
		PrivateKey:  serverKey,
	}, roots
}

func serveImplicitTLSSMTP(
	connection net.Conn,
	certificate tls.Certificate,
	maxVersion uint16,
) <-chan error {
	result := make(chan error, 1)
	go func() {
		defer connection.Close()
		tlsConnection := tls.Server(connection, &tls.Config{
			Certificates: []tls.Certificate{certificate},
			MinVersion:   tls.VersionTLS10,
			MaxVersion:   maxVersion,
		})
		if err := tlsConnection.Handshake(); err != nil {
			result <- err
			return
		}

		reader := bufio.NewReader(tlsConnection)
		writer := bufio.NewWriter(tlsConnection)
		if err := writeSMTPResponse(writer, "220 localhost ESMTP\r\n"); err != nil {
			result <- err
			return
		}
		if err := expectSMTPCommand(reader, "EHLO "); err != nil {
			result <- err
			return
		}
		result <- writeSMTPResponse(writer, "250 localhost\r\n")
	}()
	return result
}

func serveSMTPWithoutExtensions(listener net.Listener) <-chan string {
	commandAfterEHLO := make(chan string, 1)
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			commandAfterEHLO <- ""
			return
		}
		defer connection.Close()
		_ = connection.SetDeadline(time.Now().Add(time.Second))

		reader := bufio.NewReader(connection)
		writer := bufio.NewWriter(connection)
		if _, writeErr := fmt.Fprint(writer, "220 localhost ESMTP\r\n"); writeErr != nil {
			commandAfterEHLO <- ""
			return
		}
		if flushErr := writer.Flush(); flushErr != nil {
			commandAfterEHLO <- ""
			return
		}

		command, readErr := reader.ReadString('\n')
		if readErr != nil || !strings.HasPrefix(command, "EHLO ") {
			commandAfterEHLO <- strings.TrimSpace(command)
			return
		}
		_, _ = fmt.Fprint(writer, "250-localhost\r\n250 PIPELINING\r\n")
		if flushErr := writer.Flush(); flushErr != nil {
			commandAfterEHLO <- ""
			return
		}

		command, readErr = reader.ReadString('\n')
		commandAfterEHLO <- strings.TrimSpace(command)
		if readErr == nil {
			_, _ = fmt.Fprint(writer, "550 plaintext command rejected\r\n")
			_ = writer.Flush()
		}
	}()
	return commandAfterEHLO
}

func serveBrokenSTARTTLS(listener net.Listener) <-chan string {
	observedCommand := make(chan string, 1)
	go func() {
		connection, err := listener.Accept()
		if err != nil {
			observedCommand <- ""
			return
		}
		defer connection.Close()
		_ = connection.SetDeadline(time.Now().Add(time.Second))

		reader := bufio.NewReader(connection)
		writer := bufio.NewWriter(connection)
		if err := writeSMTPResponse(writer, "220 localhost ESMTP\r\n"); err != nil {
			observedCommand <- ""
			return
		}
		if err := expectSMTPCommand(reader, "EHLO "); err != nil {
			observedCommand <- ""
			return
		}
		if err := writeSMTPResponse(
			writer,
			"250-localhost\r\n250-STARTTLS\r\n250 PIPELINING\r\n",
		); err != nil {
			observedCommand <- ""
			return
		}

		command, readErr := reader.ReadString('\n')
		observedCommand <- strings.TrimSpace(command)
		if readErr == nil {
			_ = writeSMTPResponse(writer, "220 ready to start TLS\r\n")
		}
	}()
	return observedCommand
}

type deliveryResult struct {
	message string
	err     error
}

func servePlainSMTP(listener net.Listener) <-chan deliveryResult {
	delivered := make(chan deliveryResult, 1)
	go func() {
		connection, err := listener.Accept()
		if err != nil {
			delivered <- deliveryResult{err: err}
			return
		}
		defer connection.Close()
		_ = connection.SetDeadline(time.Now().Add(time.Second))

		reader := bufio.NewReader(connection)
		writer := bufio.NewWriter(connection)
		if err := writeSMTPResponse(writer, "220 localhost ESMTP\r\n"); err != nil {
			delivered <- deliveryResult{err: err}
			return
		}
		if err := expectSMTPCommand(reader, "EHLO "); err != nil {
			delivered <- deliveryResult{err: err}
			return
		}
		if err := writeSMTPResponse(writer, "250-localhost\r\n250 PIPELINING\r\n"); err != nil {
			delivered <- deliveryResult{err: err}
			return
		}

		for _, command := range []string{"MAIL FROM:", "RCPT TO:", "DATA"} {
			if err := expectSMTPCommand(reader, command); err != nil {
				delivered <- deliveryResult{err: err}
				return
			}
			response := "250 OK\r\n"
			if command == "DATA" {
				response = "354 End data with <CR><LF>.<CR><LF>\r\n"
			}
			if err := writeSMTPResponse(writer, response); err != nil {
				delivered <- deliveryResult{err: err}
				return
			}
		}

		var message strings.Builder
		for {
			line, readErr := reader.ReadString('\n')
			if readErr != nil {
				delivered <- deliveryResult{err: readErr}
				return
			}
			if line == ".\r\n" {
				break
			}
			message.WriteString(line)
		}
		if err := writeSMTPResponse(writer, "250 queued\r\n"); err != nil {
			delivered <- deliveryResult{err: err}
			return
		}
		if err := expectSMTPCommand(reader, "QUIT"); err != nil {
			delivered <- deliveryResult{err: err}
			return
		}
		if err := writeSMTPResponse(writer, "221 bye\r\n"); err != nil {
			delivered <- deliveryResult{err: err}
			return
		}
		delivered <- deliveryResult{message: message.String()}
	}()
	return delivered
}

func expectSMTPCommand(reader *bufio.Reader, prefix string) error {
	command, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if !strings.HasPrefix(command, prefix) {
		return fmt.Errorf("expected SMTP command %q, got %q", prefix, strings.TrimSpace(command))
	}
	return nil
}

func writeSMTPResponse(writer *bufio.Writer, response string) error {
	if _, err := fmt.Fprint(writer, response); err != nil {
		return err
	}
	return writer.Flush()
}

func TestValidateEnvelopeAddress(t *testing.T) {
	assert.NoError(t, ValidateEnvelopeAddress("owner@example.com"))
	assert.Error(t, ValidateEnvelopeAddress("owner@example.com\rSubject: injected"))
	assert.Error(t, ValidateEnvelopeAddress("owner@example.com\nSubject: injected"))
}
