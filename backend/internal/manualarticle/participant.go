package manualarticle

// participant.go — C06 Participant for the manual.article.edit operation.
//
// 採用済み設計（plane-EMR-502 §4 participant）:
//   owner_kind="manualarticle" / binding intent="manual.article.edit" /
//   manual_article_operations.intent="edit"。EffectsDatabaseOnly・Gate は no-op。
//   Apply: 現在の manual-edit:edit grant 再検査 → anchor（claim 時に確定した
//   不変の元対象）解決 → 対象ロックで存在・CAS 検証 → 内容版追加・current
//   pointer・履歴・必須監査・receipt を同一 owner transaction で保存。
//   anchor 由来の対象と経路 selector の一致は lock 下で検査し、一致しない
//   認可済み要求は target_mismatch（業務効果なし）。
//   anchor が読めない・対象が存在しない場合は別 article/URL/body 値へ
//   fallback せず失敗とする。

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/operation"
)

const (
	// OwnerKind は operation_bindings.owner_kind（このドメインの写 owner）。
	OwnerKind = "manualarticle"
	// IntentEdit は operation_bindings.intent（操作 family）。
	IntentEdit = "manual.article.edit"
	// historyIntentEdit は manual_article_operations.intent の公開 receipt intent。
	historyIntentEdit = "edit"
	// lookupTargetKindArticle は LookupTarget.Kind の記事対象。
	lookupTargetKindArticle = "manual_article"

	// 公開 receipt intent（ManualEditReceipt.intent の enum）。
	publicReceiptIntentEdit = "edit"
)

// EditTxStore は participant が必要とする DB 面。各メソッドは ctx の ambient
// transaction（persistence.DBOrTx）で実行され、owner write・履歴・receipt と
// 同一 transaction で原子性を持つ。
type EditTxStore interface {
	// InsertOperationTarget は claim transaction 内で anchor 行を1つ insert する。
	InsertOperationTarget(ctx context.Context, clinicID, bindingID, articleID uint64) error
	// ResolveOperationTarget は binding の保存済み anchor を返す。
	// owner_kind/intent の不一致行は見えない（見つからない扱い）。
	ResolveOperationTarget(ctx context.Context, clinicID, bindingID uint64) (articleID uint64, found bool, err error)
	// FindArticleByID は記事の存在確認（lock なし）。不存在は (nil, nil)。
	FindArticleByID(ctx context.Context, id uint64) (*model.ManualArticle, error)
	// FindVersionByID は article 配下の内容版を読む。不存在は (nil, nil)。
	FindVersionByID(ctx context.Context, articleID, versionID uint64) (*model.ManualArticleVersion, error)
	// LockArticleForUpdate は対象行を FOR UPDATE でロックして読む。
	LockArticleForUpdate(ctx context.Context, id uint64) (*model.ManualArticle, error)
	// NextContentRevision は article の次の content_revision（max+1）を返す。
	// 記事行 lock 保有下で呼ばれるため直列化される。
	NextContentRevision(ctx context.Context, articleID uint64) (int64, error)
	// InsertVersion は不変の新内容版を insert する。
	InsertVersion(ctx context.Context, v *model.ManualArticleVersion) error
	// UpdateCurrentVersionCAS は CAS update:
	// WHERE id AND version=expectedVersion AND current_version_id=expectedCurrentVersionID。
	// 1 行更新で true、0 行で false。version 自体は storage trigger が進める。
	// expectedVersion は記事 CAS 版（int64、model.ManualArticle.Version と同型）。
	UpdateCurrentVersionCAS(ctx context.Context, articleID uint64, expectedVersion int64, expectedCurrentVersionID, newCurrentVersionID uint64) (bool, error)
	// InsertOperation は manual_article_operations 履歴行を insert する。
	InsertOperation(ctx context.Context, op *model.ManualArticleOperation) error
}

// EditGrantChecker は bound actor の現在の manual-edit:edit grant を評価する。
// (true,nil)=許可、(false,nil)=明示的 denial、(false,err)=評価不能で
// いずれも fail-closed。
type EditGrantChecker func(ctx context.Context, clinicID, staffID uint64) (bool, error)

