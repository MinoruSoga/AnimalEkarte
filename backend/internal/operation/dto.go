package operation

import (
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ParseOperationKey accepts the canonical lowercase hyphenated UUID form only.
// Uppercase, braced or unprefixed forms are rejected instead of normalised, so
// one operation is never addressed by two spellings. The nil UUID is rejected.
func ParseOperationKey(raw string) (uuid.UUID, error) {
	key, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperrors.WrapInvalidInput("operation key is not a UUID")
	}
	if key.String() != raw {
		return uuid.Nil, apperrors.WrapInvalidInput("operation key must be lowercase hyphenated UUID form")
	}
	if key == uuid.Nil {
		return uuid.Nil, apperrors.WrapInvalidInput("operation key is required")
	}
	return key, nil
}

// ResultResponse is the wire shape of one operation outcome. EffectSummary is
// internal and never sent.
type ResultResponse struct {
	Status            Status            `json:"status"`
	OperationKey      string            `json:"operation_key"`
	ResultRefs        map[string]string `json:"result_refs,omitempty"`
	CommittedVersions map[string]string `json:"committed_versions,omitempty"`
	RecordedAt        *time.Time        `json:"recorded_at,omitempty"`
}

// NewResultResponse renders o for the wire. The stored result is included only
// for succeeded and replayed outcomes; every other status carries no body.
func NewResultResponse(key uuid.UUID, o Outcome) ResultResponse {
	operationKey := key
	if operationKey == uuid.Nil && o.Binding != nil {
		operationKey = o.Binding.Key
	}
	response := ResultResponse{
		Status:       o.Status,
		OperationKey: operationKey.String(),
	}
	if o.Status != StatusSucceeded && o.Status != StatusReplayed {
		return response
	}
	if o.Result == nil {
		return response
	}
	response.ResultRefs = copyStringMap(o.Result.ResultRefs)
	response.CommittedVersions = copyStringMap(o.Result.CommittedVersions)
	if !o.Result.RecordedAt.IsZero() {
		recordedAt := o.Result.RecordedAt
		response.RecordedAt = &recordedAt
	}
	return response
}

// copyStringMap returns a detached copy so a response cannot alias stored data.
func copyStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	copied := make(map[string]string, len(values))
	for name, value := range values {
		copied[name] = value
	}
	return copied
}
