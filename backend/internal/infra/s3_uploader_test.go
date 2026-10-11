package infra

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ FileUploader = (*S3Uploader)(nil)
var _ FileUploader = (*LocalUploader)(nil)

func TestS3Uploader_GetSignedURL_NilPresignFailClosed(t *testing.T) {
	u := &S3Uploader{}

	got, err := u.GetSignedURL(context.Background(), "medical-records/5/photo.png", 15*time.Minute)
	if err == nil {
		t.Fatalf("GetSignedURL succeeded with nil presign: %q", got)
	}
	if got != "" {
		t.Fatalf("fail-closed GetSignedURL must not return a URL, got %q", got)
	}
}

// stubS3HTTPClient は aws.HTTPClient (Do) を満たすレスポンススタブ。
// HeadObject の存在確認経路をネットワーク無しで検証する。
type stubS3HTTPClient struct {
	status string
	code   int
	body   string

	bucketStatus string
	bucketCode   int
	bucketBody   string
	bucketErr    error
	objectCalls  int
	bucketCalls  int
}

func (s *stubS3HTTPClient) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "" || req.URL.Path == "/" {
		s.bucketCalls++
		if s.bucketErr != nil {
			return nil, s.bucketErr
		}
		return &http.Response{
			Status:        s.bucketStatus,
			StatusCode:    s.bucketCode,
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Body:          io.NopCloser(strings.NewReader(s.bucketBody)),
			Header:        make(http.Header),
			ContentLength: int64(len(s.bucketBody)),
			Request:       req,
		}, nil
	}
	s.objectCalls++
	return &http.Response{
		Status:        s.status,
		StatusCode:    s.code,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          io.NopCloser(strings.NewReader(s.body)),
		Header:        make(http.Header),
		ContentLength: int64(len(s.body)),
		Request:       req,
	}, nil
}

type failingS3HTTPClient struct{ err error }

func (f *failingS3HTTPClient) Do(*http.Request) (*http.Response, error) {
	return nil, f.err
}

type failingCredentialsProvider struct{ err error }

func (f failingCredentialsProvider) Retrieve(context.Context) (aws.Credentials, error) {
	return aws.Credentials{}, f.err
}

func newStubbedS3Uploader(t *testing.T, stub *stubS3HTTPClient) *S3Uploader {
	t.Helper()
	return &S3Uploader{
		client: s3.New(s3.Options{
			Region:      "us-east-1",
			Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("AKID", "SECRET", "TOKEN")),
			HTTPClient:  stub,
		}),
		bucket: "test-bucket",
	}
}