// EditTxAudit は required audit entry を ambient tx で書く fail-closed 契約。
type EditTxAudit interface {
	LogEntryTx(ctx context.Context, entry AuditEntry) error
}

// EditParticipant は manual.article.edit の C06 participant。
// インスタンスは request ごとに作る（rejectCode・resolved selector を保持し
// goroutine 共有しない）。
type EditParticipant struct {
	store       EditTxStore
	checkGrant  EditGrantChecker
	txAudit     EditTxAudit
	clinicID    uint64
	staffID     uint64
	systemAdmin bool
	// route selectors — claim/apply の比較対象（route-decoded raw）。
	routeCategory model.ManualCategory
	routeSlug     string
	// cmd は strict parse 済み編集命令（PUT only。lookup は nil）。
	cmd *editCommand
	// rejectCode は最後の validation 拒否の ManualEditConflict code。
	rejectCode string
	// 解決済み anchor の selector（AuthorizeLookup で記録）。
	resolvedCategory model.ManualCategory
	resolvedSlug     string
	resolvedOK       bool
}

// NewEditParticipant は request-scope の participant を組み立てる。
// lookup 呼び出しでは cmd=nil を渡す。
func NewEditParticipant(
	store EditTxStore,
	check EditGrantChecker,
	txAudit EditTxAudit,
	clinicID, staffID uint64,
	systemAdmin bool,
	routeCategory model.ManualCategory,
	routeSlug string,
	cmd *editCommand,
) *EditParticipant {
	return &EditParticipant{
		store:         store,
		checkGrant:    check,
		txAudit:       txAudit,
		clinicID:      clinicID,
		staffID:       staffID,
		systemAdmin:   systemAdmin,
		routeCategory: routeCategory,
		routeSlug:     routeSlug,
		cmd:           cmd,
	}
}

var (
	_ operation.Participant          = (*EditParticipant)(nil)
	_ operation.ClaimMetadataWriter  = (*EditParticipant)(nil)
)

// Effects — manualarticle の変更は DB transaction のみ（副作用なし）。
func (p *EditParticipant) Effects() operation.EffectScope {
	return operation.EffectsDatabaseOnly
}

// Gate は no-op（precondition fence は Apply lock 下で再検証する）。
func (p *EditParticipant) Gate(_ context.Context) error { return nil }

// ConflictCode は validation 拒否の ManualEditConflict code を返す。
// 拒否が無い・別 status のときは空文字。
func (p *EditParticipant) ConflictCode() string { return p.rejectCode }

// ResolvedSelectors は認可済み anchor の記事 selector を返す。
// anchor 未解決・対象不存在では ok=false。
func (p *EditParticipant) ResolvedSelectors() (category model.ManualCategory, slug string, ok bool) {
	return p.resolvedCategory, p.resolvedSlug, p.resolvedOK
}

// CreateClaimMetadata は claim winner が claim tx 内で anchor 行を永続化する。
// req.Scope/req.Content だけを使い、route/body 値へ fallback しない。
func (p *EditParticipant) CreateClaimMetadata(ctx context.Context, b operation.Binding, req operation.Request) error {
	if req.Scope.OwnerKind != OwnerKind || req.Scope.Intent != IntentEdit {
		return apperrors.WrapInternalServerError("claim metadata scope mismatch")
	}
	articleID, err := strconv.ParseUint(req.Content.Fields()["article_id"], 10, 63)
	if err != nil || articleID == 0 {
		return apperrors.WrapInternalServerError("claim metadata article identity unavailable")
	}
	return p.store.InsertOperationTarget(ctx, b.ClinicID, b.ID, articleID)
}

