package billing

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// AccountingListFilters is the list query contract for FindAll / List.
// StatusOp: "is" (default) | "is_not".
// PaymentMethodOp: "is" (default) | "is_not" | "is_empty" | "is_not_empty".
type AccountingListFilters struct {
	PetID           *uint64
	OwnerID         *uint64
	Status          *string
	StatusOp        string
	StartDate       *string
	EndDate         *string
	Search          string
	PaymentMethod   *string
	PaymentMethodOp string
}

type accountingRepository struct {
	db *gorm.DB
}

func NewAccountingRepository(db *gorm.DB) *accountingRepository {
	return &accountingRepository{db: db}
}

func (r *accountingRepository) FindAll(ctx context.Context, clinicID uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error) {
	return r.findBillingsWithFilters(ctx, []uint64{clinicID}, filters, page, limit)
}

func (r *accountingRepository) FindAllForClinics(ctx context.Context, clinicIDs []uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error) {
	return r.findBillingsWithFilters(ctx, clinicIDs, filters, page, limit)
}

// billingListClinicScope は一覧クエリの clinic スコープ。pets 相関サブクエリと同居するため、
// clinic_id は必ず billings. で修飾する。
func billingListClinicScope(clinicIDs []uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("billings.clinic_id IN ?", clinicIDs)
	}
}

// applyAccountingListFilters は一覧の COUNT / FIND 両クエリに共通するフィルタ述語。
// 破損 FK を除く pets EXISTS 絞り込みは契約上 Find 側のみ（corrupt-FK 分離テストが
// total ≠ len(items) を要求する）なのでここには含めない。
// JOIN 併用のため裸列名は曖昧になり得るので全て billings. で修飾する。
func applyAccountingListFilters(q *gorm.DB, filters AccountingListFilters) *gorm.DB {
	if filters.PetID != nil {
		q = q.Where("billings.pet_id = ?", *filters.PetID)
	}
	if filters.OwnerID != nil {
		q = q.Where("billings.owner_id = ?", *filters.OwnerID)
	}
	if filters.StartDate != nil {
		q = q.Where("billings.scheduled_date >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		q = q.Where("billings.scheduled_date <= ?", *filters.EndDate)
	}
	return q
}

func (r *accountingRepository) billingListQuery(ctx context.Context, clinicIDs []uint64, filters AccountingListFilters) *gorm.DB {
	return applyAccountingListFilters(
		r.db.WithContext(ctx).Model(&model.Billing{}).Scopes(billingListClinicScope(clinicIDs)),
		filters,
	)
}

// rejectUnsupportedListFilters は置換先の reader が未接続の絞り込みを、読取の前に固定エラー（422）で拒否する。
// held for EMR-506: 状態（旧 status）・支払方法は保存先の廃止、飼主・ペット名の検索は氏名 reader（EMR-491）待ち（Q4(c) / Q8）。
// 無視して全件を返すと「絞り込めた」と誤認させるため、黙って外さない。
func rejectUnsupportedListFilters(filters AccountingListFilters) error {
	if filters.Status != nil && *filters.Status != "" {
		return apperrors.WrapFinanceConsumerPending("会計一覧の状態による絞り込みは、請求状態の表示対応が未決定のため利用できません")
	}
	if filters.Search != "" {
		return apperrors.WrapFinanceConsumerPending("会計一覧の飼主・ペット名による検索は、氏名の参照が未接続のため利用できません")
	}
	op := strings.ToLower(strings.TrimSpace(filters.PaymentMethodOp))
	hasMethod := filters.PaymentMethod != nil && strings.TrimSpace(*filters.PaymentMethod) != ""
	if hasMethod || op == "is_empty" || op == "is_not_empty" {
		return apperrors.WrapFinanceConsumerPending("会計一覧の支払方法による絞り込みは、受領の保存先が未接続のため利用できません")
	}
	return nil
}

// findBillingsWithFilters はフィルタ・ページネーション適用後に明細を付与して返す共通実装。
// COUNT と FIND を並行化して往復数を削減する（EMR-201）。
// Q4(c) / Q8: 合計・状態・支払い・返金と、owner / pet の入れ子は読まない（held for EMR-506）。
func (r *accountingRepository) findBillingsWithFilters(ctx context.Context, clinicIDs []uint64, filters AccountingListFilters, page, limit int) ([]model.Billing, int64, error) {
	if err := rejectUnsupportedListFilters(filters); err != nil {
		return nil, 0, err
	}
	billings := make([]model.Billing, 0)
	var total int64

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := r.billingListQuery(gctx, clinicIDs, filters).Count(&total).Error; err != nil {
			return apperrors.FromGORM(err, "billing", "")
		}
		return nil
	})
	g.Go(func() error {
		err := r.billingListQuery(gctx, clinicIDs, filters).
			Where("(billings.pet_id IS NULL OR EXISTS (SELECT 1 FROM pets p WHERE p.id = billings.pet_id AND p.clinic_id = billings.clinic_id))").
			Scopes(persistence.Paginate(page, limit)).
			Order("billings.scheduled_date DESC, billings.created_at DESC").
			Find(&billings).Error
		if err != nil {
			return apperrors.FromGORM(err, "billing", "")
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, 0, apperrors.Wrap(err, "fetch billings for list")
	}
	if err := r.attachBillingListItems(ctx, billings, clinicIDs); err != nil {
		return nil, 0, err
	}
	return billings, total, nil
}

// attachBillingListItems はページ内 billing の明細を取得して付与する。
// payments / payment_splits / 返金合計は廃止された保存先のため読まない。
func (r *accountingRepository) attachBillingListItems(ctx context.Context, billings []model.Billing, clinicIDs []uint64) error {
	if len(billings) == 0 {
		return nil
	}
	ids := make([]uint64, 0, len(billings))
	for i := range billings {
		ids = append(ids, billings[i].ID)
	}

	var items []model.BillingItem
	if err := r.db.WithContext(ctx).
		Scopes(persistence.BillingTenantScope("billing_items", clinicIDs)).
		Where("billing_items.billing_id IN ? AND billing_items.deleted_at IS NULL", ids).
		Find(&items).Error; err != nil {
		return apperrors.FromGORM(err, "billing_item", "")
	}

	itemsByBilling := make(map[uint64][]model.BillingItem, len(items))
	for _, item := range items {
		itemsByBilling[item.BillingID] = append(itemsByBilling[item.BillingID], item)
	}
	for i := range billings {
		billings[i].Items = itemsByBilling[billings[i].ID]
	}
	return nil
}

func (r *accountingRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	return r.findBillingByIDWithScope(persistence.DBOrTx(ctx, r.db).Scopes(persistence.ClinicScope(clinicID)), id)
}

func (r *accountingRepository) FindByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Billing, error) {
	return r.findBillingByIDWithScope(persistence.DBOrTx(ctx, r.db).Scopes(persistence.ClinicScopeIn(clinicIDs)), id)
}