func TestS3Uploader_Exists(t *testing.T) {
	t.Run("200 -> true", func(t *testing.T) {
		stub := &stubS3HTTPClient{status: "200 OK", code: 200}
		exists, err := newStubbedS3Uploader(t, stub).Exists(context.Background(), "k")
		if err != nil || !exists {
			t.Fatalf("Exists(200) = (%v, %v), want (true, nil)", exists, err)
		}
		if stub.bucketCalls != 0 {
			t.Fatalf("Exists(200) probed HeadBucket %d times; success must not probe", stub.bucketCalls)
		}
	})

	t.Run("404 NotFound + bucket 200 -> (false, nil)", func(t *testing.T) {
		stub := &stubS3HTTPClient{
			status:       "404 NotFound",
			code:         404,
			body:         `<?xml version="1.0" encoding="UTF-8"?><Error><Code>NotFound</Code><Message>nf</Message></Error>`,
			bucketStatus: "200 OK",
			bucketCode:   200,
		}
		exists, err := newStubbedS3Uploader(t, stub).Exists(context.Background(), "k")
		if err != nil || exists {
			t.Fatalf("Exists(404+bucket200) = (%v, %v), want (false, nil)", exists, err)
		}
		if stub.bucketCalls != 1 {
			t.Fatalf("ambiguous NotFound must probe HeadBucket once, got %d", stub.bucketCalls)
		}
	})

	t.Run("NoSuchKey -> (false, nil)", func(t *testing.T) {
		stub := &stubS3HTTPClient{
			status: "404 NotFound",
			code:   404,
			body:   `<?xml version="1.0" encoding="UTF-8"?><Error><Code>NoSuchKey</Code><Message>nf</Message></Error>`,
		}
		exists, err := newStubbedS3Uploader(t, stub).Exists(context.Background(), "k")
		if err != nil || exists {
			t.Fatalf("Exists(NoSuchKey) = (%v, %v), want (false, nil)", exists, err)
		}
		if stub.bucketCalls != 0 {
			t.Fatalf("explicit NoSuchKey must not probe HeadBucket, got %d calls", stub.bucketCalls)
		}
	})

	t.Run("500 -> error (fail-closed)", func(t *testing.T) {
		stub := &stubS3HTTPClient{
			status: "500 Internal Server Error",
			code:   500,
			body:   `<?xml version="1.0" encoding="UTF-8"?><Error><Code>InternalError</Code><Message>boom</Message></Error>`,
		}
		exists, err := newStubbedS3Uploader(t, stub).Exists(context.Background(), "k")
		if err == nil {
			t.Fatalf("Exists(500) = (%v, nil), want error", exists)
		}
		if exists {
			t.Fatalf("Exists(500) returned true; fail-closed must return false, error")
		}
	})

	t.Run("object 404 empty + bucket 200 -> (false, nil)", func(t *testing.T) {
		stub := &stubS3HTTPClient{status: "404 Not Found", code: 404, bucketStatus: "200 OK", bucketCode: 200}
		exists, err := newStubbedS3Uploader(t, stub).Exists(context.Background(), "k")
		if err != nil || exists {
			t.Fatalf("Exists(object404+bucket200) = (%v, %v), want (false, nil)", exists, err)
		}
		if stub.bucketCalls != 1 {
			t.Fatalf("empty-body 404 must probe HeadBucket once, got %d", stub.bucketCalls)
		}
	})

	t.Run("object 404 empty + bucket 404 empty -> error (missing bucket)", func(t *testing.T) {
		stub := &stubS3HTTPClient{status: "404 Not Found", code: 404, bucketStatus: "404 Not Found", bucketCode: 404}
		exists, err := newStubbedS3Uploader(t, stub).Exists(context.Background(), "k")
		if err == nil {
			t.Fatalf("Exists(object404+bucket404) = (%v, nil), want error — real missing bucket must propagate", exists)
		}
		if exists {
			t.Fatal("Exists(object404+bucket404) returned true")
		}
		if stub.bucketCalls != 1 {
			t.Fatalf("empty-body 404 must probe HeadBucket once, got %d", stub.bucketCalls)
		}
	})

	t.Run("object 404 empty + bucket 403 -> error (bucket denied)", func(t *testing.T) {
		stub := &stubS3HTTPClient{status: "404 Not Found", code: 404, bucketStatus: "403 Forbidden", bucketCode: 403}
		exists, err := newStubbedS3Uploader(t, stub).Exists(context.Background(), "k")
		if err == nil {
			t.Fatalf("Exists(object404+bucket403) = (%v, nil), want error — denied bucket probe must propagate", exists)
		}
		if exists {
			t.Fatal("Exists(object404+bucket403) returned true")
		}
	})

	t.Run("object 404 empty + bucket transport error -> error", func(t *testing.T) {
		stub := &stubS3HTTPClient{
			status:    "404 Not Found",
			code:      404,
			bucketErr: errors.New("dial tcp: connection refused"),
		}
		exists, err := newStubbedS3Uploader(t, stub).Exists(context.Background(), "k")
		if err == nil {
			t.Fatalf("Exists(object404+bucketErr) = (%v, nil), want error", exists)
		}
		if exists {
			t.Fatal("Exists(object404+bucketErr) returned true")
		}
	})

	t.Run("NoSuchBucket -> error (fail-closed)", func(t *testing.T) {
		stub := &stubS3HTTPClient{
			status: "404 Not Found",
			code:   404,
			body:   `<?xml version="1.0" encoding="UTF-8"?><Error><Code>NoSuchBucket</Code><Message>nb</Message></Error>`,
		}
		exists, err := newStubbedS3Uploader(t, stub).Exists(context.Background(), "k")
		if err == nil {
			t.Fatalf("Exists(NoSuchBucket) = (%v, nil), want error — bucket errors must propagate", exists)
		}
		if exists {
			t.Fatal("Exists(NoSuchBucket) returned true")
		}
		if stub.bucketCalls != 0 {
			t.Fatalf("NoSuchBucket must propagate without HeadBucket probe, got %d calls", stub.bucketCalls)
		}
	})

	t.Run("AccessDenied 403 -> error (fail-closed)", func(t *testing.T) {
		stub := &stubS3HTTPClient{
			status: "403 Forbidden",
			code:   403,
			body:   `<?xml version="1.0" encoding="UTF-8"?><Error><Code>AccessDenied</Code><Message>denied</Message></Error>`,
		}
		exists, err := newStubbedS3Uploader(t, stub).Exists(context.Background(), "k")
		if err == nil {
			t.Fatalf("Exists(AccessDenied) = (%v, nil), want error", exists)
		}
	})

	t.Run("unknown API error -> error (fail-closed)", func(t *testing.T) {
		stub := &stubS3HTTPClient{
			status: "400 Bad Request",
			code:   400,
			body:   `<?xml version="1.0" encoding="UTF-8"?><Error><Code>SignatureDoesNotMatch</Code><Message>sig</Message></Error>`,
		}
		exists, err := newStubbedS3Uploader(t, stub).Exists(context.Background(), "k")
		if err == nil {
			t.Fatalf("Exists(unknown api error) = (%v, nil), want error", exists)
		}
	})

	t.Run("transport error -> error (fail-closed)", func(t *testing.T) {
		u := &S3Uploader{
			client: s3.New(s3.Options{
				Region:      "us-east-1",
				Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("AKID", "SECRET", "TOKEN")),
				HTTPClient:  &failingS3HTTPClient{err: errors.New("dial tcp: connection refused")},
			}),
			bucket: "test-bucket",
		}
		exists, err := u.Exists(context.Background(), "k")
		if err == nil {
			t.Fatalf("Exists(transport error) = (%v, nil), want error", exists)
		}
		if exists {
			t.Fatal("Exists(transport error) returned true")
		}
	})

	t.Run("credentials error -> error (fail-closed)", func(t *testing.T) {
		u := &S3Uploader{
			client: s3.New(s3.Options{
				Region:      "us-east-1",
				Credentials: aws.NewCredentialsCache(failingCredentialsProvider{err: errors.New("no credentials")}),
				HTTPClient:  &stubS3HTTPClient{status: "200 OK", code: 200},
			}),
			bucket: "test-bucket",
		}
		exists, err := u.Exists(context.Background(), "k")
		if err == nil {
			t.Fatalf("Exists(credentials error) = (%v, nil), want error", exists)
		}
		if exists {
			t.Fatal("Exists(credentials error) returned true")
		}
	})

	t.Run("nil client -> error (fail-closed)", func(t *testing.T) {
		u := &S3Uploader{}
		if _, err := u.Exists(context.Background(), "k"); err == nil {
			t.Fatal("Exists with nil client must fail closed")
		}
	})
}

