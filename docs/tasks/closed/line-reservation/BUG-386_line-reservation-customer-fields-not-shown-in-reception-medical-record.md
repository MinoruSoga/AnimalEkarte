# BUG-386: LINE予約フォームの飼い主名・ペット名・犬種が受付/カルテ予約詳細に反映されない

**作成日**: 2026-04-15
**Status**: CLOSED
**Priority**: **HIGH** (機能定義「予約フォームからの電子カルテへの自動入力」と実装が不一致で、受付運用に必要な患者情報が欠落する)
**Affects**: `frontend/src/features/reception/api/transforms.ts`, `backend/internal/service/reservation_validators.go`, `backend/internal/service/medical_record_service.go`, `frontend/line-reserve/src/pages/ConfirmPage.tsx`

---

## 概要

LINE予約フォームで入力した `飼い主名` `ペット名` `犬種` は、予約作成時に `customer_fields` へ保存されるが、受付画面の予約詳細とカルテ自動作成フローで参照されていない。結果として、LINE予約は受付一覧/詳細で患者情報が空欄になり、受付済み変更時のカルテ自動作成も `owner_id` / `pet_id` 欠落によりスキップされる。

## 脆弱性分類（セキュリティ系の場合）

- 該当なし（本 issue は LINE予約と電子カルテ連携の機能不整合）

## 再現手順

1. LINE予約 LIFF で任意の医院を開く
2. STEP 1 で `飼い主名` `ペット名` `犬種` を入力して予約を確定する
3. 管理画面で同日の受付一覧を開き、該当予約の詳細を表示する
4. 予約ステータスを `受付済` に変更する
5. **結果**: 予約詳細の飼い主名・ペット名・犬種が空欄のままになる。さらに `owner_id` / `pet_id` が無いためカルテ自動作成がスキップされる

## 期待する動作

- LINE予約フォーム入力値が受付/予約詳細に表示されること
- `owner_id` / `pet_id` が未紐付けでも `customer_fields` の値でフォールバック表示されること
- 紐付け済み、または紐付け可能なケースでは `owner_id` / `pet_id` が予約に保持され、受付済み変更時にカルテ自動作成へ連携されること

## 現状コード

### `frontend/line-reserve/src/pages/ConfirmPage.tsx:102`
```tsx
const reservation = await liffApi.createReservation(
  clinicId,
  {
    course_id: flow.courseId,
    staff_id: flow.staffId,
    date: flow.date,
    start_time: flow.startTime,
    end_time: flow.endTime,
    customer_fields: {
      customer_name: flow.customerInfo.name,
      phone: flow.customerInfo.phone,
      owner_name: flow.customerInfo.ownerName,
      pets: flow.customerInfo.pets.map(p => ({
        name: p.name,
        type: p.type,
        is_new: p.isNew,
      })),
    },
    request_text: flow.requestText,
  },
  idToken,
);
```

### `backend/internal/service/reservation_validators.go:184`
```go
appt := &model.Appointment{
	ClinicID:          input.ClinicID,
	StartTime:         startDT,
	EndTime:           endDT,
	ReservationTypeID: input.ReservationTypeID,
	DoctorID:          doctorID,
	Status:            model.ReservationStatusConfirmed,
	Source:            model.ReservationSourceLine,
	LineCustomerID:    &input.CustomerID,
	IsStaffDelegated:  input.StaffID == 0,
	CustomerFields:    customerFields,
	Notes:             notes,
	VisitType:         model.VisitTypeRevisit,
}
```

### `frontend/src/features/reception/api/transforms.ts:71`
```tsx
const petName = reservation.pet?.name ?? "";
const petType = reservation.pet?.animal_species?.name
  ?? (reservation.pet?.animal_species_id ? ANIMAL_SPECIES_MAP[reservation.pet.animal_species_id] : "犬")
  ?? "犬";
const ownerName = reservation.owner?.name ?? "";
```

