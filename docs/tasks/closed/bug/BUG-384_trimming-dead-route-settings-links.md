# BUG-384: `config/paths.ts` と router.tsx の settings ルート不整合による複数 dead route

**作成日**: 2026-04-15
**Status**: CLOSED
**Priority**: HIGH (運用中画面 → 404、設計一貫性破綻)
**Affects**: `frontend/src/config/paths.ts`, `frontend/src/app/router.tsx`, `features/master/constants/category-config.ts`, `features/trimming`

---

## 追加観察（ブラウザ検証 2026-04-15）

BUG-384 初出は trimming-course/option 2 件だったが、`rg "/settings/"` で paths.ts / category-config.ts を grep した結果、**router.tsx に対応実装がない「嘘のパス定義」が大量**に見つかった。新タブでブラウザ遷移検証：

| 定義元 | path | router.tsx 実装 | ブラウザ結果 |
|--------|------|------|------|
| paths.ts:265 | `/settings/vaccine` | なし | **404** |
| paths.ts:269 | `/settings/examination` | なし | **404** |
| paths.ts:273 | `/settings/trimming-course` | なし | **404** |
| paths.ts:277 | `/settings/trimming-option` | なし | **404** |
| paths.ts:281 | `/settings/consultation` | なし | **404** |
| paths.ts:285 | `/settings/procedure` | なし | 未確認（404 の可能性大） |
| paths.ts:289 | `/settings/diagnosis-type` | **Navigate redirect あり** | リダイレクト OK（BUG-382 CLOSED） |
| paths.ts:293 | `/settings/diagnosis-name` | **Navigate redirect あり** | リダイレクト OK（BUG-382 CLOSED） |
| paths.ts:302 | `/settings/interview/templates` | なし | 未確認（inquiry-templates に移行済の可能性） |
| category-config.ts:237 | `/settings/inquiry-template` | なし（複数形 `inquiry-templates` が正） | **404 想定** |

---

## 旧（初出）

---

## 概要

トリミング編集画面 `/trimming/{id}` で露出している 2 つの「マスタ管理」リンクが両方 **404 (ページが見つかりません)** を返す。BUG-382 と同種の dead route 問題。

## 404 になるリンク（ブラウザ検証 2026-04-15）

| リンク文言 | href | 遷移結果 | 現行相当ルート |
|-----------|------|---------|--------------|
| 「マスタ管理」（コース欄） | `/settings/trimming-course` | **404** | `/settings/trimming` (トリミングコース 5 件) |
| 「マスタ管理」（オプション欄） | `/settings/trimming-option` | **404** | `/settings/trimming` (オプションタブ) |

## 再現手順

1. `admin@noavet.jp` でログイン
2. `/trimming` → 任意記録の「編集」 → `/trimming/{id}` 遷移
3. コース欄横の **「マスタ管理」** リンクをクリック
4. 結果: **「ページが見つかりません」**（404）
5. オプション欄横の **「マスタ管理」** リンクをクリック
6. 結果: **「ページが見つかりません」**（404）

## 影響

- **トリマー業務影響**: コース/オプションを追加したくなってマスタ管理リンクをクリック → 404 → 作業断絶
- トリマーが独自で `/settings/trimming` に辿り着く必要があり UX 悪化

## 修正方針

### Option A (推奨): router リダイレクト
```tsx
// frontend/src/app/router.tsx
{ path: "/settings/trimming-course", loader: () => redirect("/settings/trimming") },
{ path: "/settings/trimming-option", loader: () => redirect("/settings/trimming?tab=options") },
```

### Option B: href を直接修正
`frontend/src/features/trimming/routes/TrimmingForm.tsx` で:
```tsx
// 変更前
<Link to="/settings/trimming-course">マスタ管理</Link>
<Link to="/settings/trimming-option">マスタ管理</Link>

// 変更後
<Link to="/settings/trimming">マスタ管理</Link>
<Link to="/settings/trimming?tab=options">マスタ管理</Link>
```

## 関連

- **BUG-382 (CLOSED)**: 類似 dead route 4 件 (job-title / service-type / diagnosis-type / diagnosis-name)。本件はその後も残存していた 2 件の trimming 系。
- CLAUDE.md `.claude/rules/naming-conventions.md` (API パス複数形 kebab-case)

## 同系統の dead route 洗い出し結果 (ブラウザ検証 2026-04-15)

```bash
$ rg "/settings/" frontend/src/config/paths.ts frontend/src/features/master/constants/category-config.ts
```

で発見した **11 件のパス定義のうち、少なくとも 5 件は router.tsx に実装がなく 404**（実ブラウザ遷移で確認）。

## 推奨リファクタリング方針

### Step 1: router.tsx に不足リダイレクト追加
BUG-382 パターンで `/settings/vaccine` `/settings/examination` `/settings/consultation` `/settings/procedure` `/settings/trimming-course` `/settings/trimming-option` `/settings/inquiry-template`（単数形）すべてを `/settings/treatment-items` or `/settings/trimming` or `/settings/inquiry-templates` にリダイレクト。

### Step 2: `config/paths.ts` と `category-config.ts` を router.tsx と同期
単一の真実の源は router.tsx 側。paths.ts の「嘘のパス定義」を削除し、`getHref()` 結果が常に到達可能なルートを返すことを保証する。

### Step 3: 整合性を保つ CI チェック追加
paths.ts で `path` プロパティに定義された各値について、router.tsx で同一 path が実装されているか検証する lint ルールまたはテストを追加推奨。
