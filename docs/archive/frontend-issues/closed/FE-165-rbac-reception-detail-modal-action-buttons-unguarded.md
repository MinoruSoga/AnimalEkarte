# FE-165: 当日の受付 — ReceptionDetailModal のアクションボタン全般が canEdit ガードなし

## 概要

`/`（当日の受付）の `Reception.tsx` は `usePermission(ResourceReservations)` から `canCreate` しか取得していない。`ReceptionDetailModal` に渡す `onEdit`・`onCancel`・`onConfirm` が `canEdit` / `canDelete` チェックなしで常に渡されるため、閲覧のみユーザーでも予約の編集・取消・ステータス変更ができる UI になっている。

## 影響範囲

| ボタン | 操作 | API | ガード |
|-------|------|-----|--------|
| 「編集」ボタン（`onEdit`） | 予約編集ダイアログを開く → PATCH | PATCH `/v1/reservations/:id` | ❌ なし |
| 「取消」ボタン（`onCancel`） | 予約取消確認 → DELETE/PATCH | DELETE/PATCH `/v1/reservations/:id` | ❌ なし |
| 「受付済にする」「診察を開始する」「診察を終了する」等（`onConfirm`） | ステータス変更 → PATCH | PATCH `/v1/reservations/:id/status` | ❌ なし |
| 「カルテ作成」「カルテ入力」（`onCreateMedicalRecord`） | カルテ新規作成へ遷移 | POST `/v1/medical-records` | ❌ medical-records 権限チェックなし |
| 「会計へ進む」（`onCreateAccounting`） | 会計新規作成へ遷移 | POST `/v1/billings` | ❌ accounting 権限チェックなし |
| 「入院登録」（`onCreateHospitalization`） | 入院新規作成へ遷移 | POST `/v1/hospitalizations` | ❌ hospitalization 権限チェックなし |

## 根本原因

```tsx
// Reception.tsx 行 45 — canCreate のみ取得 ❌
const { canCreate: canCreateReservation } = usePermission(ResourceReservations);

// 行 370-376 — onEdit/onCancel を canEdit チェックなしで常に渡す ❌
<ReceptionDetailModal
  onConfirm={handleAdvanceStatus}   // ← ステータス変更、canEdit チェックなし
  onEdit={handleEditAppointment}    // ← 編集、canEdit チェックなし
  onCancel={handleCancelAppointment} // ← 取消、canDelete チェックなし
  ...
/>
```

```tsx
// AppointmentCard.tsx 行 164-189 — カルテ/会計/入院ボタンに各リソースの canCreate チェックなし ❌
<button onClick={handleKarteClick}>カルテ</button>           // medical-records canCreate なし
<button onClick={handleAccountingClick}>会計</button>        // accounting canCreate なし
<button onClick={handleHospitalizationClick}>入院</button>  // hospitalization canCreate なし
```

## 期待する挙動

- `canEdit=false` → 「編集」「受付済にする」「診察を開始する」等のステータス変更ボタンが非表示
- `canDelete=false`（reservations）→ 「取消」ボタンが非表示
- `canCreate=false`（medical-records）→ 「カルテ作成」ボタンが非表示
- `canCreate=false`（accounting）→ 「会計へ進む」ボタンが非表示
- `canCreate=false`（hospitalization）→ 「入院登録」ボタンが非表示

## 修正方針

```tsx
// Reception.tsx — 必要な権限を全て取得
const { canCreate: canCreateReservation, canEdit: canEditReservation, canDelete: canDeleteReservation } = usePermission(ResourceReservations);
const { canCreate: canCreateMedicalRecord } = usePermission(ResourceMedicalRecords);
const { canCreate: canCreateAccounting } = usePermission(ResourceAccounting);
const { canCreate: canCreateHospitalization } = usePermission(ResourceHospitalization);

// onEdit/onCancel を条件付きで渡す
<ReceptionDetailModal
  onConfirm={canEditReservation ? handleAdvanceStatus : undefined}
  onEdit={canEditReservation ? handleEditAppointment : undefined}
  onCancel={canDeleteReservation ? handleCancelAppointment : undefined}
  canCreateMedicalRecord={canCreateMedicalRecord}
  canCreateAccounting={canCreateAccounting}
  canCreateHospitalization={canCreateHospitalization}
  ...
/>
```

```tsx
// ReceptionDetailModal.tsx RelatedPages — 各リソースの canCreate を受け取りガード
{canCreateMedicalRecord ? (
  <button onClick={onCreateMedicalRecord}>カルテ</button>
) : null}
{canCreateAccounting ? (
  <button onClick={onCreateAccounting}>会計</button>
) : null}
{canCreateHospitalization ? (
  <button onClick={onCreateHospitalization}>入院</button>
) : null}
```

## 優先度

**HIGH** — 閲覧のみユーザーが予約の編集・取消・ステータス変更・カルテ/会計/入院作成を試みることができる。ステータス変更は診療フロー全体に影響するため、誤った操作は業務上の問題を引き起こす。

## 関連ファイル

- `frontend/src/features/reception/routes/Reception.tsx` (行 45, 373-376)
- `frontend/src/features/reception/components/ReceptionDetailModal.tsx` (行 59-96, 138-276)
- `frontend/src/features/reception/components/AppointmentCard.tsx` (行 164-193: カルテ/会計/入院ミニボタン — canCreate/canEdit チェックなし)
- `frontend/src/features/reception/hooks/use-reception-kanban.ts` (行 167: `moveCard` — canEdit チェックなし → Kanban ドラッグ&ドロップでステータス変更が誰でも可能, 行 237/253: `advanceStatus` — canEdit チェックなし, 行 299: `cancelAppointment` — canDelete チェックなし)
- 発見日: 2026-04-07（RBAC Phase 2/3 テスト中）
