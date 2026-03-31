# FE-090: 予約作成400エラー（リクエスト型不一致）

**Status**: Open
**Priority**: High
**Affects**: 予約管理 (`/reservations`) 新規予約作成
**Date Created**: 2026-03-21
**Related**: BUG-006

## Summary

予約作成時に400エラーが発生する。原因は3点:
1. `pet_id` / `owner_id` が数値型 `uint64` を期待する BE に対して文字列型 `string` で送信されている
2. `service_type` フィールド名が誤り（正しくは `service_type_id`）、かつ名前文字列を送信している
3. `doctor_id` に医師名（文字列）を送信している（BE は `*uint64` を期待）

## 現状のコード

```typescript
// frontend/src/features/reservations/api/types.ts:1-16
export interface CreateReservationRequest {
  pet_id: string;       // ← BE は uint64 を期待
  owner_id: string;     // ← BE は uint64 を期待
  doctor_id?: string;   // ← BE は *uint64 を期待
  start_time: string;
  end_time: string;
  visit_type: string;
  service_type: string; // ← フィールド名が誤り（正: service_type_id）かつ名前を送信
  is_designated?: boolean;
  notes?: string;
}
```

```typescript
// frontend/src/features/reservations/api/transforms.ts:24-39
export const transformToCreateRequest = (
  data: Partial<ReservationAppointment>,
  petId: string,
  ownerId: string
): CreateReservationRequest => {
  return {
    pet_id: petId,           // ← "1" (string) → BE が uint64 で受け取れない
    owner_id: ownerId,       // ← "1" (string) → BE が uint64 で受け取れない
    start_time: ...,
    end_time: ...,
    visit_type: data.visitType ?? "first",
    service_type: data.type ?? "",  // ← フィールド名誤り + "一般診療"等の名前を送信
    doctor_id: data.doctor || undefined,  // ← "山田 太郎"等の名前を送信
    is_designated: data.isDesignated ?? false,
    notes: data.notes,
  };
};
```

```typescript
// frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:198-202
// service_type は item.name を value として設定している（IDを使っていない）
{serviceTypes.map((item) => (
  <SelectItem key={item.id} value={item.name}>  // ← value={item.id} にすべき
    {item.name}
  </SelectItem>
))}

// frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:272-274
// doctor も s.name を value として設定している（IDを使っていない）
activeStaff.map((s) => (
  <SelectItem key={s.id} value={s.name}>  // ← value={String(s.id)} にすべき
```

バックエンドが期待するリクエスト形式（`reservation_request.go:11-13`）:
```go
ServiceTypeID uint64   `json:"service_type_id" binding:"required"`
DoctorID      *uint64  `json:"doctor_id"`
// pet_id, owner_id も uint64
```

## 必要な変更

### 1. CreateReservationRequest 型を修正

```typescript
// frontend/src/features/reservations/api/types.ts
export interface CreateReservationRequest {
  pet_id: number;           // string → number
  owner_id: number;         // string → number
  doctor_id?: number;       // string → number（optional）
  start_time: string;
  end_time: string;
  visit_type: string;
  service_type_id: number;  // service_type: string → service_type_id: number
  is_designated?: boolean;
  notes?: string;
}
```

### 2. transformToCreateRequest() を修正

```typescript
// frontend/src/features/reservations/api/transforms.ts
export const transformToCreateRequest = (
  data: Partial<ReservationAppointment>,
  petId: string,
  ownerId: string
): CreateReservationRequest => {
  return {
    pet_id: Number(petId),                          // string → number
    owner_id: Number(ownerId),                      // string → number
    start_time: data.start ? data.start.toISOString() : "",
    end_time: data.end ? data.end.toISOString() : "",
    visit_type: data.visitType ?? "first",
    service_type_id: Number(data.serviceTypeId),    // フィールド名変更 + number型
    doctor_id: data.doctorId ? Number(data.doctorId) : undefined,  // ID で送信
    is_designated: data.isDesignated ?? false,
    notes: data.notes,
  };
};
```

### 3. ReservationFormFields.tsx のセレクト値をIDに変更

```typescript
// frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx

// service_type: value を item.name → String(item.id) に変更
{serviceTypes.map((item) => (
  <SelectItem key={item.id} value={String(item.id)}>
    {item.name}
  </SelectItem>
))}
// onValueChange: formData.type → formData.serviceTypeId に変更

// doctor: value を s.name → String(s.id) に変更
activeStaff.map((s) => (
  <SelectItem key={s.id} value={String(s.id)}>
    {s.name}
  </SelectItem>
))
// onValueChange: formData.doctor → formData.doctorId に変更
```

### 4. ReservationAppointment 型 / ReservationFormData 型を更新

`formData.type`（サービス種別名）→ `formData.serviceTypeId`（サービス種別ID）に変更し、表示用の名前は別途保持するか、マスターデータから逆引きする。`formData.doctor`（医師名）→ `formData.doctorId`（スタッフID）に変更し、表示用名前はマスターデータから取得。

フォームデータ型が変更されると、フォームの既存バリデーション・初期化・表示ロジックに影響する可能性があるため、`ReservationFormModal.tsx` 内の初期化処理（`doctor: "医師A"` のハードコード）も修正が必要。

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] 型は `models.ts` から導出

## 依存関係

- バックエンドの `service_type_id` / `doctor_id` の型定義（`reservation_request.go`）が確定していること

## 完了条件

- [ ] 予約新規作成で 400 エラーが発生しない
- [ ] `pet_id` / `owner_id` が数値で送信される
- [ ] `service_type_id` が数値IDで送信される
- [ ] `doctor_id` が数値IDで送信される（未選択時は undefined）
- [ ] 予約がカレンダーに反映される
- [ ] `npm run build` が通る
- [ ] `npm run lint` がエラーなし
