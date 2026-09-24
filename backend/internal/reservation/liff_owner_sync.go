package reservation

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// liffOwnerLineLookup は ownerRepo が任意で実装する line_user_id 検索 capability。
// liffOwnerRepo（service_deps.go）は name/phone 検索のみ宣言するため、
// 具象 repository の FindByLineUserID を consumer-side の assertion で検出する。
// capability が無い注入では heal を行わず顧客をそのまま返す（fail-safe）。
type liffOwnerLineLookup interface {
	FindByLineUserID(ctx context.Context, clinicID uint64, lineUserID string) (*model.Owner, error)
}

// findCustomerWithOwnerSync は LINE 顧客を取得し、必要なら owner 紐付けを自己修復する。
//
// BUG-LIFF-HEALTHCARD-OWNER-SYNC: LinkAccount（internal/lstep/line_link_service.go）が
// owners.line_user_id だけを更新し line_customers.owner_id を同期しないため、
// 連携済み顧客が OwnerID=NULL のまま残り、健康手帳・プロフィール・予約owner反映が
// 表示名フォールバックに落ちる。ここでは読み取り経路で安全に修復する。
//
// 修復は「verified owners.line_user_id が同一 clinic に一致し、かつ
// line_customers.owner_id が NULL」のときだけ、clinic スコープの単一
// UpdateOwnerLink UPDATE を実行する。name/phone 照合は行わない（SEC-CS2-F02）。
// 既に同一 owner に紐付済みなら書き込まない（冪等）。別 owner への紐付けは
// 上書きせず警告を残して fail-closed で未変更の顧客を返す。検索・更新の失敗は
// 読み取りを失敗させず、未リンクの顧客を返す（fail-safe degrade）。
//
// 注意: リンク時に owners.line_user_id と line_customers.owner_id を同一
// トランザクションで書く本来の2行アトミック書き込みは internal/lstep 側の
// 責務であり本 lane の範囲外。ここでの修復は読み取り時の自己修復であり、
// リンク時 atomicity を代替するものではない（残存する cross-lane 制限）。
func (s *liffService) findCustomerWithOwnerSync(ctx context.Context, clinicID, customerID uint64) (*model.LineCustomer, error) {
	customer, err := s.customerRepo.FindByID(ctx, clinicID, customerID)
	if err != nil {
		return nil, err
	}
	if customer == nil || customer.LineUserID == "" {
		return customer, nil
	}
	lookup, ok := s.ownerRepo.(liffOwnerLineLookup)
	if !ok {
		return customer, nil
	}
	owner, err := lookup.FindByLineUserID(ctx, clinicID, customer.LineUserID)
	if err != nil {
		if !apperrors.IsNotFound(err) {
			slog.WarnContext(ctx, "liff owner sync: owner lookup failed; returning unlinked customer (best-effort)",
				"clinic_id", clinicID, "customer_id", customerID, "error", err)
		}
		return customer, nil
	}
	if owner == nil {
		return customer, nil
	}
	if customer.OwnerID != nil {
		if *customer.OwnerID != owner.ID {
			slog.WarnContext(ctx, "liff owner sync: customer linked to a different owner; leaving unchanged (fail-closed)",
				"clinic_id", clinicID, "customer_id", customerID, "owner_id", *customer.OwnerID, "resolved_owner_id", owner.ID)
		}
		return customer, nil
	}
	if err := s.customerRepo.UpdateOwnerLink(ctx, clinicID, customerID, &owner.ID); err != nil {
		slog.WarnContext(ctx, "liff owner sync: owner link update failed; returning unlinked customer (best-effort)",
			"clinic_id", clinicID, "customer_id", customerID, "owner_id", owner.ID, "error", err)
		return customer, nil
	}
	slog.InfoContext(ctx, "liff owner sync: linked line customer to owner",
		"clinic_id", clinicID, "customer_id", customerID, "owner_id", owner.ID)
	healed, err := s.customerRepo.FindByID(ctx, clinicID, customerID)
	if err != nil || healed == nil {
		if err != nil {
			slog.WarnContext(ctx, "liff owner sync: re-read after link failed; returning pre-heal customer (best-effort)",
				"clinic_id", clinicID, "customer_id", customerID, "owner_id", owner.ID, "error", err)
		}
		return customer, nil
	}
	return healed, nil
}
