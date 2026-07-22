package service

import "context"

// mockTransactor remains for non-trimming legacy service tests. Remove with the service test
// aggregator in BE9-2F after those consumers move to their owning domains.
type mockTransactor struct {
	withTxErr error
	withTxFn  func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}
	if m.withTxErr != nil {
		return m.withTxErr
	}
	return fn(ctx)
}
