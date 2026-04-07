# BUG-132: 一部ページで create 権限なしでも作成ボタンが表示される（BUG-124 修正漏れ）

## 概要
BUG-124 の修正で `MasterListPage` / `MasterCRUDPage` 経由のマスタページは create ボタンが
権限制御されるようになったが、カスタムレイアウトのページに修正が適用されていない。
また、ダッシュボード（当日の受付）の予約作成ボタンも権限制御されていない。

## 脆弱性分類
- **CWE-284**: Improper Access Control (UI 層)
- **影響**: セキュリティ実害なし（API で 403）。UX 問題。

## ブラウザテスト結果（RBAC検証用グループ: 一般花子）

### マスタページ

| ページ | パス | create 権限 | 作成ボタン | 判定 |
|--------|------|-----------|-----------|------|
| 物販・フード | `/settings/merchandise-items` | **F** | 「新しい品目を追加...」表示 | ❌ |
| トリミング | `/settings/trimming` | **F** | 「新規登録」表示 | ❌ |
| 診療サービス | `/settings/service-type` | F | 非表示 | ✅ |
| 権限グループ | `/settings/permission-groups` | F | 非表示 | ✅ |
| スタッフ管理 | `/settings/staff` | T | 表示 | ✅ |
| 保険マスタ | `/settings/insurance` | T | 表示 | ✅ |
| 動物種類 | `/settings/animal-species` | T | 表示 | ✅ |

### 診療系ページ

| ページ | パス | create 権限 | 作成ボタン | 判定 |
|--------|------|-----------|-----------|------|
| ダッシュボード | `/` | reservations C=**F** | 「新規予約登録」「新規追加」表示 | ❌ |
| 飼主 | `/owners` | T | 表示 | ✅ |
| 会計管理 | `/accounting` | F | 非表示 | ✅ |
| 在庫管理 | `/inventory` | F | 非表示 | ✅ |
| 検査管理 | `/examinations` | F | 非表示 | ✅ |
| 予約管理 | `/reservations` | F | 非表示 | ✅ |

## 影響範囲（修正が必要な3箇所）

### 1. `/settings/merchandise-items` — インラインフォーム
**ファイル**: `frontend/src/features/master/routes/MerchandiseItemSettings.tsx`

「新しい品目を追加...」ボタンが `usePermission` でガードされていない。
`MasterCRUDPage` を経由していないカスタムレイアウト。

### 2. `/settings/trimming` — トリミング設定
**ファイル**: `frontend/src/features/master/routes/TrimmingSettings.tsx`

「新規登録」ボタンが create 権限で制御されていない。

### 3. `/` — ダッシュボード（当日の受付）
**ファイル**: `frontend/src/features/reception/routes/Reception.tsx`

「新規予約登録」ボタンと各カンバンレーンの「新規追加」ボタンが
`reservations.create` 権限で制御されていない。

## 修正方針

各コンポーネントで `usePermission` を使って create ボタンを条件表示:

```typescript
const { canCreate } = usePermission(ResourceXxx);

// 新規登録ボタン
{canCreate ? <Button onClick={handleCreate}>新規登録</Button> : null}
```

ダッシュボードの場合:
```typescript
const { canCreate: canCreateReservation } = usePermission(ResourceReservations);

{canCreateReservation ? <Button>新規予約登録</Button> : null}
{canCreateReservation ? <Button>新規追加</Button> : null}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Conditional Render
> 必ず `? (...) : null`（`&&` 禁止）

### `.claude/rules/security.md`
> "Validate on both client and server"

UI 層の create ボタン非表示は多層防御の一環。API で 403 になるが、UX として不適切。

## 優先度
**Low** — セキュリティ実害なし。BUG-124 修正の残対応。

## 関連チケット
- BUG-124（修正済み・一部漏れ）: マスタページ操作ボタン表示制御

## 関連ファイル
- `frontend/src/features/master/routes/MerchandiseItemSettings.tsx` — インラインフォーム
- `frontend/src/features/master/routes/TrimmingSettings.tsx` — 新規登録ボタン
- `frontend/src/features/reception/routes/Reception.tsx` — 新規予約登録ボタン
