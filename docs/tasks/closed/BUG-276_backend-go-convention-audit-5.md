# BUG-276: バックエンド Go コード規約準拠監査（第5回 — 最終）

## 概要

第4回監査（BUG-270〜275）修正後の最終確認。CRITICAL/HIGH 違反はゼロ。
残存は MEDIUM レベルのコード品質改善のみ。

## 検証結果

- `go vet ./...` — クリーン
- `go test ./internal/service/... ./internal/handler/... ./internal/repository/...` — 全 PASS

## CRITICAL / HIGH

**なし。**

## MEDIUM 残存事項

### 1. slog 監査ログ欠落（2箇所）

| ファイル | メソッド | 内容 |
|----------|---------|------|
| `account_service.go:43` | UpdatePasswordHash | パスワード変更ログなし（監査上重要） |
| `reservation_schedule_service.go:73,103` | Upsert, Delete | スケジュール変更ログなし |

### 2. Handler → Repository 直接アクセス（認証基盤、3箇所）

| ファイル | 行 | 内容 | 許容理由 |
|----------|-----|------|----------|
| `auth_handler.go` | 563 | `h.repos.PermissionGroup.GetEffectivePermissionsByStaffID` | 認証/認可基盤 |
| `clinic_handler.go` | 63 | 同上 | 認証/認可基盤 |
| `reservation_line_routes.go` | 57 | `middleware.LiffAuth(h.repos....)` | middleware 初期化 |

これらは認証基盤のアーキテクチャ上やむを得ない箇所。service 層への移管は可能だが優先度は低い。

### 3. Repository clinicID なしメソッド（設計上の意図あり）

| リポジトリ | 理由 |
|-----------|------|
| `account_repository` | email ベースログイン。クリニック横断が設計意図 |
| `animal_species_repository` | グローバルマスタ（ClinicID カラムなし） |
| `clinic_repository` | クリニック自体のCRUD |
| `permission_group_repository` | staff 紐付けで横断アクセス |
| `record_image_repository` | handler で所有権チェック済み |

### 4. not-implemented エンドポイント（2箇所）

| ファイル | 行 | 内容 |
|----------|-----|------|
| `reservation_course_handler.go` | 162 | `c.JSON(http.StatusNotImplemented, ...)` |
| `reservation_staff_handler.go` | 180 | 同上 |

実装予定の画像アップロードエンドポイント。`// NOTE: Intentional direct response` コメント付き。

## 修正済み項目（第1回〜第4回監査）

| 監査回 | BUG 範囲 | 主要修正 |
|--------|---------|----------|
| 第1回 | BUG-244〜252 | Price ポインタ、staff ビジネスロジック、clinical_plan clinic_id、FromGORM、naked return、auth直接repo、gorm.Is |
| 第2回 | BUG-253〜260 | multitenancy 8ドメイン、Reorder FromGORM、naked return 20+、slog 8サービス、handler c.JSON、FK check |
| 第3回 | BUG-261〜266 | multitenancy 第2波（treatment_plan, vital等）、naked return 41箇所、slog 18箇所、FromGORM 5箇所、json タグ、secret |
| 第4回 | BUG-270〜275 | handler→repo 直接アクセス解消、slog 8箇所、Reorder 二重ラップ、mock source、swaggerignore、liff URLエンコード、FK check |

## 結論

**4回の監査で主要な規約違反は全て解消された。**

残存する MEDIUM 事項は、`account_service` のパスワード変更ログと `reservation_schedule_service` のスケジュール変更ログの2サービス分のみ。これらは機能的な影響はなく、監査証跡の完全性に関する改善項目である。

## 実施日

2026-04-10
