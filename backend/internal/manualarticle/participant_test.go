package manualarticle

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/operation"
)

// ---- fake EditTxStore ----
//
// 全部の Write 呼出を計数する。Apply の「業務効果なし拒否」は writes==0 で証明する。

type fakeEditTxStore struct {
	anchorID    uint64
	anchorFound bool
	articleByID *model.ManualArticle
	locked      *model.ManualArticle
	versionByID map[uint64]*model.ManualArticleVersion
	nextRev     int64
	casResult   bool

	insertTargetErr error
	resolveErr      error
	articleErr      error
	versionErr      error
	lockErr         error
	revisionErr     error
	insertVErr      error
	casErr          error
	insertOpErr     error

	insertedVersion *model.ManualArticleVersion
	insertedOp      *model.ManualArticleOperation
	insertedTargets []uint64
	writes          int
}

func (f *fakeEditTxStore) InsertOperationTarget(_ context.Context, _, _, articleID uint64) error {
	if f.insertTargetErr != nil {
		return f.insertTargetErr
	}
	f.insertedTargets = append(f.insertedTargets, articleID)
	return nil
}

func (f *fakeEditTxStore) ResolveOperationTarget(_ context.Context, _, _ uint64) (uint64, bool, error) {
	if f.resolveErr != nil {
		return 0, false, f.resolveErr
	}
	return f.anchorID, f.anchorFound, nil
}

func (f *fakeEditTxStore) FindArticleByID(_ context.Context, _ uint64) (*model.ManualArticle, error) {
	return f.articleByID, f.articleErr
}

func (f *fakeEditTxStore) FindVersionByID(_ context.Context, _, versionID uint64) (*model.ManualArticleVersion, error) {
	if f.versionErr != nil {
		return nil, f.versionErr
	}
	return f.versionByID[versionID], nil
}

func (f *fakeEditTxStore) LockArticleForUpdate(_ context.Context, _ uint64) (*model.ManualArticle, error) {
	if f.lockErr != nil {
		return nil, f.lockErr
	}
	return f.locked, nil
}

func (f *fakeEditTxStore) NextContentRevision(_ context.Context, _ uint64) (int64, error) {
	if f.revisionErr != nil {
		return 0, f.revisionErr
	}
	return f.nextRev, nil
}

func (f *fakeEditTxStore) InsertVersion(_ context.Context, v *model.ManualArticleVersion) error {
	f.writes++
	if f.insertVErr != nil {
		return f.insertVErr
	}
	f.insertedVersion = v
	if v.ID == 0 {
		v.ID = 1001
	}
	return nil
}

func (f *fakeEditTxStore) UpdateCurrentVersionCAS(_ context.Context, _ uint64, _ int64, _, _ uint64) (bool, error) {
	f.writes++
	if f.casErr != nil {
		return false, f.casErr
	}
	return f.casResult, nil
}

func (f *fakeEditTxStore) InsertOperation(_ context.Context, op *model.ManualArticleOperation) error {
	f.writes++
	if f.insertOpErr != nil {
		return f.insertOpErr
	}
	f.insertedOp = op
	if op.ID == 0 {
		op.ID = 77
	}
	return nil
}

// ---- helpers ----

func newCmdForTest(t *testing.T) *editCommand {
	t.Helper()
	orderText := "1.5"
	contentDigest := manualContentDigestHex("タイトル", orderText, "基本", "本文")
	commandDigest := manualEditCommandDigestHex("42", string(model.ManualCategoryScreens), "intro", contentDigest)
	var num pgtype.Numeric
	require.NoError(t, num.Scan(orderText))
	return &editCommand{
		OperationKey:             uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		ArticleID:                42,
		ArticleIDText:            "42",
		ExpectedVersion:          3,
		ExpectedContentVersionID: 7,
		Title:                    "タイトル",
		OrderText:                orderText,
		OrderValue:               model.ExactNumeric{Numeric: num},
		Section:                  "基本",
		BodyMarkdown:             "本文",
		ContentDigestHex:         contentDigest,
		CommandDigestHex:         commandDigest,
	}
}

func reqForTest(t *testing.T, cmd *editCommand) operation.Request {
	t.Helper()
	content, err := cmd.operationContent()
	require.NoError(t, err)
	return operation.Request{
		Scope: operation.Scope{
			ClinicID:  1,
			OwnerKind: OwnerKind,
			ActorKind: operation.ActorStaff,
			ActorRef:  "7",
			Intent:    IntentEdit,
		},
		Key:           cmd.OperationKey,
		Content:       content,
		Preconditions: cmd.operationPreconditions(),
	}
}

