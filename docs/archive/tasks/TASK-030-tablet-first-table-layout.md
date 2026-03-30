# TASK-030: テーブル一覧ページのタブレットファースト UI 改善

**作成日**: 2026-03-26
**ステータス**: Open
**依頼元**: 一覧ページのテーブル表示の各項目の横幅などずれてる箇所があるので修正してください。タブレットファーストな表示UIにしてください。

---

## 概要

全一覧ページを横断するテーブル表示の品質問題を解消し、iPad（768px〜1024px）でも快適に使えるタブレットファーストなレイアウトに改善する。

## 依頼内容（原文）

> 一覧ページのテーブル表示の各項目の横幅などずれてる箇所があるので修正してください。
> タブレットファーストな表示UIにしてください。

## 仕様確認ログ

確認事項なし（コードベース調査で問題を特定済み）

---

## 現状の問題点

### 問題 1: テキストサイズ・セルパディング不統一

コードベース全調査で判明した inconsistency：

| ページ | 問題 | 該当カラム |
|--------|------|-----------|
| 検査管理 (`Examinations.tsx`) | **全セルが `text-sm`** — 他の全ページは `text-base` | 全7カラム |
| カルテ管理 (`MedicalRecords.tsx`) | 主訴セルのみ `text-sm` — 他セルは `text-base` | 主訴 |
| 検査管理, 会計管理, 予防接種 等 | セル padding が `py-2` — デザイントークン `STYLE.tableCell` は `py-2.5` | 全セル |

デザイントークン定義（`src/lib/design-tokens.ts`）：
```typescript
tableCell:     `text-base ${C.text} py-2.5`,   // 正規
tableCellMono: `font-mono text-base ${C.text} py-2.5`,
```

現状ページは大半が `py-2` でハードコード → デザイントークンと乖離。

### 問題 2: テーブルの `min-w-[800px]` がタブレットで横スクロールを強制

`DataTable.tsx:31` に `<Table className="min-w-[800px]">` が固定されている。

- iPad（768px 縦持ち）にサイドバー（256px）があると残り **512px** しかない
- iPad Pro ランドスケープ（1024px）でも残り **768px** — 8〜11 カラムのテーブルは収まらない
- 特に OwnersList（11 カラム）は最悪

### 問題 3: カラム数が多すぎるページでタブレット表示が破綻

| ページ | カラム数 | タブレット(768px)での問題 |
|--------|---------|------------------------|
| 飼主一覧 | 11 | 横スクロール必須、必須情報が見えない |
| トリミング管理 | 9 | 体重・スタイル希望が潰れる |
| カルテ管理 | 9 | 主訴・関連など余白がない |

### 問題 4: カラム幅の不統一

同種データでも幅がバラバラ：

| データ種別 | 現状の幅 | 出現ページ |
|-----------|---------|-----------|
| 日付カラム | `w-[120px]` または `w-[140px]` | 会計管理のみ 140px |
| ステータス | `w-[80px]` / `w-[100px]` / `w-[110px]` | バラバラ |
| 操作 | `w-[60px]` / `w-[80px]` / `w-[100px]` | バラバラ |
| 担当医 | `w-[100px]` (固定) または 幅指定なし | |

---

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | テキストサイズ・パディング統一（Examinations + MedicalRecords） | FE | FE-124 | - | [ ] |
| 2 | DataTable タブレット対応 + 各ページ列幅・列表示の最適化 | FE | FE-125 | FE-124 | [ ] |

---

## 受入条件（Acceptance Criteria）

- [ ] AC-1: 検査管理一覧のテーブルセルフォントが他ページと同じ `text-base` になっている
- [ ] AC-2: カルテ管理の主訴セルが `text-base` になっている（truncate は維持）
- [ ] AC-3: iPad Pro（1024px × 768px）でサイドバー展開時に主要一覧ページが横スクロールなしで表示される
- [ ] AC-4: iPad（768px）でも飼主一覧・カルテ・検査等が「最低限の必須情報」を横スクロールなしで確認できる
- [ ] AC-5: セルの縦 padding が全ページで統一されている（`py-2.5`）
- [ ] AC-6: 同種データ（日付・ステータス・操作）のカラム幅が統一されている

