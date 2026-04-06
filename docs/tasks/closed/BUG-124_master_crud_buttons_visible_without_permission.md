# BUG-124: マスタページで create/edit/delete 権限がないユーザーに操作ボタンが表示される

## 概要
全マスタ設定ページで、`create`/`edit`/`delete` 権限がないユーザーにも「新規登録」ボタンや
各行の「操作」ボタン（編集・削除）が表示される。API は BUG-122 修正により 403 で正しくブロックするが、
ボタンが見えること自体が UX として不適切。

## 脆弱性分類
- **CWE-284**: Improper Access Control (UI層)
- **影響**: セキュリティ実害なし（API で 403）。UX 上の問題 — ユーザーがボタンをクリック → エラーという不快な体験。

## 再現手順
1. `vet@example.com` / `password`（一般権限）でログイン
2. `http://localhost:3003/settings/staff` にアクセス
3. **結果**: 「新規登録」ボタンが表示され、各行に「操作」ボタンが表示される

## 一般権限の master-staff パーミッション
```json
{
  "master-staff": {
    "view": true,
    "create": false,
    "edit": false,
    "delete": false
  }
}
```

view のみ許可されているため、一覧表示は正しい。しかし create/edit/delete 権限がないのに
操作 UI が表示されている。

## 期待する動作
- `create` 権限がない場合 → 「新規登録」ボタンを非表示
- `edit` 権限がない場合 → 行の「操作」ボタン（編集）を非表示
- `delete` 権限がない場合 → 行の「操作」ボタン（削除）を非表示
- 全操作権限がない場合 → 「操作」カラム自体を非表示

## 影響範囲

全マスタ設定ページに共通する問題:

| ページ | パス | リソース |
|--------|------|---------|
| スタッフマスタ | `/settings/staff` | `master-staff` |
| 職種マスタ | `/settings/occupations` | `master-staff` |
| 権限グループマスタ | `/settings/permission-groups` | `master-permission` |
| 動物種類マスタ | `/settings/animal-species` | `master-animal-species` |
| 診療項目マスタ | `/settings/treatment-items` | `master-medical` |
| 診断マスタ | `/settings/diagnosis` | `master-medical` |
| 薬剤マスタ | `/settings/medicine` | `master-medical` |
| 問診テンプレート | `/settings/inquiry-templates` | `master-medical` |
| 予約区分マスタ | `/settings/service-type` | `master-service-type` |
| 入院マスタ | `/settings/hospitalization` | `master-hospitalization` |
| ケージマスタ | `/settings/cage` | `master-hospitalization` |
| トリミングマスタ | `/settings/trimming` | `master-trimming` |
| 商品マスタ | `/settings/merchandise-items` | `master-merchandise` |
| 保険マスタ | `/settings/insurance` | `master-insurance` |
| 医院設定 | `/settings/clinic` | `hospital-settings` |

## 現状コード

### `frontend/src/features/master/routes/StaffSettings.tsx`（例）

「新規登録」ボタンが無条件に表示されている:
```typescript
<PageLayout
  title="スタッフマスタ"
  actions={
    <Button onClick={handleCreate}>
      <Plus className={ICON.button} />
      新規登録
    </Button>
  }
>
```

各行の「操作」ボタンも無条件:
```typescript
<DropdownMenu>
  <DropdownMenuTrigger asChild>
    <Button variant="ghost" size="icon">操作</Button>
  </DropdownMenuTrigger>
  <DropdownMenuContent>
    <DropdownMenuItem onClick={() => handleEdit(staff)}>編集</DropdownMenuItem>
    <DropdownMenuItem onClick={() => handleDelete(staff.id)}>削除</DropdownMenuItem>
  </DropdownMenuContent>
</DropdownMenu>
```

## 修正方針

### 1. usePermission フックで権限を取得

各マスタページコンポーネント内で:
```typescript
import { usePermission } from "@/features/auth";

export function StaffSettings() {
  const { canCreate, canEdit, canDelete } = usePermission("master-staff");
  // ...
}
```

### 2. 「新規登録」ボタンの条件表示

```typescript
<PageLayout
  title="スタッフマスタ"
  actions={
    canCreate ? (
      <Button onClick={handleCreate}>
        <Plus className={ICON.button} />
        新規登録
      </Button>
    ) : null
  }
>
```

### 3. 「操作」ボタンの条件表示

```typescript
{canEdit || canDelete ? (
  <DropdownMenu>
    <DropdownMenuTrigger asChild>
      <Button variant="ghost" size="icon">操作</Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent>
      {canEdit ? <DropdownMenuItem onClick={() => handleEdit(staff)}>編集</DropdownMenuItem> : null}
      {canDelete ? <DropdownMenuItem onClick={() => handleDelete(staff.id)}>削除</DropdownMenuItem> : null}
    </DropdownMenuContent>
  </DropdownMenu>
) : null}
```

### 4. 「操作」カラムの条件表示

DataTable のカラム定義で操作カラムを条件付きにする:
```typescript
const columns = useMemo(() => {
  const cols = [nameCol, occupationCol, groupCol, statusCol];
  if (canEdit || canDelete) {
    cols.push(actionCol);
  }
  return cols;
}, [canEdit, canDelete]);
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Frontend ベストプラクティス
- **Conditional Render**: `? (...) : null` を使用（`&&` 禁止）
- **Design Tokens**: ボタンスタイルは既存の `ICON.button` 等を維持

### `.claude/rules/typescript-react.md` — React 19 Patterns
- **useMemo**: カラム定義は `useMemo` でキャッシュし、権限変更時のみ再計算
- **useCallback**: ハンドラは `useCallback` で安定化

### `.claude/rules/security.md` — Input Validation
> "Validate on both client and server"

UI 層での操作ボタン非表示は **UX 改善** が主目的。
API 層（BUG-122 修正済み）が唯一の信頼できる認可境界。
本修正は多層防御の UI 層として機能する。

### プロジェクト内参照実装

診療系ページ（owners, medical-records 等）の `router.tsx` では `action="create"` の
`RequirePermission` ガードでページレベルの制御を行っている。
マスタページでは同一ページ内にリスト + 作成/編集があるため、ボタンレベルの制御が必要。

## 優先度
**Medium** — セキュリティ実害なし（API で 403）。UX 改善。全15マスタページに影響。

## 関連チケット
- BUG-121（修正済み）: `/settings` ルートガード
- BUG-122（修正済み）: バックエンド API 権限チェック — ボタンクリック時に 403 で防御
- BUG-123（修正済み）: マスタインデックスカード権限フィルタリング

## 関連ファイル
- `frontend/src/features/master/routes/StaffSettings.tsx` — 修正対象（代表例）
- `frontend/src/features/master/routes/*.tsx` — 全マスタ設定ページ（15ファイル）
- `frontend/src/features/auth/hooks/use-auth.tsx` — `usePermission` フック
- `frontend/src/features/auth/index.ts` — `usePermission` エクスポート
