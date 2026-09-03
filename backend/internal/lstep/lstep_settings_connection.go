package lstep

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/line"
	infralstep "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

const connectionProbeTimeout = 10 * time.Second

var (
	errConnectionRedirect = errors.New("redirect disallowed")
	errBlockedDialAddress = errors.New("blocked dial address")
	lookupProbeIPAddr     = defaultLookupProbeIPAddr
	connectionProbeClient = newOutboundProbeClient(connectionProbeTimeout)
)

func defaultLookupProbeIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	return net.DefaultResolver.LookupIPAddr(ctx, host)
}

func probeCheckRedirect(*http.Request, []*http.Request) error {
	return errConnectionRedirect
}

func newOutboundProbeClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:       timeout,
		CheckRedirect: probeCheckRedirect,
		Transport: &http.Transport{
			Proxy:                 nil,
			DialContext:           probeDialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          4,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   timeout,
			ExpectContinueTimeout: time.Second,
		},
	}
}

func probeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := lookupProbeIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, errBlockedDialAddress
	}
	for _, ipa := range ips {
		if isForbiddenProbeIP(ipa.IP) {
			return nil, errBlockedDialAddress
		}
	}
	d := &net.Dialer{Timeout: connectionProbeTimeout}
	return d.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
}

func isForbiddenProbeIP(ip net.IP) bool {
	return infralstep.IsForbiddenDialIP(ip)
}

func (s *lstepSettingsService) TestConnection(ctx context.Context, clinicID uint64) (*LstepConnectionTestResult, error) {
	records, err := s.repo.FindByClinicAndService(ctx, clinicID, model.IntegrationServiceLstep)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find lstep settings for test", "error", err)
		return nil, apperrors.Wrap(err, "failed to load settings for connection test")
	}

	kvMap := make(map[string]string, len(records))
	for _, r := range records {
		val, decErr := s.decrypt(r.KeyName, r.KeyValue)
		if decErr != nil {
			// LSB-04 / DEC-35: 復号失敗を空文字へ置換して握り潰さない（サイレント停止を防ぐ）
			slog.ErrorContext(ctx, "failed to decrypt integration value", "key_name", r.KeyName, "error", decErr)
			return nil, apperrors.Wrap(decErr, "failed to decrypt integration value")
		}
		kvMap[r.KeyName] = val
	}

	result := &LstepConnectionTestResult{}

	// Lステップ疎通確認 — only allowlisted base URL (LSA-01)
	lstepKey := kvMap[model.IntegrationKeyLstepAPIKey]
	lstepBase, baseErr := ValidateLstepBaseURL(kvMap[model.IntegrationKeyLstepBaseURL])
	if lstepKey != "" {
		if baseErr != nil {
			slog.WarnContext(ctx, "lstep connection test rejected base URL", "error", baseErr, "clinic_id", clinicID)
			result.LstepOK = false
			result.LstepError = "invalid_base_url"
		} else if err := testLstepAPI(ctx, lstepBase, lstepKey); err != nil {
			slog.WarnContext(ctx, "lstep connection test failed", "error", err, "clinic_id", clinicID)
			result.LstepOK = false
			result.LstepError = classifyConnectionProbeError(err)
		} else {
			result.LstepOK = true
		}
	}

	// LINE Messaging API疎通確認
	lineToken := kvMap[model.IntegrationKeyLineChannelAccessToken]
	if lineToken != "" {
		if err := testLineAPI(ctx, line.APIHost, lineToken); err != nil {
			slog.WarnContext(ctx, "line connection test failed", "error", err, "clinic_id", clinicID)
			result.LineOK = false
			result.LineError = classifyConnectionProbeError(err)
		} else {
			result.LineOK = true
		}
	}

	return result, nil
}

func testLstepAPI(ctx context.Context, baseURL, apiKey string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v1/tags", http.NoBody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := connectionProbeClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()               //nolint:errcheck // close error on connectivity probe is not actionable
	_, _ = io.Copy(io.Discard, resp.Body) //nolint:errcheck // body drain failure on connectivity probe is not actionable
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return errConnectionUnauthorized
	}
	return nil
}

func testLineAPI(ctx context.Context, baseURL, channelToken string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v2/bot/info", http.NoBody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+channelToken)
	resp, err := connectionProbeClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()               //nolint:errcheck // close error on connectivity probe is not actionable
	_, _ = io.Copy(io.Discard, resp.Body) //nolint:errcheck // body drain failure on connectivity probe is not actionable
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return errConnectionUnauthorized
	}
	return nil
}

var errConnectionUnauthorized = errors.New("authentication failed")

// classifyConnectionProbeError returns stable codes for JSON (LSA-08); raw details stay in slog.
func classifyConnectionProbeError(err error) string {
	if err == nil {
		return "unreachable"
	}
	if errors.Is(err, errConnectionUnauthorized) {
		return "unauthorized"
	}
	if errors.Is(err, errConnectionRedirect) || errors.Is(err, errBlockedDialAddress) {
		return "unreachable"
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return "timeout"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout"
	}
	// Fallback only: remaining Contains must not proliferate. Prefer typed errors.Is/As above.
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline"):
		return "timeout"
	case strings.Contains(msg, "unauthorized") || strings.Contains(msg, "forbidden"):
		return "unauthorized"
	default:
		return "unreachable"
	}
}