// ResolveLookupTarget は保存済み anchor から不変の元対象を解決する。
// anchor 欠落・binding 不一致は失敗（別対象への fallback はしない）。
func (p *EditParticipant) ResolveLookupTarget(ctx context.Context, b operation.Binding, _ *operation.Result) (operation.LookupTarget, error) {
	articleID, found, err := p.store.ResolveOperationTarget(ctx, b.ClinicID, b.ID)
	if err != nil {
		return operation.LookupTarget{}, err
	}
	if !found {
		return operation.LookupTarget{}, apperrors.WrapInternalServerError("operation target anchor missing")
	}
	return operation.LookupTarget{
		BindingID: b.ID,
		ClinicID:  b.ClinicID,
		Kind:      lookupTargetKindArticle,
		Ref:       strconv.FormatUint(articleID, 10),
	}, nil
}

// AuthorizeLookup は anchor 解決後に対象の存在と現在 grant を再検査する。
// 認可成功後にだけ selector を記録する（handler が route selector と比較する）。
func (p *EditParticipant) AuthorizeLookup(ctx context.Context, target operation.LookupTarget) error {
	if target.Kind != lookupTargetKindArticle {
		return errors.New("manual article lookup target kind mismatch")
	}
	articleID, err := strconv.ParseUint(target.Ref, 10, 63)
	if err != nil || articleID == 0 {
		return errors.New("manual article lookup target invalid")
	}
	article, err := p.store.FindArticleByID(ctx, articleID)
	if err != nil {
		return err
	}
	if article == nil {
		return errors.New("manual article lookup target missing")
	}
	allowed, err := p.currentEditGrant(ctx)
	if err != nil {
		return err
	}
	if !allowed {
		return errors.New("manual article lookup not authorized")
	}
	p.resolvedCategory = article.Category
	p.resolvedSlug = article.Slug
	p.resolvedOK = true
	return nil
}