### `backend/internal/service/medical_record_service.go:293`
```go
func (s *medicalRecordService) AutoCreateFromReservation(ctx context.Context, clinicID uint64, reservation *model.Appointment) {
	if reservation.PetID == nil || reservation.OwnerID == nil {
		slog.WarnContext(ctx, "autoCreateFromReservation: skipped — reservation has no pet_id or owner_id",
			slog.Uint64("reservation_id", reservation.ID))
		return
	}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// frontend/src/features/reservations/api/transforms.ts:24
const cf = parseCustomerFields(reservation.customer_fields);
const ownerName =
  reservation.owner?.name ??
  reservation.pet?.owner?.name ??
  cf.owner_name ??
  cf.customer_name ??
  "";
const petName = reservation.pet?.name ?? cf.pets?.[0]?.name ?? "";
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/line-reserve/src/pages/ConfirmPage.tsx` | LIFF予約作成時に `customer_fields` を送信 | 影響あり |
| `backend/internal/service/reservation_validators.go` | LINE予約作成時に `owner_id` / `pet_id` を保持していない | 影響あり |
| `frontend/src/features/reception/api/transforms.ts` | 受付一覧/詳細の表示変換で `customer_fields` をフォールバックしていない | 影響あり |
| `frontend/src/features/reservations/api/transforms.ts` | 別画面では `customer_fields` フォールバック実装が存在する | 参照実装あり |
| `backend/internal/service/medical_record_service.go` | `owner_id` / `pet_id` 欠落時にカルテ自動作成をスキップ | 影響あり |
| `docs/line/reservation-spec.md:2116` | 仕様では `customer_fields` を既存UI互換のフォールバック前提としている | 仕様不一致 |

## 修正方針

### 1. 受付画面変換で `customer_fields` をフォールバックする — `frontend/src/features/reception/api/transforms.ts:65`
```tsx
interface CustomerFieldsJSON {
  customer_name?: string;
  owner_name?: string;
  pets?: Array<{ name?: string; type?: string }>;
}

function parseCustomerFields(raw: string | undefined): CustomerFieldsJSON {
  if (!raw) return {};
  try {
    return JSON.parse(raw) as CustomerFieldsJSON;
  } catch {
    return {};
  }
}

export function transformReservationToReceptionAppointment(
  reservation: BackendReceptionReservation
): ReceptionAppointment {
  const cf = parseCustomerFields(reservation.customer_fields);

  const petName = reservation.pet?.name ?? cf.pets?.[0]?.name ?? "";
  const petType = reservation.pet?.animal_species?.name
    ?? (reservation.pet?.animal_species_id ? ANIMAL_SPECIES_MAP[reservation.pet.animal_species_id] : undefined)
    ?? cf.pets?.[0]?.type
    ?? "犬";
  const ownerName = reservation.owner?.name ?? cf.owner_name ?? cf.customer_name ?? "";

  return {
    // 既存フィールド維持
  };
}
```

### 2. LINE予約の owner / pet 紐付けを予約レコードへ反映する — `backend/internal/service/liff_service.go:263`, `backend/internal/repository/appointment_repository.go`
```go
appt, err := s.validators.ValidateAndCreate(ctx, input)
if err != nil {
	return nil, apperrors.Wrap(err, "failed to validate and create appointment")
}

if len(input.CustomerFields) > 0 && string(input.CustomerFields) != "{}" {
	if err := s.customerRepo.UpdateAdditionalFields(ctx, clinicID, customerID, input.CustomerFields); err != nil {
		slog.WarnContext(ctx, "failed to update customer additional fields (best-effort)", "error", err)
	}
}

s.tryAutoLinkOwner(ctx, clinicID, customerID, input.CustomerFields)

// owner 紐付け後、予約へも owner_id / pet_id を反映する
customer, err := s.customerRepo.FindByID(ctx, clinicID, customerID)
if err == nil && customer != nil && customer.OwnerID != nil {
	var petID *uint64
	if profilePetID := resolvePetIDFromCustomerFields(customer, input.CustomerFields); profilePetID != nil {
		petID = profilePetID
	}

	if err := s.reservationRepo.UpdateOwnerPet(ctx, clinicID, appt.ID, customer.OwnerID, petID); err != nil {
		slog.WarnContext(ctx, "failed to update appointment owner/pet link (best-effort)", "error", err)
	}
}
```

