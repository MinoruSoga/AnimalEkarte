# BUG-271: Handler → Repository 直接アクセス（Service 迂回）

## 概要

handler 層から repository を直接呼び出しており、handler → service → repository のレイヤード規約に違反。
ビジネスロジックが handler に漏出し、テスト困難かつ slog 監査ログも出力されない。

## 影響範囲

### `staff_handler.go` — 5箇所

| 行 | メソッド | 呼び出し |
|-----|---------|----------|
| 118 | CreateStaff（リロード） | `h.repos.Staff.FindByID(ctx, staff.ID)` |
| 251 | GetStaffPermissionGroups | `h.repos.PermissionGroup.GetGroupIDsByStaffID(...)` |
| 283 | SetStaffPermissionGroups | `h.repos.PermissionGroup.SetStaffGroups(...)` |
| 363 | GetStaffExcludedServiceTypes | `h.repos.ReservationStaff.FindExcludedServiceTypes(...)` |
| 399 | SetStaffExcludedServiceTypes | `h.repos.ReservationStaff.ReplaceExcludedServiceTypes(...)` |

### `reservation_staff_handler.go` — 4箇所（うち1箇所は N+1 クエリ）

| 行 | メソッド | 呼び出し | 備考 |
|-----|---------|----------|------|
| 26 | ListReservationStaffs | `h.repos.ReservationStaff.FindExcludedServiceTypes(...)` | **ループ内 = N+1** |
| 59 | CreateReservationStaff | 同上 | 単件リロード |
| 95 | UpdateReservationStaff | 同上 | 単件リロード |
| 142 | PatchReservationStaffStatus | 同上 | 単件リロード |

### `permission_group_handler.go` — 1箇所

| 行 | メソッド | 呼び出し |
|-----|---------|----------|
| 143 | SetPermissionGroupRules | `h.repos.PermissionGroup.GetGroupIDsByStaffID(...)` |

## 修正方針

1. **StaffService に PermissionGroup/ExcludedServiceType 操作メソッドを追加**:
   - `GetPermissionGroupIDs(ctx, clinicID, staffID) ([]uint64, error)`
   - `SetPermissionGroupIDs(ctx, clinicID, staffID, groupIDs []uint64) error`
   - `GetExcludedServiceTypes(ctx, clinicID, staffID) ([]model.ServiceType, error)`
   - `SetExcludedServiceTypes(ctx, clinicID, staffID, typeIDs []uint64) error`

2. **ReservationStaffService の List レスポンスに excluded を含める** (N+1 解消):
   - Preload パターンまたは `FindExcludedServiceTypesByStaffIDs(ctx, ids)` バルク取得

3. **handler から `h.repos.` 呼び出しを全て `h.svc.` に置換**

## 準拠すべきプロジェクト規約

### `backend/CLAUDE.md` — 依存関係の方向
> handler (プレゼンテーション層) → service (ビジネスロジック層) → repository (データアクセス層)
> 上位層は下位層にのみ依存。逆方向禁止。

## 優先度

**High** — アーキテクチャ違反。reservation_staff_handler の N+1 は本番パフォーマンスにも影響。

## 関連チケット

- BUG-246: staff_handler ビジネスロジック漏出（第1回監査）
- BUG-250: auth_handler 直接 repo アクセス（第1回監査）
- BUG-270: 第4回監査 親チケット
