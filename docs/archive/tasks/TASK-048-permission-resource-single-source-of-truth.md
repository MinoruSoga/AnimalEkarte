# TASK-048: 権限リソース定義の単一情報源化

**作成日**: 2026-03-29
**ステータス**: Open
**優先度**: High
**領域**: Backend + Frontend / Security / Architecture

---

## 背景・問題

権限チェックに使用するリソースキー（`"owners"`, `"medical-records"` 等 15件）が
**フロントエンドのソースコードにハードコード**されており、DB・バックエンド・フロントエンドの
3箇所が「信頼ベースの一致」で動いている。

```
現在の状態（問題あり）:

  frontend/src/features/master/types/permission-resources.ts
    PERMISSION_RESOURCES = [{ key: "dashboard", ... }, ...]   ← ハードコード定数

  backend/internal/model/clinic.go
    PermissionGroupRule.Resource string                        ← 型制約なし（任意文字列）

  backend/migrations/001_init.sql
    resource varchar(50) NOT NULL                             ← DB制約なし

  frontend/src/app/router.tsx
    <RequirePermission resource="medical-records">            ← 文字列リテラル直書き
```

### 具体的なリスク

| リスク | 内容 |
|--------|------|
| **Typo による権限素通り** | `"medical-records"` → `"medicalrecords"` と書いてもコンパイルエラーにならず、`hasPermission()` が常に `false` を返して全員アクセス不可、またはチェックが無効化される |
| **新機能追加時の同期漏れ** | 新リソースを追加する際に `permission-resources.ts` / DB seed / router.tsx の3箇所を手動で更新する必要があり、1箇所でも漏れると権限制御の穴になる |
| **バックエンド認可ミドルウェア実装時（BUG-056）の不一致** | バックエンドで `RequirePermission` ミドルウェアを実装する際、Go 側にリソース定数がないため文字列リテラルを使い続けることになりタイポリスクが続く |
| **テスト困難** | リソースキーに型がないため、無効なキーを渡してもコンパイル時・型チェック時に気づけない |

---

## 解決方針: Go 定数を単一情報源とする

このプロジェクトは **tygo による Goモデル → `models.ts` 自動生成パイプライン**が整備済み。
この仕組みを活用し、Go 側のリソース定数定義を唯一の真実とする。

```
修正後の情報フロー:

  backend/internal/model/permission.go
    type Resource string
    const (
      ResourceDashboard      Resource = "dashboard"
      ResourceOwners         Resource = "owners"
      ResourceMedicalRecords Resource = "medical-records"
      ...
    )
          ↓ make codegen（tygo）
  frontend/src/types/generated/models.ts
    export type Resource = string
    export const ResourceDashboard      = "dashboard"
    export const ResourceOwners         = "owners"
    export const ResourceMedicalRecords = "medical-records"
    ...
          ↓
  permission-resources.ts を廃止
  router.tsx / RequirePermission は models.ts の定数を参照
  DB の CHECK 制約でリソースキーを強制
```

### 採用しない案とその理由

| 案 | 採用しない理由 |
|----|----------------|
| `permission_resources` テーブルをDB管理 | リソースはアプリのコードと1対1対応するため実行時に変更する意味がない。API を増やすだけでコード管理と乖離する |
| フロント定数をそのまま維持 | 現状維持。バックエンド認可ミドルウェア（BUG-056）実装時に必ず問題になる |

---

## 実装スコープ

### BE-077: Go 側の Resource 型定数定義 + DB 制約追加

- `backend/internal/model/permission.go` に `Resource` 型と 15定数を定義
- `backend/migrations/001_init.sql` の `permission_group_rules.resource` に CHECK 制約を追加
- バックエンドで文字列リテラルを使用している箇所（seed SQL、handler）を定数に置換
- `make codegen` 実行で `models.ts` に定数を自動出力されることを確認

### FE-132: フロントエンドの定数移行

- `permission-resources.ts` の `PERMISSION_RESOURCES` キーを `models.ts` 生成定数に置換
- `router.tsx` の `<RequirePermission resource="...">` を定数参照に変更
- `usePermission("...")` 呼び出しを定数参照に変更
- `permission-resources.ts` を削除（または label のみのマッピングとして縮小）

---

## 完了条件

- [ ] `Resource` 型が Go 側で定義されており、任意文字列を `Resource` 型に代入するとコンパイルエラーになる
- [ ] `make codegen` 実行後、`models.ts` に `ResourceXxx` 定数が出力されている
- [ ] `permission_group_rules.resource` に DB CHECK 制約が追加されており、不正なリソースキーの INSERT がエラーになる
- [ ] `permission-resources.ts` が削除または label のみのファイルに縮小されている
- [ ] `router.tsx` および `usePermission()` 呼び出し箇所で文字列リテラルが残っていない
- [ ] `docker compose exec frontend npm run build` が成功する
- [ ] `docker compose exec backend go build ./...` が成功する
- [ ] `docker compose exec frontend npm run lint` がエラー 0 件

---

## 依存タスク・関連チケット

| チケット | 関係 |
|----------|------|
| BUG-056 | バックエンド認可ミドルウェア実装時に本タスク完了が前提条件になる |
| BUG-054 | クロスクリニック権限バグ（本タスクとは独立して対処可能） |

---

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BE-077 | Backend | `Resource` 型定数定義・DB CHECK 制約・codegen 確認 |
| FE-132 | Frontend | 定数への移行・`permission-resources.ts` 削除 |
| FE-133 | Frontend | サイドバーメニューを権限でフィルタリング（`canView = false` を非表示） |
| FE-134 | Frontend | `usePermission` hook を各 feature コンポーネントに統合（ボタン・フォームの表示制御） |