```go
// backend/internal/repository/appointment_repository.go
func (r *appointmentRepository) UpdateOwnerPet(
	ctx context.Context,
	clinicID, appointmentID uint64,
	ownerID, petID *uint64,
) error {
	result := r.db.WithContext(ctx).
		Model(&model.Appointment{}).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", appointmentID).
		Updates(map[string]any{
			"owner_id": ownerID,
			"pet_id":   petID,
		})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "appointment", fmt.Sprintf("%d", appointmentID))
	}
	return nil
}
```

### 3. 回帰テストを追加する — `frontend/src/features/reception/api/transforms.test.ts`, `backend/internal/service/medical_record_service_test.go`
```ts
it("should fallback to customer_fields for line reservations without owner/pet relation", () => {
  const result = transformReservationToReceptionAppointment({
    source: "line",
    owner: undefined,
    pet: undefined,
    customer_fields: JSON.stringify({
      owner_name: "田中太郎",
      pets: [{ name: "ポチ", type: "柴犬" }],
    }),
  } as BackendReceptionReservation);

  expect(result.ownerName).toBe("田中太郎");
  expect(result.petName).toBe("ポチ");
  expect(result.petType).toBe("柴犬");
});
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Frontend ベストプラクティス参照実装
> **型安全性最優先**: Go/TypeScript 共に `any` を禁止し、厳格な型定義を行う。

`customer_fields` のパース追加でも `any` を使わず、専用の `CustomerFieldsJSON` 型と安全な JSON parse を使うこと。

### `.claude/CLAUDE.md` — バックエンド・アーキテクチャ規約
> **アーキテクチャ遵守**: handler → service → repository の軽量レイヤードを徹底。

`owner_id` / `pet_id` の補完ロジックは handler へ直接追加せず service 層に集約し、受付/カルテの連携責務を service に置くこと。

### `.claude/rules/typescript-react.md` — TypeScript / React 19 Rules
> **❌ 禁止: any**

受付画面の変換ロジックへ `customer_fields` フォールバックを追加する際、パース結果は明示型で扱うこと。

### `.claude/rules/testing.md` — Testing Rules
> **Bug fixes: Add regression test**

受付表示フォールバックとカルテ自動作成スキップ条件の双方に回帰テストを追加すること。

### プロジェクト内参照実装

- `frontend/src/features/reservations/api/transforms.ts:24` — `customer_fields` のフォールバック表示
- `backend/internal/service/liff_service.go:268` — LIFF予約後にプロフィール保存を行う best-effort 実装

## 優先度

**High** — 患者情報が受付/カルテに出ず、LINE予約の主要要件が未達のまま運用影響が発生しているため。

## 関連チケット

- BUG-387: 2回目以降のLINE予約でお客様情報が自動復元されない
  - 独立して修正可能だが、LINE予約まわりの回帰確認は合わせて実施する

## 関連ファイル

- `frontend/line-reserve/src/pages/ConfirmPage.tsx:102` — LINE予約作成リクエスト組み立て
- `backend/internal/service/reservation_validators.go:184` — LINE予約レコード作成
- `frontend/src/features/reception/api/transforms.ts:65` — 受付画面用の予約変換
- `frontend/src/features/reservations/api/transforms.ts:21` — `customer_fields` フォールバックの参照実装
- `backend/internal/service/medical_record_service.go:293` — 受付済み変更時のカルテ自動作成
- `docs/line/reservation-spec.md:2116` — `customer_fields` と既存カラムの仕様
