# BUG-280: 安全に削除可能なデッドコード — 空ディレクトリ・未使用コンポーネント

## 概要
プロジェクト全体のどこからもimportされていない未使用コンポーネントと、ファイルが存在しない空ディレクトリが残存している。削除リスクなし。

## 再現手順
1. `frontend/src/components/shared/SearchBox/` を確認 → ファイルが1件も存在しない
2. `DateRangePicker.tsx` を grep → 自ファイル以外でimportされていない
3. `grep -r "DateRangePicker" frontend/src --include="*.tsx" --include="*.ts" -l` → 1件（自身のみ）

## 期待する動作
- 未使用コンポーネントはリポジトリに存在しない
- 空ディレクトリはリポジトリに存在しない

## 現状コード

### `frontend/src/components/shared/SearchBox/`
```
（ファイルが存在しない空ディレクトリ）
```

### `frontend/src/components/shared/DateRangePicker/DateRangePicker.tsx`
```typescript
// プロジェクト全体でimportなし
// grep 結果: 自ファイル以外 0件
```

### 比較: 正しい実装（プロジェクト内参照実装）
```typescript
// 使用中のコンポーネント例:
// frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx
// → 多数のファイルからimportされている
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/components/shared/SearchBox/` | 空ディレクトリ | 削除対象 |
| `frontend/src/components/shared/DateRangePicker/DateRangePicker.tsx` | 未使用コンポーネント | 削除対象 |

## 修正方針

### 1. 空ディレクトリ削除

```bash
rm -rf frontend/src/components/shared/SearchBox/
```

### 2. 未使用コンポーネント削除

```bash
rm -rf frontend/src/components/shared/DateRangePicker/
```

削除前に念のため最終確認:
```bash
grep -r "DateRangePicker" frontend/src --include="*.tsx" --include="*.ts" -l
# 出力: 0件 → 削除安全
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — ベストプラクティス
> Feature外部（app/等）からのインポートは必ず `index.ts` を経由

使われていないコンポーネントはバンドルサイズへの影響はないが（tree-shaking済み）、コードベースの認知的負荷を増やす。

### プロジェクト内参照実装
- `components/shared/NotionDatePicker/NotionDatePicker.tsx` — 正しく使用されている共有コンポーネントの例

## 優先度
**Low** — 動作への影響なし。コードベースのノイズ削減目的。

## 関連チケット
- BUG-281: DataStates barrel index.ts 欠落
- BUG-282: SidePeek barrel index.ts 欠落

## 関連ファイル
- `frontend/src/components/shared/SearchBox/` — 空ディレクトリ
- `frontend/src/components/shared/DateRangePicker/DateRangePicker.tsx` — 未使用コンポーネント
