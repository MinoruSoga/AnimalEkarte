package trimming

import (
	"errors"
	"net/url"
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

	if trimmingImageHasHTTPScheme(stored) {
		parsed, parseErr := url.Parse(stored)
		if parseErr != nil || parsed.Host == "" {
			return "", false, errTrimmingImageStorageKeyUnresolved
		}
		// Host is intentionally not checked: only the path's
		// trimming-images/<clinic>/... segment is extracted, then re-validated
		// against the caller's clinic and signed by OUR uploader only. A foreign
		// host can therefore reference nothing more than the caller's own keys —
		// equivalent to submitting the key directly.
		extracted, ok := trimmingImageKeyFromURLPath(parsed.Path)
		if !ok {
			return "", false, nil
		}
		if !trimmingImageObjectKeyHasName(extracted) {
			return "", false, errTrimmingImageStorageKeyUnresolved
		}
		return extracted, true, nil
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

func trimmingImageKeyFromURLPath(path string) (string, bool) {
	marker := "/" + trimmingImageStorageKeyPrefix
	idx := strings.Index(path, marker)
	if idx < 0 {
		return "", false
	}
	return path[idx+1:], true
}

// trimmingImageKeyBelongsToClinic reports whether key has the exact shape
// trimming-images/<clinic_id>/<name> for the given clinic. Foreign clinics,
// traversal segments, and malformed shapes are fail-closed. The name segment
// deliberately allows characters like '?' or '/' — object keys are not
// filesystem paths and may legally contain them (the local uploader still
// keeps them inside the clinic directory); '..' is rejected because a
// local-storage backend could traverse out of it.
func trimmingImageKeyBelongsToClinic(key string, clinicID uint64) bool {
	rest, ok := strings.CutPrefix(key, trimmingImageStorageKeyPrefix)
	if !ok {
		return false
	}
	idSeg, name, ok := strings.Cut(rest, "/")
	if !ok || name == "" || strings.Contains(name, "..") {
		return false
	}
	parsed, err := strconv.ParseUint(idSeg, 10, 64)
	return err == nil && parsed == clinicID
}