---

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| タブレット対応方式 | Tailwind `hidden lg:table-cell` で任意カラムを Desktop のみ表示 | DataTable を変更せず各ページの column 定義に className を付与するだけで実装可能 | `min-w` を動的化、JS での幅計算 |
| 必須カラムの定義 | 飼主名・ペット名・日付・ステータス・操作 を常時表示。それ以外は `hidden lg:table-cell` | 診察現場でまず「誰の何の状態か」を確認するのが主目的 | |
| padding 統一 | `STYLE.tableCell` トークンを使用（`py-2.5`）に統一 | 既存 design token に準拠 | `py-2` に統一（トークンを変更） |

---

## 各ページのタブレット対応方針

### 常時表示（全幅）
すべてのページで以下は必須表示：
- 日付系（診療日・実施日・入院日）
- 飼主名
- ペット名
- ステータス
- 操作

### タブレット(〜1023px)で非表示 `hidden lg:table-cell`

| ページ | 非表示カラム |
|--------|------------|
| 飼主一覧 | ペット番号、生死、生年月日、体重、環境、前回来院 |
| カルテ管理 | 種、関連 |
| 検査管理 | 結果概要 |
| トリミング管理 | 種、体重、スタイル希望 |
| 入院管理 | 種、退院予定日 |
| 定期健診 | 次回予定、結果・所見 |
| 会計管理 | カルテリンク列 |
| 在庫管理 | 最低在庫、有効期限 |

### カラム幅統一規約

```
DATE:   w-[120px]  — 日付（YYYY-MM-DD）
STATUS: w-[100px]  — ステータスバッジ
ACTION: w-[80px]   — 操作ボタン
DOCTOR: w-[100px]  — 担当医名
NARROW: w-[80px]   — 短い固定値（種・体重など）
```

---

## 影響範囲

### Frontend

| ファイル | 変更内容 |
|---------|---------|
| `components/shared/DataTable/DataTable.tsx` | `min-w-[800px]` を `min-w-[640px]` に緩和（FE-125） |
| `features/examinations/routes/Examinations.tsx` | `text-sm` → `text-base` + `py-2` → `py-2.5`（FE-124） |
| `features/medical-records/routes/MedicalRecords.tsx` | 主訴 `text-sm` → `text-base`（FE-124） |
| `features/accounting/routes/Accounting.tsx` | カラム幅統一 + tablet hidden（FE-125） |
| `features/owners/routes/OwnersList.tsx` | tablet hidden 6カラム（FE-125） |
| `features/trimming/routes/TrimmingList.tsx` | tablet hidden 3カラム（FE-125） |
| `features/medical-records/routes/MedicalRecords.tsx` | tablet hidden 2カラム（FE-125） |
| `features/hospitalization/components/HospitalizationListView.tsx` | tablet hidden 2カラム（FE-125） |
| `features/checkups/routes/CheckupsList.tsx` | tablet hidden 2カラム（FE-125） |
| `features/inventory/routes/InventoryList.tsx` | tablet hidden 2カラム（FE-125） |

### Backend
- 変更なし

---

## 参照実装

- `features/trimming/routes/TrimmingList.tsx` — `text-base` + `py-2` の現行スタンダード（padding のみ要修正）
- `features/accounting/routes/Accounting.tsx` — `STYLE.tableCell` 未使用の典型

---

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| `hidden lg:table-cell` で非表示になる列を検索・ソートしている場合は機能する（DOM に存在する） | 低 | 非表示＝CSS の `display:none`、フィルタロジックは影響なし |
| `py-2` → `py-2.5` でページのスクロール量が微増 | 低 | 許容範囲（2px 相当）|
| OwnersList で 6 カラム非表示 → 表示情報の大幅減少 | 中 | 飼主名・ペット名・種・操作 の最低 4 カラムを維持。クリックで詳細遷移可能 |

---

## 実装順序

1. FE-124: テキスト・padding 修正（2ファイル）
2. FE-125: DataTable min-w 緩和 + 各ページ tablet hidden + カラム幅統一

## 関連イシュー

- [FE-124: テキストサイズ・パディング統一](../../frontend/issues/open/FE-124-table-text-size-padding-unification.md)
- [FE-125: DataTable タブレット対応と列幅統一](../../frontend/issues/open/FE-125-datatable-tablet-first-responsive.md)
