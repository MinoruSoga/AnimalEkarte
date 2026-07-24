package service

import "context"

// mockTransactor is the test-only transaction boundary used by the retained
// clinic target-domain regression suite.
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