func bindingForTest(cmd *editCommand) operation.Binding {
	return operation.Binding{ID: 55, ClinicID: 1, Key: cmd.OperationKey}
}

func resolvedSourceVersion(id uint64) *model.ManualArticleVersion {
	digest := []byte{0x01}
	revision := int64(1)
	kind := "edited"
	return &model.ManualArticleVersion{
		ID:              id,
		ArticleID:       42,
		ContentDigest:   &digest,
		ContentRevision: &revision,
		SourceKind:      &kind,
	}
}

func lockedArticle(currentID *uint64) *model.ManualArticle {
	publishedID := uint64(9)
	return &model.ManualArticle{
		ID:                 42,
		Category:           model.ManualCategoryScreens,
		Slug:               "intro",
		Version:            3,
		CurrentVersionID:   currentID,
		PublishedVersionID: &publishedID,
	}
}

// happyFixture は全検査を通過する store/grant/audit の組を返す。
func happyFixture() (*fakeEditTxStore, EditGrantChecker, *mockEditTxAudit) {
	currentID := uint64(7)
	store := &fakeEditTxStore{
		anchorID:    42,
		anchorFound: true,
		articleByID: &model.ManualArticle{ID: 42, Category: model.ManualCategoryScreens, Slug: "intro"},
		locked:      lockedArticle(&currentID),
		versionByID: map[uint64]*model.ManualArticleVersion{7: resolvedSourceVersion(7)},
		nextRev:     2,
		casResult:   true,
	}
	grant := func(context.Context, uint64, uint64) (bool, error) { return true, nil }
	return store, grant, &mockEditTxAudit{}
}

func newParticipantForTest(store EditTxStore, grant EditGrantChecker, audit EditTxAudit, cmd *editCommand) *EditParticipant {
	return NewEditParticipant(store, grant, audit, 1, 7, false, model.ManualCategoryScreens, "intro", cmd)
}

func rejectionStatus(err error) (operation.Status, bool) {
	var rej *operation.Rejection
	if errors.As(err, &rej) {
		return rej.Status, true
	}
	return "", false
}

// ---- Apply safety cases ----

func TestEditParticipant_GrantRevoked(t *testing.T) {
	store, _, audit := happyFixture()
	cmd := newCmdForTest(t)
	p := newParticipantForTest(store, func(context.Context, uint64, uint64) (bool, error) { return false, nil }, audit, cmd)
	res, err := p.Apply(context.Background(), bindingForTest(cmd), reqForTest(t, cmd))
	assert.Equal(t, operation.Result{}, res)
	st, ok := rejectionStatus(err)
	require.True(t, ok)
	assert.Equal(t, operation.StatusEditRightLost, st)
	assert.Zero(t, store.writes)
}

func TestEditParticipant_AnchorMissing(t *testing.T) {
	store, grant, audit := happyFixture()
	store.anchorFound = false
	cmd := newCmdForTest(t)
	p := newParticipantForTest(store, grant, audit, cmd)
	res, err := p.Apply(context.Background(), bindingForTest(cmd), reqForTest(t, cmd))
	assert.Equal(t, operation.Result{}, res)
	require.Error(t, err)
	// 別記事・request値への fallback はしない
	assert.Zero(t, store.writes)
}

func TestEditParticipant_AnchorArticleMismatch(t *testing.T) {
	store, grant, audit := happyFixture()
	store.anchorID = 43 // anchor は別記事。Content.article_id=42 とは一致しない。
	cmd := newCmdForTest(t)
	p := newParticipantForTest(store, grant, audit, cmd)
	res, err := p.Apply(context.Background(), bindingForTest(cmd), reqForTest(t, cmd))
	assert.Equal(t, operation.Result{}, res)
	st, ok := rejectionStatus(err)
	require.True(t, ok)
	assert.Equal(t, operation.StatusValidation, st)
	assert.Equal(t, "target_mismatch", p.ConflictCode())
	assert.Zero(t, store.writes)
}

