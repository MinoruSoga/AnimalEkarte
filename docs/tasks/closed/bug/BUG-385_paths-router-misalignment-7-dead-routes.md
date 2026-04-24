# BUG-385: `config/paths.ts` と `router.tsx` の不整合で 7 件の settings ルートが 404

**作成日**: 2026-04-15
**Status**: OPEN
**Priority**: **HIGH** (設計一貫性破綻・運用画面からリンク切れが大量に発生)
**Affects**: `frontend/src/config/paths.ts`, `frontend/src/app/router.tsx`, `frontend/src/features/master/constants/category-config.ts`

---

## 概要

`src/config/paths.ts`（全ルートの型安全 URL マップ）と `features/master/constants/category-config.ts` に定義されている `/settings/xxx` パスのうち **7 件が router.tsx に対応実装がなく 404** を返す。`paths.ts` が「嘘の URL」を `getHref()` として提供しており、これを使った `<Link>` は作った瞬間に dead link になる**設計バグ**。

過去 BUG-382 で 4 件（job-title / service-type / diagnosis-type / diagnosis-name）、BUG-384 で 2 件（trimming-course / trimming-option）のリダイレクト実装が一部行われたが、他の 5 件（vaccine / examination / consultation / procedure / inquiry-template）は手付かず。本 issue で全件網羅的な修正を要求。

## 確定 dead route（ブラウザ検証 2026-04-15）

新タブで直接 URL アクセスし `notFound: true`（「ページが見つかりません」表示）を 7 件確認：

| # | URL | 定義元 | 現行相当 |
|---|-----|--------|----------|
| 1 | `/settings/vaccine` | paths.ts:265, category-config.ts:77 | `/settings/treatment-items?tab=vaccination` |
| 2 | `/settings/examination` | paths.ts:269, category-config.ts:67 | `/settings/treatment-items?tab=examination` |
| 3 | `/settings/consultation` | paths.ts:281, category-config.ts:97 | `/settings/treatment-items?tab=consultation` |
| 4 | `/settings/procedure` | paths.ts:285, category-config.ts:117 | `/settings/treatment-items?tab=procedure` |
| 5 | `/settings/trimming-course` | paths.ts:273, category-config.ts:147 | `/settings/trimming` |
| 6 | `/settings/trimming-option` | paths.ts:277, category-config.ts:157 | `/settings/trimming?tab=options` |
| 7 | `/settings/inquiry-template` | category-config.ts:237 | `/settings/inquiry-templates`（複数形） |

## 影響（本番運用中リンク切れ確定箇所・5 ページ）

`src/components/shared/MasterLink.tsx` が `paths.settings.xxx.getHref()` を `<Link to>` に流し込んでおり、以下 5 箇所で **運用画面に dead link が露出**:

| # | 画面 | ファイル:行 | 対象マスタ | 表示ラベル |
|---|------|-----------|----------|----------|
| 1 | トリミング編集 | `features/trimming/routes/TrimmingForm.tsx:96` | `trimming_course` | 「マスタ管理」 |
| 2 | トリミング編集 | `features/trimming/routes/TrimmingForm.tsx:133` | `trimming_option` | 「マスタ管理」 |
| 3 | 検査フォーム | `features/examinations/routes/ExaminationForm.tsx:103` | `examination` | 「編集」 |
| 4 | 予防接種フォーム | `features/vaccinations/routes/VaccinationForm.tsx:239` | `vaccine` | 「編集」 |
| 5 | カルテ内予防接種フォーム | `features/medical-records/components/VaccinationForm.tsx:86` | `vaccine` | 「編集」 |

- **運用中のリンク切れ**: トリマー・医師・看護師が「マスタ管理」「編集」リンクをクリック → 404 で作業断絶
- **型安全性の破綻**: `config/paths.ts` は「ハードコードされたURLパス文字列は禁止。getHref() を使用」という CLAUDE.md ルールの背骨だが、実装上は嘘を返す。ルール遵守の努力が無効化される
- **将来の同類バグ量産**: paths.ts を信頼して新規 `<Link>` を追加するたび dead link が増える

### 実害のない（未使用の）定義
`consultation` / `procedure` / `inquiry-template` は `MasterLink` などから参照されておらず、paths.ts 上の嘘定義のみ。クリーンアップのみで良い。

## 修正方針

### Step 1: router.tsx に 7 件リダイレクト追加（BUG-382 パターン）

```tsx
// frontend/src/app/router.tsx:680 の後に追加
{ path: "vaccine", element: <Navigate to="/settings/treatment-items?tab=vaccination" replace /> },
{ path: "examination", element: <Navigate to="/settings/treatment-items?tab=examination" replace /> },
{ path: "consultation", element: <Navigate to="/settings/treatment-items?tab=consultation" replace /> },
{ path: "procedure", element: <Navigate to="/settings/treatment-items?tab=procedure" replace /> },
{ path: "trimming-course", element: <Navigate to="/settings/trimming" replace /> },
{ path: "trimming-option", element: <Navigate to="/settings/trimming?tab=options" replace /> },
{ path: "inquiry-template", element: <Navigate to="/settings/inquiry-templates" replace /> },
```

### Step 2: `config/paths.ts` を router.tsx と整合

嘘のパス定義を削除または修正。`getHref()` の戻り値が router に存在するルートを必ず返すよう保証。

### Step 3: `category-config.ts` の `settingsPath` も同期修正

ファイル内 20+ のエントリを router.tsx の実在ルートに揃える。

### Step 4 (推奨): CI テスト追加

router.tsx で定義された path セットと、paths.ts の `path` 値セットが subset 関係にあることを検証する Vitest テスト。

```ts
// frontend/src/__tests__/paths-router-consistency.test.ts
const routerPaths = extractPathsFromRouter(); // router.tsx から再帰抽出
const definedPaths = extractPathsFromPathsTs(); // paths.ts から抽出
definedPaths.forEach(p => {
  expect(routerPaths.has(p) || routerPaths.hasRedirect(p)).toBe(true);
});
```

## 関連

- **BUG-382 (CLOSED)**: 旧 dead route 4 件の先行修正
- **BUG-384 (CLOSED)**: trimming 系 2 件の先行修正試み（本 issue で再整理）
- CLAUDE.md `config/paths.ts でURL管理`（ルール）

## ブラウザ検証再現手順

```bash
# 各 URL を順次新タブで開く（全て「ページが見つかりません」を確認）
http://localhost:3003/settings/vaccine
http://localhost:3003/settings/examination
http://localhost:3003/settings/consultation
http://localhost:3003/settings/procedure
http://localhost:3003/settings/trimming-course
http://localhost:3003/settings/trimming-option
http://localhost:3003/settings/inquiry-template
```

## 確認済み**正常**ルート（参考）

`/settings/interview/templates` は router.tsx に実装あり、問診テンプレートマスタ表示 OK。inquiry-templates（複数形）も動作確認済。