func TestS3Uploader_Exists_PathStyleRouting(t *testing.T) {
	const bucket = "fixture-bucket"
	const key = "k"

	type wireRequest struct {
		method string
		host   string
		path   string
	}
	var mu sync.Mutex
	var wire []wireRequest
	var unexpected []string
	bucketCode := http.StatusOK

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimSuffix(r.URL.Path, "/")
		mu.Lock()
		wire = append(wire, wireRequest{method: r.Method, host: r.Host, path: r.URL.Path})
		switch {
		case p == "/"+bucket:
			code := bucketCode
			mu.Unlock()
			w.WriteHeader(code)
		case strings.HasPrefix(p, "/"+bucket+"/"):
			mu.Unlock()
			w.WriteHeader(http.StatusNotFound)
		default:
			unexpected = append(unexpected, r.Method+" "+r.URL.Path)
			mu.Unlock()
			w.WriteHeader(http.StatusNotImplemented)
		}
	}))
	t.Cleanup(srv.Close)
	endpointHost := strings.TrimPrefix(srv.URL, "http://")

	u := &S3Uploader{
		client: s3.New(s3.Options{
			Region:       "us-east-1",
			Credentials:  aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("AKID", "SECRET", "TOKEN")),
			BaseEndpoint: aws.String(srv.URL),
			UsePathStyle: true,
		}),
		bucket: bucket,
	}

	for _, tc := range []struct {
		name       string
		bucketCode int
		wantErr    bool
	}{
		{"empty object404 + path-style bucket200 -> (false, nil)", http.StatusOK, false},
		{"empty object404 + path-style bucket404 -> error (missing bucket)", http.StatusNotFound, true},
		{"empty object404 + path-style bucket403 -> error (denied bucket)", http.StatusForbidden, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mu.Lock()
			wire = nil
			unexpected = nil
			bucketCode = tc.bucketCode
			mu.Unlock()

			exists, err := u.Exists(context.Background(), key)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.False(t, exists)

			mu.Lock()
			reqs := append([]wireRequest(nil), wire...)
			unexp := append([]string(nil), unexpected...)
			mu.Unlock()
			require.Empty(t, unexp, "server saw requests outside the path-style bucket route")
			require.Len(t, reqs, 2, "ambiguous object 404 must issue HeadObject then one HeadBucket probe")
			assert.Equal(t, "/"+bucket+"/"+key, reqs[0].path)
			assert.Equal(t, "/"+bucket, strings.TrimSuffix(reqs[1].path, "/"))
			for _, r := range reqs {
				assert.Equal(t, http.MethodHead, r.method, "Exists must issue HEAD requests")
				assert.Equal(t, endpointHost, r.host, "requests must reach the configured endpoint host")
			}
			assert.NotEqual(t, reqs[0].path, reqs[1].path,
				"object and bucket wire paths must be distinct")
		})
	}
}
