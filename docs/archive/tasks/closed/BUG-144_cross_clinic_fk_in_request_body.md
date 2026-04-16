# BUG-144: リクエストボディ内の FK ID でクロスクリニック参照が可能

## 概要
予約作成時の `staff_id` に他クリニックのスタッフ ID を指定すると、
クロスクリニックの FK 参照が成功する。X-Clinic-ID ヘッダーのチェック（BUG-128 修正済み）は
クエリスコープのみに適用されており、リクエストボディ内の FK ID はクリニック境界を超えて参照可能。

## 脆弱性分類
- **CWE-639**: Authorization Bypass Through User-Controlled Key (IDOR)
- **影響**: 他クリニックのスタッフを自クリニックの予約に割り当て可能。データ整合性の破壊。

## 再現手順
```bash
# 八王子院 (clinic_id=3) のユーザーでログイン
# 城東医院のスタッフ (staff_id=16) を予約に割り当て
curl -X POST /api/v1/reservations \
  -H 'Content-Type: application/json' \
  -d '{
    "pet_id": 1,
    "owner_id": 1, 
    "service_type_id": 1,
    "staff_id": 16,
    "start_time": "2026-04-15T10:00:00Z",
    "end_time": "2026-04-15T11:00:00Z"
  }'
# → 201 Created ❌
```

## ブラウザテスト結果

| テスト | 期待 | 実際 |
|--------|------|------|
| `staff_id=16`（城東医院）で予約作成 | 400/403 | **201（成功）** ❌ |
| `owner_id=25`（城東医院）で pet 作成 | 400 | 400 ✅ |

## 影響範囲

リクエストボディに FK ID を含む全エンドポイントが影響を受ける可能性:

| エンドポイント | FK フィールド | チェック |
|--------------|-------------|---------|
| `POST /reservations` | `staff_id` | ❌ 未チェック |
| `POST /reservations` | `pet_id`, `owner_id` | 要確認 |
| `POST /vaccinations` | `pet_id`, `vaccine_id` | 要確認 |
| `POST /examinations` | `pet_id` | 要確認 |
| `POST /trimmings` | `pet_id` | 要確認 |
| `POST /medical-records` | `pet_id`, `doctor_id` | 要確認 |
| `POST /hospitalizations` | `pet_id`, `cage_id` | 要確認 |

## 修正方針

### Service 層で FK ID のクリニック所属を検証

```go
func (s *ReservationService) Create(ctx context.Context, clinicID uint64, input CreateInput) error {
    // staff_id がこのクリニックに所属しているか
    if input.StaffID != nil {
        staff, err := s.staffRepo.FindByID(ctx, *input.StaffID)
        if err != nil || !s.isStaffInClinic(staff, clinicID) {
            return apperrors.WrapInvalidInput("指定されたスタッフはこのクリニックに所属していません")
        }
    }
    // pet_id, owner_id も同様にクリニック所属チェック
}
```

### または DB レベルで制約

FK 制約に `clinic_id` を含める複合外部キーを使用。ただし既存スキーマの大幅変更が必要。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/database-design.md` — マルチテナント設計
> **全テーブルに `clinic_id`** / **WHERE 句は `clinic_id` から開始**

FK 参照時もクリニック境界を維持すべき。

### `.claude/rules/api.md` — Security
> "Validate all user input"

リクエストボディ内の FK ID もクリニックスコープで検証が必要。

## 優先度
**Medium** — データ整合性の問題。直接的なデータ漏洩ではないが、他クリニックのリソースを
自クリニックのレコードに紐づけられる。

## 関連チケット
- BUG-128（修正済み）: X-Clinic-ID ヘッダーのクロスクリニックアクセス

## 関連ファイル
- `backend/internal/service/reservation_service.go` — Create
- `backend/internal/service/vaccination_service.go` — Create
- 全 Service 層の Create/Update メソッド