func TestEditParticipant_RouteSelectorMismatch(t *testing.T) {
	store, grant, audit := happyFixture()
	store.locked.Category = model.ManualCategoryWorkflows // 経路 screens とは別 category
	cmd := newCmdForTest(t)
	p := newParticipantForTest(store, grant, audit, cmd)
	res, err := p.Apply(context.Background(), bindingForTest(cmd), reqForTest(t, cmd))
	assert.Equal(t, operation.Result{}, res)
	st, ok := rejectionStatus(err)
	require.True(t, ok)
	assert.Equal(t, operation.StatusValidation, st)
	assert.Equal(t, "target_mismatch", p.ConflictCode())
	assert.Zero(t, store.writes)
}

func TestEditParticipant_CASVersionMismatch(t *testing.T) {
	store, grant, audit := happyFixture()
	store.locked.Version = 4 // expected=3 とは不整合
	cmd := newCmdForTest(t)
	p := newParticipantForTest(store, grant, audit, cmd)
	res, err := p.Apply(context.Background(), bindingForTest(cmd), reqForTest(t, cmd))
	assert.Equal(t, operation.Result{}, res)
	st, ok := rejectionStatus(err)
	require.True(t, ok)
	assert.Equal(t, operation.StatusVersionConflict, st)
	assert.Zero(t, store.writes)
}

func TestEditParticipant_CASContentVersionMismatch(t *testing.T) {
	store, grant, audit := happyFixture()
	otherID := uint64(8)
	store.locked.CurrentVersionID = &otherID // expected_content_version_id=7 とは不整合
	store.versionByID[8] = resolvedSourceVersion(8)
	cmd := newCmdForTest(t)
	p := newParticipantForTest(store, grant, audit, cmd)
	res, err := p.Apply(context.Background(), bindingForTest(cmd), reqForTest(t, cmd))
	assert.Equal(t, operation.Result{}, res)
	st, ok := rejectionStatus(err)
	require.True(t, ok)
	assert.Equal(t, operation.StatusVersionConflict, st)
	assert.Zero(t, store.writes)
}

func TestEditParticipant_CurrentNullIsSourceUnresolved(t *testing.T) {
	store, grant, audit := happyFixture()
	store.locked.CurrentVersionID = nil // current=NULL → CAS より先に source_unresolved
	cmd := newCmdForTest(t)
	p := newParticipantForTest(store, grant, audit, cmd)
	res, err := p.Apply(context.Background(), bindingForTest(cmd), reqForTest(t, cmd))
	assert.Equal(t, operation.Result{}, res)
	st, ok := rejectionStatus(err)
	require.True(t, ok)
	assert.Equal(t, operation.StatusValidation, st)
	assert.Equal(t, "source_unresolved", p.ConflictCode())
	assert.Zero(t, store.writes)
}

func TestEditParticipant_LegacyUnknownIsSourceUnresolved(t *testing.T) {
	store, grant, audit := happyFixture()
	kind := "legacy_unknown"
	digest := []byte{0x01}
	revision := int64(1)
	store.versionByID[7] = &model.ManualArticleVersion{
		ID: 7, ArticleID: 42,
		ContentDigest: &digest, ContentRevision: &revision, SourceKind: &kind,
	}
	cmd := newCmdForTest(t)
	p := newParticipantForTest(store, grant, audit, cmd)
	res, err := p.Apply(context.Background(), bindingForTest(cmd), reqForTest(t, cmd))
	assert.Equal(t, operation.Result{}, res)
	st, ok := rejectionStatus(err)
	require.True(t, ok)
	assert.Equal(t, operation.StatusValidation, st)
	assert.Equal(t, "source_unresolved", p.ConflictCode())
	assert.Zero(t, store.writes)
}

func TestEditParticipant_TamperedContentDigest(t *testing.T) {
	store, grant, audit := happyFixture()
	cmd := newCmdForTest(t)
	req := reqForTest(t, cmd)
	// fingerprint 済み Content を cmd とは別の値で組み立て直す（tamper）。
	tampered, err := operation.NewContent(map[string]string{
		"article_id":     "42",
		"content_digest": "0000000000000000000000000000000000000000000000000000000000000000",
		"command_digest": cmd.CommandDigestHex,
	})
	require.NoError(t, err)
	req.Content = tampered
	p := newParticipantForTest(store, grant, audit, cmd)
	res, err := p.Apply(context.Background(), bindingForTest(cmd), req)
	assert.Equal(t, operation.Result{}, res)
	require.Error(t, err)
	assert.Zero(t, store.writes)
}

