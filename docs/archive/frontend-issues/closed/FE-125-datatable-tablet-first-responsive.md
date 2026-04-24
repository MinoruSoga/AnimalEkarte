# FE-125: DataTable タブレットファースト対応と列幅統一

**Status**: Open
**Priority**: High
**Affects**: DataTable 共通コンポーネント + 全一覧ページ
**Date Created**: 2026-03-26
**Related**: TASK-030, FE-124

## Summary

現状のテーブルは `min-w-[800px]` 固定でタブレット（iPad 768px〜1024px）での横スクロールが避けられない。Tailwind の `hidden lg:table-cell` パターンでタブレットで不要なカラムを非表示にし、また各ページのカラム幅を統一する。DataTable 本体の `min-w` も緩和する。

## 現状のコード

### DataTable の min-w 固定

```typescript
// frontend/src/components/shared/DataTable/DataTable.tsx:31
<Table className="min-w-[800px]">
```

iPad（768px）縦持ち + サイドバー（256px）= テーブル幅 **512px** しかない。
→ 800px の min-w により **横スクロール必須**。

### OwnersList — 11 カラムでタブレット表示不可

```typescript
// frontend/src/features/owners/routes/OwnersList.tsx:343-409
const columns = [
  { header: "飼主No", className: "w-[100px]" },
  { header: "飼主名", className: "w-[180px]" },
  { header: "ペット番号", className: "w-[100px]" },
  { header: "ペット名", className: "w-[120px]" },
  { header: "生死", className: "w-[60px]" },
  { header: "種", className: "w-[60px]" },
  { header: "生年月日", className: "w-[100px]" },
  { header: "体重", className: "w-[80px]" },
  { header: "環境", className: "w-[120px]" },
  { header: "前回来院", className: "w-[100px]" },
  { header: "操作", className: "w-[100px]" },
];
// 合計 min-width ≈ 1120px — タブレット絶対不可
```

### カラム幅不統一の例

```
日付カラム:   w-[120px]（多数）、w-[140px]（Accounting のみ）
ステータス:   w-[80px]（Examinations）、w-[100px]（多数）、w-[110px]（Estimates）
操作:         w-[60px]（Estimates）、w-[80px]（Inventory, Examinations）、w-[100px]（多数）
```

## 必要な変更

### 1. DataTable.tsx — min-w を緩和

```typescript
// Before
<Table className="min-w-[800px]">

// After
<Table className="min-w-[640px]">
```

640px = iPad（768px）縦持ちからサイドバー最小幅を引いた値。
`overflow-auto` はすでにあるため、実際のテーブルが 640px を超えれば横スクロールが発生する設計は維持される。

### 2. 各ページのタブレット非表示カラム

以下のように `className` に `hidden lg:table-cell` を追加する。
`DataTable` は `col.className` をヘッダーセルに適用するが、ボディセルの `TableCell` にも同じ className を付与する必要がある。

#### 飼主一覧 (`OwnersList.tsx`)

```typescript
// Before（11 カラム）
{ header: "飼主No", className: "w-[100px]" },
{ header: "ペット番号", className: "w-[100px]" },
{ header: "生死", className: "w-[60px]" },
{ header: "生年月日", className: "w-[100px]" },
{ header: "体重", className: "w-[80px]" },
{ header: "環境", className: "w-[120px]" },
{ header: "前回来院", className: "w-[100px]" },

// After（lg 以上で表示）
{ header: "飼主No", className: "w-[100px] hidden lg:table-cell" },
{ header: "ペット番号", className: "w-[100px] hidden lg:table-cell" },
{ header: "生死", className: "w-[60px] hidden lg:table-cell" },
{ header: "生年月日", className: "w-[100px] hidden lg:table-cell" },
{ header: "体重", className: "w-[80px] hidden lg:table-cell" },
{ header: "環境", className: "w-[120px] hidden lg:table-cell" },
{ header: "前回来院", className: "w-[100px] hidden lg:table-cell" },
```

対応する `TableCell` にも同じ className を追加：
```typescript
<TableCell className={`${STYLE.tableCell} whitespace-nowrap hidden lg:table-cell`}>
  {p.ownerNo}
</TableCell>
```

#### カルテ管理 (`MedicalRecords.tsx`)

```typescript
// タブレット非表示カラム
{ header: "種", className: "w-[80px] hidden lg:table-cell" },
{ header: "関連", className: "w-[100px] hidden lg:table-cell" },
```

#### 検査管理 (`Examinations.tsx`)