// Apply は edit の owner write。全検証は anchor + lock + 現在 grant だけに依存し、
// 再計算した両 digest と fingerprint 済み Content field を照合する。
func (p *EditParticipant) Apply(ctx context.Context, b operation.Binding, req operation.Request) (operation.Result, error) {
	if p.cmd == nil {
		return operation.Result{}, apperrors.WrapInternalServerError("manual edit command unavailable")
	}
	if req.Scope.OwnerKind != OwnerKind || req.Scope.Intent != IntentEdit {
		return operation.Result{}, apperrors.WrapInternalServerError("manual edit scope mismatch")
	}
	// 現在 grant 再検査（失効は明示的拒否 = 403）。
	allowed, err := p.currentEditGrant(ctx)
	if err != nil {
		return operation.Result{}, err
	}
	if !allowed {
		return operation.Result{}, &operation.Rejection{Status: operation.StatusEditRightLost}
	}
	// anchor から不変の元対象を解決する。anchor 欠落は fail-closed（
	// 別 article や request 値への代替はしない）。
	anchorArticleID, found, err := p.store.ResolveOperationTarget(ctx, b.ClinicID, b.ID)
	if err != nil {
		return operation.Result{}, err
	}
	if !found {
		return operation.Result{}, apperrors.WrapInternalServerError("operation target anchor missing")
	}
	fields := req.Content.Fields()
	contentArticleID, err := strconv.ParseUint(fields["article_id"], 10, 63)
	if err != nil || contentArticleID == 0 || contentArticleID != anchorArticleID {
		return operation.Result{}, p.rejectTargetMismatch()
	}
	// 対象をロックして存在を検証する。
	locked, err := p.store.LockArticleForUpdate(ctx, anchorArticleID)
	if err != nil {
		return operation.Result{}, err
	}
	// 認可済み anchor の対象 identity と経路 selector / command の一致。
	if locked.ID != anchorArticleID ||
		string(locked.Category) != string(p.routeCategory) || locked.Slug != p.routeSlug {
		return operation.Result{}, p.rejectTargetMismatch()
	}
	// CAS 値は precondition（fingerprint 対象外の実行前提）から取り、
	// command の値と厳格一致を確認する。不一致は fail-closed。
	preExpectedVersion, err := strconv.ParseInt(req.Preconditions["expected_version"], 10, 64)
	if err != nil || preExpectedVersion != p.cmd.ExpectedVersion {
		return operation.Result{}, apperrors.WrapInternalServerError("manual edit precondition mismatch")
	}
	preExpectedContentID, err := strconv.ParseUint(req.Preconditions["expected_content_version_id"], 10, 63)
	if err != nil || preExpectedContentID != p.cmd.ExpectedContentVersionID {
		return operation.Result{}, apperrors.WrapInternalServerError("manual edit precondition mismatch")
	}
	// CAS fence: CAS 版と期待元内容版を lock 下で再検証する（値は req.Preconditions）。
	if locked.Version != preExpectedVersion ||
		locked.CurrentVersionID == nil || *locked.CurrentVersionID != preExpectedContentID {
		return operation.Result{}, &operation.Rejection{Status: operation.StatusVersionConflict}
	}
	// 元内容の出所解決: current=NULL / legacy_unknown / digest・revision NULL は
	// 架空値で埋めず編集を止める。
	current, err := p.currentVersionRow(ctx, locked)
	if err != nil {
		return operation.Result{}, err
	}
	if current == nil || !isResolvedSource(current) {
		return operation.Result{}, p.rejectSourceUnresolved()
	}
	// 両 digest を再計算し fingerprint 済み Content と照合する。
	// 不一致は業務効果なしで失敗（未実行保証）。
	recomputedContent := manualContentDigestHex(p.cmd.Title, p.cmd.OrderText, p.cmd.Section, p.cmd.BodyMarkdown)
	recomputedCommand := manualEditCommandDigestHex(p.cmd.ArticleIDText, string(locked.Category), locked.Slug, recomputedContent)
	if fields["content_digest"] != recomputedContent ||
		fields["command_digest"] != recomputedCommand ||
		fields["article_id"] != p.cmd.ArticleIDText {
		return operation.Result{}, apperrors.WrapInternalServerError("manual edit content commitment mismatch")
	}
	// 新内容版（revision max+1、source_kind=edited、digest 付き）を作る。
	revision, err := p.store.NextContentRevision(ctx, locked.ID)
	if err != nil {
		return operation.Result{}, err
	}
	digestBytes := sha256Bytes(p.cmd.ContentDigestHex)
	sourceKind := "edited"
	newVersion := &model.ManualArticleVersion{
		ArticleID:       locked.ID,
		Title:           p.cmd.Title,
		OrderValue:      p.cmd.OrderValue,
		Section:         p.cmd.Section,
		BodyMarkdown:    p.cmd.BodyMarkdown,
		EditedByStaffID: &p.staffID,
		ContentDigest:   &digestBytes,
		ContentRevision: &revision,
		SourceKind:      &sourceKind,
	}
	if err := p.store.InsertVersion(ctx, newVersion); err != nil {
		return operation.Result{}, err
	}
	// CAS update で current pointer だけを差し替える（published は不変）。
	updated, err := p.store.UpdateCurrentVersionCAS(
		ctx, locked.ID, preExpectedVersion, preExpectedContentID, newVersion.ID)
	if err != nil {
		return operation.Result{}, err
	}
	if !updated {
		return operation.Result{}, &operation.Rejection{Status: operation.StatusVersionConflict}
	}
	resultingVersion := locked.Version + 1
	// 履歴行: intent=edit、previous/resulting、public pointer 不変、before/after state。
	history := &model.ManualArticleOperation{
		ClinicID:           b.ClinicID,
		ArticleID:          locked.ID,
		OperationID:        b.ID,
		Intent:             historyIntentEdit,
		PreviousVersion:    locked.Version,
		ResultingVersion:   resultingVersion,
		ContentVersionID:   &newVersion.ID,
		PublishedVersionID: locked.PublishedVersionID,
		BeforeState:        json.RawMessage(articleStateJSON(locked.Version, locked.CurrentVersionID, locked.PublishedVersionID)),
		AfterState:         json.RawMessage(articleStateJSON(resultingVersion, &newVersion.ID, locked.PublishedVersionID)),
		RecordedBy:         p.staffID,
	}
	if err := p.store.InsertOperation(ctx, history); err != nil {
		return operation.Result{}, err
	}
	// 必須監査（fail-closed: 失敗は全体 rollback）。
	if err := p.txAudit.LogEntryTx(ctx, AuditEntry{
		ClinicID:   &b.ClinicID,
		ActorID:    &p.staffID,
		ActorType:  string(operation.ActorStaff),
		Action:     model.AuditActionManualArticleUpsert,
		Resource:   "manual_article",
		ResourceID: &locked.ID,
		OldValue:   json.RawMessage(history.BeforeState),
		NewValue:   json.RawMessage(history.AfterState),
	}); err != nil {
		return operation.Result{}, err
	}
	resultRefs := map[string]string{
		"article_id":            strconv.FormatUint(locked.ID, 10),
		"operation_history_id":  strconv.FormatUint(history.ID, 10),
		"content_version_id":    strconv.FormatUint(newVersion.ID, 10),
		"content_digest":        p.cmd.ContentDigestHex,
	}
	if locked.PublishedVersionID != nil {
		resultRefs["published_version_id"] = strconv.FormatUint(*locked.PublishedVersionID, 10)
	}
	return operation.Result{
		ResultRefs: resultRefs,
		CommittedVersions: map[string]string{
			"previous_version":  strconv.FormatInt(locked.Version, 10),
			"resulting_version": strconv.FormatInt(resultingVersion, 10),
			"content_revision":  strconv.FormatInt(revision, 10),
		},
		EffectSummary: map[string]string{
			"article_id": strconv.FormatUint(locked.ID, 10),
			"intent":     historyIntentEdit,
		},
	}, nil
}

