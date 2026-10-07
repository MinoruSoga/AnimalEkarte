package trimming

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// resolveTrimmingImageStorageKey mirrors the medical-record stored-value
// resolution contract: trim whitespace, recognize own-prefix keys (raw,
// /uploads/-wrapped, or inside an http URL), echo non-object values, and fail
// closed on malformed own-looking values.
func TestResolveTrimmingImageStorageKey(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		stored  string
		wantKey string
		wantObj bool
		wantErr bool
	}{
		{name: "empty", stored: "", wantKey: "", wantObj: false},
		{name: "whitespace only", stored: "   ", wantKey: "", wantObj: false},
		{name: "raw own key", stored: "trimming-images/1/abc.jpg", wantKey: "trimming-images/1/abc.jpg", wantObj: true},
		{name: "key with whitespace", stored: "  trimming-images/1/abc.jpg  ", wantKey: "trimming-images/1/abc.jpg", wantObj: true},
		{name: "bare prefix is malformed", stored: "trimming-images/", wantErr: true},
		{name: "local uploads wrapper", stored: "/uploads/trimming-images/1/abc.jpg", wantKey: "trimming-images/1/abc.jpg", wantObj: true},
		{name: "uploads wrapper empty rest", stored: "/uploads/", wantErr: true},
		{name: "uploads wrapper non-trimming", stored: "/uploads/other/1/abc.jpg", wantObj: false},
		{name: "http url with own path", stored: "https://cdn.example/bucket/trimming-images/1/abc.jpg", wantKey: "trimming-images/1/abc.jpg", wantObj: true},
		{name: "http url without own path echoes", stored: "https://cdn.example/legacy.jpg", wantObj: false},
		{name: "http url malformed", stored: "http://%zz", wantErr: true},
		{name: "data url echoes", stored: "data:image/png;base64,AAAA", wantObj: false},
		{name: "relative legacy path echoes", stored: "images/style.jpg", wantObj: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key, isObject, err := resolveTrimmingImageStorageKey(tc.stored)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantObj, isObject)
			assert.Equal(t, tc.wantKey, key)
		})
	}
}

// trimmingImageKeyBelongsToClinic pins the fail-closed clinic-prefix check:
// only keys of the exact shape trimming-images/<clinic_id>/<name> for the
// record's clinic pass; '..', missing segments, and foreign clinics fail.
func TestTrimmingImageKeyBelongsToClinic(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		key      string
		clinicID uint64
		want     bool
	}{
		{name: "own clinic key", key: "trimming-images/1/abc.jpg", clinicID: 1, want: true},
		{name: "foreign clinic key", key: "trimming-images/2/abc.jpg", clinicID: 1, want: false},
		{name: "missing clinic segment", key: "trimming-images/abc.jpg", clinicID: 1, want: false},
		{name: "empty object name", key: "trimming-images/1/", clinicID: 1, want: false},
		{name: "dotdot traversal", key: "trimming-images/1/../2/x.jpg", clinicID: 1, want: false},
		{name: "non-numeric clinic segment", key: "trimming-images/x/abc.jpg", clinicID: 1, want: false},
		{name: "wrong prefix", key: "medical-records/1/abc.jpg", clinicID: 1, want: false},
		{name: "nested name allowed", key: "trimming-images/1/a/b/c.jpg", clinicID: 1, want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, trimmingImageKeyBelongsToClinic(tc.key, tc.clinicID))
		})
	}
}

func TestNormalizeTrimmingImageForPersist(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		stored  string
		want    string
		wantErr bool
	}{
		{name: "empty clears", stored: "  ", want: ""},
		{name: "own key stays", stored: "trimming-images/1/abc.jpg", want: "trimming-images/1/abc.jpg"},
		{name: "signed url collapses to key", stored: "https://cdn.example/bucket/trimming-images/1/abc.jpg?X-Amz-Signature=secret", want: "trimming-images/1/abc.jpg"},
		{name: "uploads wrapper collapses", stored: "/uploads/trimming-images/1/abc.jpg", want: "trimming-images/1/abc.jpg"},
		{name: "legacy relative passes through", stored: "before.jpg", want: "before.jpg"},
		{name: "data uri passes through", stored: "data:image/png;base64,AAAA", want: "data:image/png;base64,AAAA"},
		{name: "external http rejected", stored: "https://cdn.example/legacy.jpg", wantErr: true},
		{name: "stray uploads rejected", stored: "/uploads/other/1/abc.jpg", wantErr: true},
		{name: "foreign key rejected", stored: "trimming-images/2/abc.jpg", wantErr: true},
		{name: "malformed prefix rejected", stored: "trimming-images/", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeTrimmingImageForPersist(1, tc.stored)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
