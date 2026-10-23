package operation

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

const testCommandSchema = "test.price_ack.v1"

// jsonProjector is a pure projector: the command bytes are a JSON object of
// strings and project to themselves.
func jsonProjector(_ string, command []byte) (map[string]string, error) {
	var fields map[string]string
	if err := json.Unmarshal(command, &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

func mustContent(t *testing.T, fields map[string]string) Content {
	t.Helper()
	c, err := NewContent(fields)
	require.NoError(t, err)
	return c
}

func mustCommandBytes(t *testing.T, fields map[string]string) []byte {
	t.Helper()
	b, err := json.Marshal(fields)
	require.NoError(t, err)
	return b
}

func TestNewVerifiedCommand_AcceptsMatchingProjection(t *testing.T) {
	fields := map[string]string{"billing_id": "42", "after_amount": "12000"}
	content := mustContent(t, fields)
	raw := mustCommandBytes(t, fields)

	vc, err := NewVerifiedCommand(testCommandSchema, raw, content, jsonProjector)

	require.NoError(t, err)
	require.False(t, vc.IsZero())
	require.Equal(t, testCommandSchema, vc.Schema())
	require.Equal(t, raw, vc.Command())
}

func TestVerifiedCommand_ZeroValueIsAbsent(t *testing.T) {
	var vc VerifiedCommand

	require.True(t, vc.IsZero())
	require.Empty(t, vc.Schema())
	require.Empty(t, vc.Command())
}

func TestNewVerifiedCommand_DeepCopiesInputAndGettersReturnCopies(t *testing.T) {
	fields := map[string]string{"billing_id": "42"}
	content := mustContent(t, fields)
	raw := mustCommandBytes(t, fields)
	original := append([]byte(nil), raw...)

	vc, err := NewVerifiedCommand(testCommandSchema, raw, content, jsonProjector)
	require.NoError(t, err)

	// Mutating the caller's slice after construction must not change the value.
	for i := range raw {
		raw[i] = 'x'
	}
	require.Equal(t, original, vc.Command())

	// Mutating a returned slice must not change the value either.
	got := vc.Command()
	for i := range got {
		got[i] = 'y'
	}
	require.Equal(t, original, vc.Command())

	// The value still verifies against the content after both mutations.
	require.NoError(t, vc.verifyAgainst(content))
}

func TestNewVerifiedCommand_RejectsProjectionMismatch(t *testing.T) {
	base := map[string]string{"billing_id": "42", "after_amount": "12000"}
	content := mustContent(t, base)

	cases := []struct {
		name    string
		schema  string
		command map[string]string
		project CommandProjector
	}{
		{name: "extra_key", schema: testCommandSchema, command: map[string]string{"billing_id": "42", "after_amount": "12000", "extra": "1"}, project: jsonProjector},
		{name: "missing_key", schema: testCommandSchema, command: map[string]string{"billing_id": "42"}, project: jsonProjector},
		{name: "different_value", schema: testCommandSchema, command: map[string]string{"billing_id": "42", "after_amount": "99999"}, project: jsonProjector},
		{name: "projector_error", schema: testCommandSchema, command: base, project: func(string, []byte) (map[string]string, error) { return nil, errors.New("decode failed") }},
		{name: "nil_projector", schema: testCommandSchema, command: base, project: nil},
		{name: "empty_schema", schema: "", command: base, project: jsonProjector},
		{name: "blank_schema", schema: "  ", command: base, project: jsonProjector},
		{name: "projection_exceeds_content_limits", schema: testCommandSchema, command: base, project: func(string, []byte) (map[string]string, error) {
			return map[string]string{" ": "1"}, nil
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vc, err := NewVerifiedCommand(tc.schema, mustCommandBytes(t, tc.command), content, tc.project)

			require.Error(t, err)
			require.True(t, vc.IsZero())
		})
	}
}

func TestNewVerifiedCommand_ProjectorCannotAlterHeldBytes(t *testing.T) {
	fields := map[string]string{"billing_id": "42"}
	content := mustContent(t, fields)
	raw := mustCommandBytes(t, fields)
	scribbler := func(schema string, command []byte) (map[string]string, error) {
		out, err := jsonProjector(schema, command)
		for i := range command {
			command[i] = 0 // a misbehaving projector scribbles over its argument
		}
		return out, err
	}

	vc, err := NewVerifiedCommand(testCommandSchema, raw, content, scribbler)

	require.NoError(t, err)
	require.Equal(t, mustCommandBytes(t, fields), vc.Command())
	require.NoError(t, vc.verifyAgainst(content))
}

func TestVerifiedCommand_PartialValueFailsClosed(t *testing.T) {
	content := mustContent(t, map[string]string{"a": "1"})
	// Only reachable inside the package; guards against a half-built value.
	partial := VerifiedCommand{schema: testCommandSchema}

	require.False(t, partial.IsZero())
	require.Error(t, partial.verifyAgainst(content))
}