```typescript
// タブレット非表示カラム
{ header: "結果概要" },  // → hidden lg:table-cell を追加
```

#### トリミング管理 (`TrimmingList.tsx`)

```typescript
// タブレット非表示カラム（3カラム）
{ header: "種", className: "w-[80px] hidden lg:table-cell" },
{ header: "体重", className: "w-[80px] hidden lg:table-cell" },
{ header: "スタイル希望" },  // → hidden lg:table-cell を追加
```

#### 入院・ホテル管理 (`HospitalizationListView.tsx`)

```typescript
// タブレット非表示カラム
{ header: "種", className: "w-[80px] hidden lg:table-cell" },
{ header: "退院予定日", className: "w-[120px] hidden lg:table-cell" },
```

#### 定期健診 (`CheckupsList.tsx`)

```typescript
// タブレット非表示カラム
{ header: "次回予定", className: "w-[120px] hidden lg:table-cell" },
{ header: "結果・所見" },  // → hidden lg:table-cell を追加
```

#### 会計管理 (`Accounting.tsx`)

```typescript
// タブレット非表示カラム
{ header: "カルテ", className: "w-[80px] hidden lg:table-cell" },
```

#### 在庫管理 (`InventoryList.tsx`)

```typescript
// タブレット非表示カラム
{ header: "最低在庫", className: "w-[100px] hidden lg:table-cell" },
{ header: "有効期限", className: "w-[120px] hidden lg:table-cell" },
```

### 3. カラム幅統一（統一規約の適用）

```
DATE_COL:   "w-[120px]"   日付 YYYY-MM-DD
STATUS_COL: "w-[100px]"   ステータスバッジ
ACTION_COL: "w-[80px]"    操作ボタン（DropdownまたはButton）
DOCTOR_COL: "w-[100px]"   担当医名
NARROW_COL: "w-[80px]"    短い固定値（種・体重等）
```

変更が必要な箇所：

| ファイル | 現状 | 修正後 |
|---------|------|-------|
| `Accounting.tsx` 日時 | `w-[140px]` | `w-[120px]` |
| `Examinations.tsx` ステータス | `w-[80px]` | `w-[100px]` |
| `Examinations.tsx` 操作 | `w-[80px]` | `w-[80px]` ✅ |
| `EstimateList.tsx` ステータス | `w-[110px]` | `w-[100px]` |
| `EstimateList.tsx` 操作 | `w-[60px]` | `w-[80px]` |

### 4. タブレット表示後の残存カラム確認

各ページのタブレット(lg未満)での表示カラム：

| ページ | 表示されるカラム |
|--------|---------------|
| 飼主一覧 | 飼主名、ペット名、種、操作（4カラム） |
| カルテ管理 | 診療日、飼主名、ペット名、主訴、担当医、ステータス、操作（7カラム） |
| 検査管理 | 日時、飼主名、ペット名、検査種別、担当医、ステータス、操作（7カラム） |
| 予防接種 | 実施日、飼主名、ペット名、予防接種名、次回予定、操作（6カラム）— 変更なし |
| トリミング | 診療日、飼主名、ペット名、担当、ステータス、操作（6カラム） |
| 入院管理 | 入院No、飼主名、ペット名、タイプ、入院開始日、ステータス、操作（7カラム） |
| 定期健診 | 実施日、飼主名、ペット名、健診種別、担当医、操作（6カラム） |
| 会計管理 | 日時、飼主名、ペット名、請求金額、支払方法、ステータス、操作（7カラム） |
| 在庫管理 | 品名、カテゴリ、在庫数、保管場所、ステータス、操作（6カラム） |
| 見積書 | 変更なし（7カラム） |

## プロジェクトルール遵守チェック

- [x] `any` 型なし
- [x] `FC` / `forwardRef` なし
- [x] barrel index 経由 import なし
- [x] 条件レンダー `? ... : null`
- [x] `useCallback` / `useMemo` の deps は primitive

## 依存関係

- FE-124 が先に完了していることが望ましい（同ファイルへの変更が重なる場合）

## 完了条件

- [ ] `DataTable.tsx` の min-w が `min-w-[640px]` になっている
- [ ] iPad Pro（1024px）でサイドバー展開時に全主要一覧が横スクロールなし（目視確認）
- [ ] iPad（768px 縦持ち相当）で少なくとも飼主名・ペット名・ステータス・操作が表示される
- [ ] 日付・ステータス・操作カラムの幅が全ページで統一されている
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス（ESLint エラーなし）
