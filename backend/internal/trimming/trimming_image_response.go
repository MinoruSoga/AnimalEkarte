package trimming

import (
	"errors"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	trimmingImageSignedURLTTL     = 15 * time.Minute
	trimmingImageStorageKeyPrefix = "trimming-images/"
	trimmingImageLocalURLPrefix   = "/uploads/"
)

var errTrimmingImageStorageKeyUnresolved = errors.New("unable to resolve trimming image storage key")

// trimmingImageUploadResponse は POST /v1/trimmings/images/upload の応答。
// Key が永続化すべきオブジェクトキー（style_image/completed_image に保存する値）、
// SignedURL が表示専用の署名付き URL。署名 URL を DB に書き戻してはならない
// （INV-IMG-FIELD-ASYM: 保存は key、署名は応答時のみ）。
type trimmingImageUploadResponse struct {
	Key       string `json:"key"`
	SignedURL string `json:"signed_url"`
}

func trimmingImageHasHTTPScheme(stored string) bool {
	lower := strings.ToLower(stored)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func trimmingImageObjectKeyHasName(key string) bool {
	if !strings.HasPrefix(key, trimmingImageStorageKeyPrefix) {
		return key != ""
	}
	return strings.TrimPrefix(key, trimmingImageStorageKeyPrefix) != ""
}

// resolveTrimmingImageStorageKey extracts a storage object key from a stored image value.
// isStorageObject=false (and err=nil) means the value is not a storage object and should be echoed.
// err!=nil is fail-closed: the stored value looks like ours but a key cannot be extracted.
func resolveTrimmingImageStorageKey(stored string) (key string, isStorageObject bool, err error) {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return "", false, nil
	}

	// http(s) values never resolve to an object key: no configured origin is
	// available to verify the host against, so host-agnostic path extraction
	// would let a foreign URL alias any same-named internal object (and let
	// cleanup delete it). Echoed unsigned on read, rejected on write —
	// only canonical keys and the local /uploads/ wrap are trusted references.
	if trimmingImageHasHTTPScheme(stored) {
		return "", false, nil
	}

	if strings.HasPrefix(stored, trimmingImageStorageKeyPrefix) {
		if !trimmingImageObjectKeyHasName(stored) {
			return "", false, errTrimmingImageStorageKeyUnresolved
		}
		return stored, true, nil
	}

	if strings.HasPrefix(stored, trimmingImageLocalURLPrefix) {
		rest := strings.TrimPrefix(stored, trimmingImageLocalURLPrefix)
		// /uploads/-wrapped values may carry a query or fragment (e.g. an old
		// signed URL persisted relative); strip both before key extraction so
		// they never become part of the resolved key.
		rest, _, _ = strings.Cut(rest, "?")
		rest, _, _ = strings.Cut(rest, "#")
		if rest == "" {
			return "", false, errTrimmingImageStorageKeyUnresolved
		}
		if !strings.HasPrefix(rest, trimmingImageStorageKeyPrefix) {
			return "", false, nil
		}
		if !trimmingImageObjectKeyHasName(rest) {
			return "", false, errTrimmingImageStorageKeyUnresolved
		}
		return rest, true, nil
	}

	return "", false, nil
}

// trimmingImageKeyBelongsToClinic reports whether key has the exact shape
// trimming-images/<clinic_id>/<name> for the given clinic. Foreign clinics,
// traversal segments, filesystem aliases, and malformed shapes are fail-closed.
// Nested names are allowed, but every accepted key must already be canonical:
// local storage cleans paths, so ./, repeated /, and trailing / must not acquire
// independent reference locks or deletion tombstones for the same local file.
func trimmingImageKeyBelongsToClinic(key string, clinicID uint64) bool {
	if path.Clean(key) != key || strings.Contains(key, "\\") {
		return false
	}
	rest, ok := strings.CutPrefix(key, trimmingImageStorageKeyPrefix)
	if !ok {
		return false
	}
	idSeg, name, ok := strings.Cut(rest, "/")
	if !ok || name == "" || strings.Contains(name, "..") {
		return false
	}
	return idSeg == strconv.FormatUint(clinicID, 10)
}

// Previously accepted local path aliases may remain in stored records. Count
// them conservatively during reference checks, but never rewrite a storage
// deletion target: on R2 an alias can be a distinct object. New writes/deletes
// reject aliases through trimmingImageKeyBelongsToClinic.
func trimmingImageStoredReferenceMatchesKey(stored, key string) bool {
	resolved, isObject, err := resolveTrimmingImageStorageKey(stored)
	return err == nil && isObject && (resolved == key || path.Clean(resolved) == key)
}