func (p *EditParticipant) currentEditGrant(ctx context.Context) (bool, error) {
	if p.systemAdmin {
		return true, nil
	}
	if p.checkGrant == nil {
		return false, errors.New("manual edit grant checker unavailable")
	}
	return p.checkGrant(ctx, p.clinicID, p.staffID)
}

func (p *EditParticipant) currentVersionRow(ctx context.Context, locked *model.ManualArticle) (*model.ManualArticleVersion, error) {
	if locked.CurrentVersionID == nil {
		return nil, nil
	}
	row, err := p.store.FindVersionByID(ctx, locked.ID, *locked.CurrentVersionID)
	if err != nil {
		return nil, err
	}
	return row, nil
}

func (p *EditParticipant) rejectTargetMismatch() error {
	p.rejectCode = "target_mismatch"
	return &operation.Rejection{Status: operation.StatusValidation}
}

func (p *EditParticipant) rejectSourceUnresolved() error {
	p.rejectCode = "source_unresolved"
	return &operation.Rejection{Status: operation.StatusValidation}
}

// isResolvedSource は元内容版が「既知の出所」であるかを返す。
// source_kind NULL / legacy_unknown / digest・revision NULL は未解決出所。
func isResolvedSource(v *model.ManualArticleVersion) bool {
	if v.SourceKind == nil || *v.SourceKind == "legacy_unknown" {
		return false
	}
	return v.ContentDigest != nil && v.ContentRevision != nil
}

func sha256Bytes(hexDigest string) []byte {
	// hexDigest は strict parse 時に算出した 64 lower-hex。decode は失敗しない。
	b, _ := hex.DecodeString(hexDigest)
	return b
}

func articleStateJSON(version int64, currentVersionID, publishedVersionID *uint64) []byte {
	current := json.RawMessage("null")
	if currentVersionID != nil {
		current = json.RawMessage(strconv.FormatUint(*currentVersionID, 10))
	}
	published := json.RawMessage("null")
	if publishedVersionID != nil {
		published = json.RawMessage(strconv.FormatUint(*publishedVersionID, 10))
	}
	raw, _ := json.Marshal(map[string]json.RawMessage{
		"version":              json.RawMessage(strconv.FormatInt(version, 10)),
		"current_version_id":   current,
		"published_version_id": published,
	})
	return raw
}