func TestEditParticipant_AuditFailurePropagates(t *testing.T) {
	store, grant, audit := happyFixture()
	audit.err = errors.New("audit unavailable")
	cmd := newCmdForTest(t)
	p := newParticipantForTest(store, grant, audit, cmd)
	res, err := p.Apply(context.Background(), bindingForTest(cmd), reqForTest(t, cmd))
	assert.Equal(t, operation.Result{}, res)
	require.Error(t, err)
	// 監査は owner write 後に呼ばれる（rollback は caller tx が行う）。
	assert.Greater(t, store.writes, 0)
}

func TestEditParticipant_Success(t *testing.T) {
	store, grant, audit := happyFixture()
	cmd := newCmdForTest(t)
	p := newParticipantForTest(store, grant, audit, cmd)
	res, err := p.Apply(context.Background(), bindingForTest(cmd), reqForTest(t, cmd))
	require.NoError(t, err)

	require.NotNil(t, store.insertedVersion)
	assert.Equal(t, uint64(42), store.insertedVersion.ArticleID)
	assert.Equal(t, "edited", *store.insertedVersion.SourceKind)
	assert.NotNil(t, store.insertedVersion.ContentDigest)
	require.NotNil(t, store.insertedVersion.ContentRevision)
	assert.Equal(t, int64(2), *store.insertedVersion.ContentRevision)

	require.NotNil(t, store.insertedOp)
	assert.Equal(t, "edit", store.insertedOp.Intent)
	assert.Equal(t, int64(3), store.insertedOp.PreviousVersion)
	assert.Equal(t, int64(4), store.insertedOp.ResultingVersion)
	assert.Equal(t, uint64(1001), *store.insertedOp.ContentVersionID)
	require.NotNil(t, store.insertedOp.PublishedVersionID)
	assert.Equal(t, uint64(9), *store.insertedOp.PublishedVersionID)

	assert.Equal(t, 1, audit.calls)

	refs := res.ResultRefs
	assert.Equal(t, "42", refs["article_id"])
	assert.Equal(t, strconv.FormatUint(1001, 10), refs["content_version_id"])
	assert.Equal(t, "9", refs["published_version_id"])
	assert.Equal(t, strconv.FormatUint(77, 10), refs["operation_history_id"])
	assert.Equal(t, cmd.ContentDigestHex, refs["content_digest"])
	assert.Equal(t, "3", res.CommittedVersions["previous_version"])
	assert.Equal(t, "4", res.CommittedVersions["resulting_version"])
	assert.Equal(t, "2", res.CommittedVersions["content_revision"])
}

// ---- lookup authorization cases ----

func TestEditParticipant_ResolveLookupTarget_UsesAnchorOnly(t *testing.T) {
	store, grant, audit := happyFixture()
	p := newParticipantForTest(store, grant, audit, nil)
	target, err := p.ResolveLookupTarget(context.Background(), operation.Binding{ID: 55, ClinicID: 1}, nil)
	require.NoError(t, err)
	assert.Equal(t, uint64(55), target.BindingID)
	assert.Equal(t, uint64(1), target.ClinicID)
	assert.Equal(t, lookupTargetKindArticle, target.Kind)
	assert.Equal(t, "42", target.Ref)

	// anchor 欠落は fail-closed（別対象・request 値への代替はしない）
	store.anchorFound = false
	_, err = p.ResolveLookupTarget(context.Background(), operation.Binding{ID: 55, ClinicID: 1}, nil)
	require.Error(t, err)
}

func TestEditParticipant_AuthorizeLookup(t *testing.T) {
	store, grant, audit := happyFixture()

	// grant 拒否: selector は記録されない
	p := newParticipantForTest(store, func(context.Context, uint64, uint64) (bool, error) { return false, nil }, audit, nil)
	err := p.AuthorizeLookup(context.Background(), operation.LookupTarget{
		BindingID: 55, ClinicID: 1, Kind: lookupTargetKindArticle, Ref: "42",
	})
	require.Error(t, err)
	_, _, ok := p.ResolvedSelectors()
	assert.False(t, ok)

	// grant 許可 + 対象存在: 認可後に selector が記録される
	p2 := newParticipantForTest(store, grant, audit, nil)
	require.NoError(t, p2.AuthorizeLookup(context.Background(), operation.LookupTarget{
		BindingID: 55, ClinicID: 1, Kind: lookupTargetKindArticle, Ref: "42",
	}))
	cat, slug, ok := p2.ResolvedSelectors()
	require.True(t, ok)
	assert.Equal(t, model.ManualCategoryScreens, cat)
	assert.Equal(t, "intro", slug)
}
