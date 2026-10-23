package operation

import (
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// CommandProjector is the consumer-supplied PURE projection of canonical
// command bytes onto the default Content fields (the exact fingerprinted
// string map). It must be deterministic and side-effect free: no clock, random
// source, database, I/O or captured mutable state. The runner calls it more
// than once (constructor, before claim, immediately before Apply) and treats
// any difference between calls as a mismatch.
//
// It receives a private copy of the command bytes on every call, so it can
// neither observe nor alter the value's own copy.
type CommandProjector func(schema string, command []byte) (map[string]string, error)

// VerifiedCommand is the operation-package-owned optional immutable value that
// carries the typed command of one request (the 採択済みQ contract in
// docs/design/shared/operation-participation.md). It holds a schema identifier
// and the canonical command bytes, and proves at construction that the
// consumer's projector maps those bytes to exactly the request's default
// Content.
//
// Q is request-memory only. It is never stored in bindings, receipts or logs and
// is not fingerprinted separately: Content stays the only fingerprint input, and
// CAS/fence values stay in Preconditions. The zero value means "no verified
// command" and leaves a Content-only consumer exactly as it was.
//
// Fields are unexported and the getter returns detached copies, so a value
// cannot be altered after construction and request copies may share it safely.
type VerifiedCommand struct {
	schema    string
	command   []byte
	projector CommandProjector
}

// NewVerifiedCommand copies schema and command and verifies that project maps
// the command bytes to a Content whose key set and values equal content
// exactly (no extra or missing key). A blank schema, a nil projector, a
// projector error, or any difference is rejected. The projector is retained
// for the runner's later re-verification.
func NewVerifiedCommand(schema string, command []byte, content Content, project CommandProjector) (VerifiedCommand, error) {
	if strings.TrimSpace(schema) == "" {
		return VerifiedCommand{}, apperrors.WrapInvalidInput("operation verified command schema is required")
	}
	if project == nil {
		return VerifiedCommand{}, apperrors.WrapInvalidInput("operation verified command projector is required")
	}
	copied := make([]byte, len(command))
	copy(copied, command)
	vc := VerifiedCommand{schema: schema, command: copied, projector: project}
	if err := vc.verifyAgainst(content); err != nil {
		return VerifiedCommand{}, err
	}
	return vc, nil
}

// IsZero reports whether no verified command is present.
func (v VerifiedCommand) IsZero() bool {
	return v.schema == "" && v.command == nil && v.projector == nil
}

// Schema returns the command schema identifier.
func (v VerifiedCommand) Schema() string { return v.schema }

// Command returns a detached copy of the canonical command bytes. Apply decodes
// its typed input from these bytes and nowhere else.
func (v VerifiedCommand) Command() []byte {
	out := make([]byte, len(v.command))
	copy(out, v.command)
	return out
}

// verifyAgainst re-runs the projector on the held bytes and requires the result
// to equal content exactly. It never mutates v.
func (v VerifiedCommand) verifyAgainst(content Content) error {
	if strings.TrimSpace(v.schema) == "" || v.projector == nil {
		return apperrors.WrapInvalidInput("operation verified command must be built with NewVerifiedCommand")
	}
	projected, err := v.projector(v.schema, v.Command())
	if err != nil {
		return apperrors.WrapInvalidInput("operation verified command could not be projected")
	}
	projectedContent, err := NewContent(projected)
	if err != nil {
		return apperrors.WrapInvalidInput("operation verified command projection is not a valid content")
	}
	if !projectedContent.Equal(content) {
		return apperrors.WrapInvalidInput("operation verified command does not match the request content")
	}
	return nil
}
