package medicalrecord

import "context"

type masterTestTransactor struct{}

func (masterTestTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type masterTestAudit struct{}

func (masterTestAudit) LogEntryTx(context.Context, *AuditEntry) error { return nil }
func masterTestWriteDependencies() MasterInsuranceWriteDependencies {
	return MasterInsuranceWriteDependencies{Transactor: masterTestTransactor{}, Audit: masterTestAudit{}}
}
func masterTestActorID() *uint64 { actor := uint64(12); return &actor }

func masterTestSortOrder(value int) *int { return &value }