// findBillingByIDWithScope は clinic スコープ適用済みのクエリで billing を1件取得し、明細だけを付与する。
// payments / payment_splits / refunds は廃止された保存先で読まず、owner / pet の入れ子も
// 氏名の reader（EMR-491）が接続されるまで付与しない（Q4(c) / Q8。held for EMR-506）。
func (r *accountingRepository) findBillingByIDWithScope(q *gorm.DB, id uint64) (*model.Billing, error) {
	var billing model.Billing
	err := q.
		Preload("Items", "deleted_at IS NULL").
		Where("id = ?", id).First(&billing).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", id))
	}
	return &billing, nil
}

// LockAndFindByID は FOR UPDATE で請求を行ロック取得する。明細の変更（billing_item_service）が
// claim_state の判定と同一 tx で使い、TOCTOU を防止する。
// BE-refactor.md R1-1 (D2): dbOrTx で ambient tx に参加する。参加しないと FOR UPDATE ロックが
// 別セッションで即座に解放され、TOCTOU 防止が機能しない。
func (r *accountingRepository) LockAndFindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error) {
	var billing model.Billing
	err := persistence.DBOrTx(ctx, r.db).
		Preload("Items", "deleted_at IS NULL").
		Scopes(persistence.ClinicScope(clinicID)).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).First(&billing).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", id))
	}
	return &billing, nil
}

// BE-refactor.md X-12: 会計完了(completed)時、Create は accounting_service_core.Create の
// Transactor.WithTx から txCtx 付きで呼ばれ、後続の reservation.CompleteForAccounting と単一 tx に
// 参加する（dbOrTx が無ければ従来どおり db.WithContext(ctx) と等価・挙動保存）。
func (r *accountingRepository) Create(ctx context.Context, clinicID uint64, accounting *model.Billing) error {
	accounting.ClinicID = clinicID
	if err := persistence.DBOrTx(ctx, r.db).Create(accounting).Error; err != nil {
		if persistence.IsUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("billing", accounting.ScheduledDate.String())
		}
		return apperrors.FromGORM(err, "billing", "")
	}
	return nil
}

// Update は指定フィールドのみを更新し、更新後のレコードを返す。
// GORM のゼロ値スキップは unexported update の map で回避する。
// P2: service 層で逆遷移を拒否（修正 1）、repo は RowsAffected チェックで clinic scope/soft-delete を検証
// BE-refactor.md R1-2: Cancel が本メソッドを ambient tx（監査と原子化）から呼ぶため dbOrTx で参加する。
// ambient tx が無ければ従来どおり db.WithContext(ctx) と等価（挙動保存）。
func (r *accountingRepository) Update(ctx context.Context, clinicID, billingID uint64, cmd AccountingUpdate) (*model.Billing, error) {
	return r.update(ctx, clinicID, billingID, cmd.toFields())
}

func (r *accountingRepository) update(ctx context.Context, clinicID, billingID uint64, fields map[string]any) (*model.Billing, error) {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Billing{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", billingID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "billing", fmt.Sprintf("%d", billingID))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("billing", fmt.Sprintf("%d", billingID))
	}
	var billing model.Billing
	if err := persistence.DBOrTx(ctx, r.db).
		Preload("Items", "deleted_at IS NULL").
		Scopes(persistence.ClinicScope(clinicID)).
		First(&billing, "id = ?", billingID).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("%d", billingID))
	}
	return &billing, nil
}
